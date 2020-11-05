package objectpool

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/repository"
	"gitlab.com/gitlab-org/gitaly/internal/git/updateref"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

const (
	sourceRemote       = "origin"
	sourceRefNamespace = "refs/remotes/" + sourceRemote
)

// FetchFromOrigin initializes the pool and fetches the objects from its origin repository
func (o *ObjectPool) FetchFromOrigin(ctx context.Context, origin *gitalypb.Repository) error {
	if err := o.Init(ctx); err != nil {
		return err
	}

	originPath, err := o.locator.GetPath(origin)

	if err != nil {
		return err
	}

	getRemotes, err := git.SafeCmd(ctx, o, nil, git.SubCmd{Name: "remote"})
	if err != nil {
		return err
	}

	remoteReader := bufio.NewScanner(getRemotes)

	var originExists bool
	for remoteReader.Scan() {
		if remoteReader.Text() == sourceRemote {
			originExists = true
		}
	}
	if err := getRemotes.Wait(); err != nil {
		return err
	}

	var setOriginCmd *command.Command
	if originExists {
		setOriginCmd, err = git.SafeCmd(ctx, o, nil, git.SubCmd{
			Name: "remote",
			Args: []string{"set-url", sourceRemote, originPath},
		}, git.WithRefTxHook(ctx, helper.ProtoRepoFromRepo(o), config.Config))
		if err != nil {
			return err
		}
	} else {
		setOriginCmd, err = git.SafeCmd(ctx, o, nil, git.SubCmd{
			Name: "remote",
			Args: []string{"add", sourceRemote, originPath},
		}, git.WithRefTxHook(ctx, helper.ProtoRepoFromRepo(o), config.Config))
		if err != nil {
			return err
		}
	}

	if err := setOriginCmd.Wait(); err != nil {
		return err
	}

	if err := o.logStats(ctx, "before fetch"); err != nil {
		return err
	}

	refSpec := fmt.Sprintf("+refs/*:%s/*", sourceRefNamespace)
	fetchCmd, err := git.SafeCmd(ctx, o, nil, git.SubCmd{
		Name:  "fetch",
		Flags: []git.Option{git.Flag{Name: "--quiet"}},
		Args:  []string{sourceRemote, refSpec},
	})
	if err != nil {
		return err
	}

	if err := fetchCmd.Wait(); err != nil {
		return err
	}

	if err := rescueDanglingObjects(ctx, o); err != nil {
		return err
	}

	if err := o.logStats(ctx, "after fetch"); err != nil {
		return err
	}

	packRefs, err := git.SafeCmd(ctx, o, nil, git.SubCmd{
		Name:  "pack-refs",
		Flags: []git.Option{git.Flag{Name: "--all"}},
	})
	if err != nil {
		return err
	}
	if err := packRefs.Wait(); err != nil {
		return err
	}

	return repackPool(ctx, o)
}

const danglingObjectNamespace = "refs/dangling"

// rescueDanglingObjects creates refs for all dangling objects if finds
// with `git fsck`, which converts those objects from "dangling" to
// "not-dangling". This guards against any object ever being deleted from
// a pool repository. This is a defense in depth against accidental use
// of `git prune`, which could remove Git objects that a pool member
// relies on. There is currently no way for us to reliably determine if
// an object is still used anywhere, so the only safe thing to do is to
// assume that every object _is_ used.
func rescueDanglingObjects(ctx context.Context, repo repository.GitRepo) error {
	fsck, err := git.SafeCmd(ctx, repo, nil, git.SubCmd{
		Name:  "fsck",
		Flags: []git.Option{git.Flag{Name: "--connectivity-only"}, git.Flag{Name: "--dangling"}},
	})
	if err != nil {
		return err
	}

	updater, err := updateref.New(ctx, repo)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(fsck)
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), " ", 3)
		if len(split) != 3 {
			continue
		}

		if split[0] != "dangling" {
			continue
		}

		ref := danglingObjectNamespace + "/" + split[2]
		if err := updater.Create(ref, split[2]); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := fsck.Wait(); err != nil {
		return fmt.Errorf("git fsck: %v", err)
	}

	return updater.Wait()
}

func repackPool(ctx context.Context, pool repository.GitRepo) error {
	repackArgs := []git.Option{
		git.ValueFlag{"-c", "pack.island=" + sourceRefNamespace + "/he(a)ds"},
		git.ValueFlag{"-c", "pack.island=" + sourceRefNamespace + "/t(a)gs"},
		git.ValueFlag{"-c", "pack.islandCore=a"},
		git.ValueFlag{"-c", "pack.writeBitmapHashCache=true"},
	}

	repackCmd, err := git.SafeCmd(ctx, pool, repackArgs, git.SubCmd{
		Name:  "repack",
		Flags: []git.Option{git.Flag{Name: "-aidb"}},
	})
	if err != nil {
		return err
	}

	if err := repackCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (o *ObjectPool) logStats(ctx context.Context, when string) error {
	fields := logrus.Fields{
		"when": when,
	}

	for key, dir := range map[string]string{
		"poolObjectsSize": "objects",
		"poolRefsSize":    "refs",
	} {
		var err error
		fields[key], err = sizeDir(ctx, filepath.Join(o.FullPath(), dir))
		if err != nil {
			return err
		}
	}

	forEachRef, err := git.SafeCmd(ctx, o, nil, git.SubCmd{
		Name:  "for-each-ref",
		Flags: []git.Option{git.Flag{Name: "--format=%(objecttype)"}},
		Args:  []string{danglingObjectNamespace},
	})
	if err != nil {
		return err
	}

	counts := make(map[string]int)
	scanner := bufio.NewScanner(forEachRef)
	for scanner.Scan() {
		counts[scanner.Text()]++
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	if err := forEachRef.Wait(); err != nil {
		return err
	}

	for _, key := range []string{"blob", "commit", "tag", "tree"} {
		fields["dangling."+key+".ref"] = counts[key]
	}

	ctxlogrus.Extract(ctx).WithFields(fields).Info("pool dangling ref stats")

	return nil
}

func sizeDir(ctx context.Context, dir string) (int64, error) {
	// du -k reports size in KB
	cmd, err := command.New(ctx, exec.Command("du", "-sk", dir), nil, nil, nil)
	if err != nil {
		return 0, err
	}

	sizeLine, err := ioutil.ReadAll(cmd)
	if err != nil {
		return 0, err
	}

	if err := cmd.Wait(); err != nil {
		return 0, err
	}

	sizeParts := bytes.Split(sizeLine, []byte("\t"))
	if len(sizeParts) != 2 {
		return 0, fmt.Errorf("malformed du output: %q", sizeLine)
	}

	size, err := strconv.ParseInt(string(sizeParts[0]), 10, 0)
	if err != nil {
		return 0, err
	}

	// Convert KB to B
	return size * 1024, nil
}
