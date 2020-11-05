package operations

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/alternates"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	userSquashImplCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitaly_user_squash_counter",
			Help: "Number of calls to UserSquash rpc for each implementation (ruby/go)",
		},
		[]string{"impl"},
	)
)

func init() {
	prometheus.MustRegister(userSquashImplCounter)
}

const (
	squashWorktreePrefix  = "squash"
	gitlabWorktreesSubDir = "gitlab-worktree"
)

func (s *server) UserSquash(ctx context.Context, req *gitalypb.UserSquashRequest) (*gitalypb.UserSquashResponse, error) {
	if err := validateUserSquashRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "UserSquash: %v", err)
	}

	if featureflag.IsEnabled(ctx, featureflag.GoUserSquash) {
		userSquashImplCounter.WithLabelValues("go").Inc()
		return s.userSquashGo(ctx, req)
	}

	userSquashImplCounter.WithLabelValues("ruby").Inc()

	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, req.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserSquash(clientCtx, req)
}

func validateUserSquashRequest(req *gitalypb.UserSquashRequest) error {
	if req.GetRepository() == nil {
		return fmt.Errorf("empty Repository")
	}

	if req.GetUser() == nil {
		return fmt.Errorf("empty User")
	}

	if req.GetSquashId() == "" {
		return fmt.Errorf("empty SquashId")
	}

	if req.GetStartSha() == "" {
		return fmt.Errorf("empty StartSha")
	}

	if req.GetEndSha() == "" {
		return fmt.Errorf("empty EndSha")
	}

	if len(req.GetCommitMessage()) == 0 {
		return fmt.Errorf("empty CommitMessage")
	}

	if req.GetAuthor() == nil {
		return fmt.Errorf("empty Author")
	}

	return nil
}

type gitError struct {
	// ErrMsg error message from 'git' executable if any.
	ErrMsg string
	// Err is an error that happened during rebase process.
	Err error
}

func (er gitError) Error() string {
	return er.ErrMsg + ": " + er.Err.Error()
}

func (s *server) userSquashGo(ctx context.Context, req *gitalypb.UserSquashRequest) (*gitalypb.UserSquashResponse, error) {
	if strings.Contains(req.GetSquashId(), "/") {
		return nil, helper.ErrInvalidArgument(errors.New("worktree id can't contain slashes"))
	}

	repoPath, env, err := alternates.PathAndEnv(req.GetRepository())
	if err != nil {
		return nil, helper.ErrInternal(fmt.Errorf("alternate path: %w", err))
	}

	sha, err := s.runUserSquashGo(ctx, req, env, repoPath)
	if err != nil {
		var gitErr gitError
		if errors.As(err, &gitErr) {
			if gitErr.ErrMsg != "" {
				// we log an actual error as it would be lost otherwise (it is not sent back to the client)
				ctxlogrus.Extract(ctx).WithError(err).Error("user squash")
				return &gitalypb.UserSquashResponse{GitError: gitErr.ErrMsg}, nil
			}
		}

		return nil, helper.ErrInternal(err)
	}

	return &gitalypb.UserSquashResponse{SquashSha: sha}, nil
}

func (s *server) runUserSquashGo(ctx context.Context, req *gitalypb.UserSquashRequest, env []string, repoPath string) (string, error) {
	sparseDiffFiles, err := s.diffFiles(ctx, env, repoPath, req)
	if err != nil {
		return "", fmt.Errorf("define diff files: %w", err)
	}

	if len(sparseDiffFiles) == 0 {
		sha, err := s.userSquashWithNoDiff(ctx, req, repoPath, env)
		if err != nil {
			return "", fmt.Errorf("without sparse diff: %w", err)
		}

		return sha, nil
	}

	sha, err := s.userSquashWithDiffInFiles(ctx, req, repoPath, env, sparseDiffFiles)
	if err != nil {
		return "", fmt.Errorf("with sparse diff: %w", err)
	}

	return sha, nil
}

func (s *server) diffFiles(ctx context.Context, env []string, repoPath string, req *gitalypb.UserSquashRequest) ([]byte, error) {
	var stdout, stderr bytes.Buffer
	cmd, err := git.SafeBareCmd(ctx, git.CmdStream{Out: &stdout, Err: &stderr}, env,
		[]git.Option{git.ValueFlag{Name: "--git-dir", Value: repoPath}},
		git.SubCmd{
			Name:  "diff",
			Flags: []git.Option{git.Flag{Name: "--name-only"}, git.Flag{Name: "--diff-filter=ar"}, git.Flag{Name: "--binary"}},
			Args:  []string{diffRange(req)},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create 'git diff': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("on 'git diff' awaiting: %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	return stdout.Bytes(), nil
}

var errNoFilesCheckedOut = errors.New("no files checked out")

func (s *server) userSquashWithDiffInFiles(ctx context.Context, req *gitalypb.UserSquashRequest, repoPath string, env []string, diffFilesOut []byte) (string, error) {
	repo := req.GetRepository()
	worktreePath := newSquashWorktreePath(repoPath, req.GetSquashId())

	if err := s.addWorktree(ctx, repo, worktreePath, ""); err != nil {
		return "", fmt.Errorf("add worktree: %w", err)
	}

	defer func(worktreeName string) {
		if err := s.removeWorktree(ctx, repo, worktreeName); err != nil {
			ctxlogrus.Extract(ctx).WithField("worktree_name", worktreeName).WithError(err).Error("failed to remove worktree")
		}
	}(filepath.Base(worktreePath))

	worktreeGitPath, err := s.revParseGitDir(ctx, worktreePath)
	if err != nil {
		return "", fmt.Errorf("define git dir for worktree: %w", err)
	}

	if err := runCmd(ctx, repo, "config", []git.Option{git.ConfigPair{Key: "core.sparseCheckout", Value: "true"}}, nil); err != nil {
		return "", fmt.Errorf("on 'git config core.sparseCheckout true': %w", err)
	}

	if err := s.createSparseCheckoutFile(worktreeGitPath, diffFilesOut); err != nil {
		return "", fmt.Errorf("create sparse checkout file: %w", err)
	}

	if err := s.checkout(ctx, worktreePath, req); err != nil {
		if !errors.Is(err, errNoFilesCheckedOut) {
			return "", fmt.Errorf("perform 'git checkout' with core.sparseCheckout true: %w", err)
		}

		// try to perform checkout with disabled sparseCheckout feature
		if err := runCmd(ctx, repo, "config", []git.Option{git.ConfigPair{Key: "core.sparseCheckout", Value: "false"}}, nil); err != nil {
			return "", fmt.Errorf("on 'git config core.sparseCheckout false': %w", err)
		}

		if err := s.checkout(ctx, worktreePath, req); err != nil {
			return "", fmt.Errorf("perform 'git checkout' with core.sparseCheckout false: %w", err)
		}
	}

	sha, err := s.applyDiff(ctx, req, worktreePath, env)
	if err != nil {
		return "", fmt.Errorf("apply diff: %w", err)
	}

	return sha, nil
}

func (s *server) checkout(ctx context.Context, worktreePath string, req *gitalypb.UserSquashRequest) error {
	var stderr bytes.Buffer
	checkoutCmd, err := git.SafeBareCmdInDir(ctx, worktreePath, git.CmdStream{Err: &stderr}, nil, nil,
		git.SubCmd{
			Name:  "checkout",
			Flags: []git.Option{git.Flag{Name: "--detach"}},
			Args:  []string{req.GetStartSha()},
		},
	)
	if err != nil {
		return fmt.Errorf("create 'git checkout': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	if err = checkoutCmd.Wait(); err != nil {
		if strings.Contains(stderr.String(), "error: Sparse checkout leaves no entry on working directory") {
			return errNoFilesCheckedOut
		}

		return fmt.Errorf("wait for 'git checkout': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	return nil
}

func (s *server) revParseGitDir(ctx context.Context, worktreePath string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd, err := git.SafeBareCmdInDir(ctx, worktreePath, git.CmdStream{Out: &stdout, Err: &stderr}, nil, nil, git.SubCmd{
		Name:  "rev-parse",
		Flags: []git.Option{git.Flag{Name: "--git-dir"}},
	})
	if err != nil {
		return "", fmt.Errorf("creation of 'git rev-parse --git-dir': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("wait for 'git rev-parse --git-dir': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	return text.ChompBytes(stdout.Bytes()), nil
}

func (s *server) userSquashWithNoDiff(ctx context.Context, req *gitalypb.UserSquashRequest, repoPath string, env []string) (string, error) {
	repo := req.GetRepository()
	worktreePath := newSquashWorktreePath(repoPath, req.GetSquashId())

	if err := s.addWorktree(ctx, repo, worktreePath, req.GetStartSha()); err != nil {
		return "", fmt.Errorf("add worktree: %w", err)
	}

	defer func(worktreeName string) {
		if err := s.removeWorktree(ctx, repo, worktreeName); err != nil {
			ctxlogrus.Extract(ctx).WithField("worktree_name", worktreeName).WithError(err).Error("failed to remove worktree")
		}
	}(filepath.Base(worktreePath))

	sha, err := s.applyDiff(ctx, req, worktreePath, env)
	if err != nil {
		return "", fmt.Errorf("apply diff: %w", err)
	}

	return sha, nil
}

func (s *server) addWorktree(ctx context.Context, repo *gitalypb.Repository, worktreePath string, committish string) error {
	if err := runCmd(ctx, repo, "config", []git.Option{git.ConfigPair{Key: "core.splitIndex", Value: "false"}}, nil); err != nil {
		return fmt.Errorf("on 'git config core.splitIndex false': %w", err)
	}

	args := []string{worktreePath}
	flags := []git.Option{git.SubSubCmd{Name: "add"}, git.Flag{Name: "--detach"}}
	if committish != "" {
		args = append(args, committish)
	} else {
		flags = append(flags, git.Flag{Name: "--no-checkout"})
	}

	var stderr bytes.Buffer
	cmd, err := git.SafeCmd(ctx, repo, nil,
		git.SubCmd{Name: "worktree", Flags: flags, Args: args},
		git.WithStderr(&stderr),
		git.WithRefTxHook(ctx, repo, s.cfg),
	)
	if err != nil {
		return fmt.Errorf("creation of 'git worktree add': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for 'git worktree add': %w", gitError{ErrMsg: stderr.String(), Err: err})
	}

	return nil
}

func (s *server) removeWorktree(ctx context.Context, repo *gitalypb.Repository, worktreeName string) error {
	cmd, err := git.SafeCmd(ctx, repo, nil,
		git.SubCmd{
			Name:  "worktree",
			Flags: []git.Option{git.SubSubCmd{Name: "remove"}, git.Flag{Name: "--force"}},
			Args:  []string{worktreeName},
		},
		git.WithRefTxHook(ctx, repo, s.cfg),
	)
	if err != nil {
		return fmt.Errorf("creation of 'worktree remove': %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for 'worktree remove': %w", err)
	}

	return nil
}

func (s *server) applyDiff(ctx context.Context, req *gitalypb.UserSquashRequest, worktreePath string, env []string) (string, error) {
	diffRange := diffRange(req)

	var diffStderr bytes.Buffer
	cmdDiff, err := git.SafeCmd(ctx, req.GetRepository(), nil,
		git.SubCmd{
			Name: "diff",
			Flags: []git.Option{
				git.Flag{Name: "--binary"},
			},
			Args: []string{diffRange},
		},
		git.WithStderr(&diffStderr),
	)
	if err != nil {
		return "", fmt.Errorf("creation of 'git diff' for range %q: %w", diffRange, gitError{ErrMsg: diffStderr.String(), Err: err})
	}

	var applyStderr bytes.Buffer
	cmdApply, err := git.SafeBareCmdInDir(ctx, worktreePath, git.CmdStream{In: command.SetupStdin, Err: &applyStderr}, env, nil, git.SubCmd{
		Name: "apply",
		Flags: []git.Option{
			git.Flag{Name: "--index"},
			git.Flag{Name: "--3way"},
			git.Flag{Name: "--whitespace=nowarn"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("creation of 'git apply' for range %q: %w", diffRange, gitError{ErrMsg: applyStderr.String(), Err: err})
	}

	if _, err := io.Copy(cmdApply, cmdDiff); err != nil {
		return "", fmt.Errorf("piping 'git diff' -> 'git apply' for range %q: %w", diffRange, gitError{ErrMsg: applyStderr.String(), Err: err})
	}

	if err := cmdDiff.Wait(); err != nil {
		return "", fmt.Errorf("wait for 'git diff' for range %q: %w", diffRange, gitError{ErrMsg: diffStderr.String(), Err: err})
	}

	if err := cmdApply.Wait(); err != nil {
		return "", fmt.Errorf("wait for 'git apply' for range %q: %w", diffRange, gitError{ErrMsg: applyStderr.String(), Err: err})
	}

	commitEnv := append(env,
		"GIT_COMMITTER_NAME="+string(req.GetUser().Name),
		"GIT_COMMITTER_EMAIL="+string(req.GetUser().Email),
		"GIT_AUTHOR_NAME="+string(req.GetAuthor().Name),
		"GIT_AUTHOR_EMAIL="+string(req.GetAuthor().Email),
	)

	var commitStderr bytes.Buffer
	cmdCommit, err := git.SafeBareCmdInDir(ctx, worktreePath, git.CmdStream{Err: &commitStderr}, commitEnv, nil, git.SubCmd{
		Name: "commit",
		Flags: []git.Option{
			git.Flag{Name: "--no-verify"},
			git.Flag{Name: "--quiet"},
			git.ValueFlag{Name: "--message", Value: string(req.GetCommitMessage())},
		},
	})
	if err != nil {
		return "", fmt.Errorf("creation of 'git commit': %w", gitError{ErrMsg: commitStderr.String(), Err: err})
	}

	if err := cmdCommit.Wait(); err != nil {
		return "", fmt.Errorf("wait for 'git commit': %w", gitError{ErrMsg: commitStderr.String(), Err: err})
	}

	var revParseStdout, revParseStderr bytes.Buffer
	revParseCmd, err := git.SafeBareCmdInDir(ctx, worktreePath, git.CmdStream{Out: &revParseStdout, Err: &revParseStderr}, env, nil, git.SubCmd{
		Name: "rev-parse",
		Flags: []git.Option{
			git.Flag{Name: "--quiet"},
			git.Flag{Name: "--verify"},
		},
		Args: []string{"HEAD^{commit}"},
	})
	if err != nil {
		return "", fmt.Errorf("creation of 'git rev-parse': %w", gitError{ErrMsg: revParseStderr.String(), Err: err})
	}

	if err := revParseCmd.Wait(); err != nil {
		return "", fmt.Errorf("wait for 'git rev-parse': %w", gitError{ErrMsg: revParseStderr.String(), Err: err})
	}

	return text.ChompBytes(revParseStdout.Bytes()), nil
}

func (s *server) createSparseCheckoutFile(worktreeGitPath string, diffFilesOut []byte) error {
	if err := os.MkdirAll(filepath.Join(worktreeGitPath, "info"), 0755); err != nil {
		return fmt.Errorf("create 'info' dir for worktree %q: %w", worktreeGitPath, err)
	}

	if err := ioutil.WriteFile(filepath.Join(worktreeGitPath, "info", "sparse-checkout"), diffFilesOut, 0666); err != nil {
		return fmt.Errorf("create 'sparse-checkout' file for worktree %q: %w", worktreeGitPath, err)
	}

	return nil
}

func diffRange(req *gitalypb.UserSquashRequest) string {
	return req.GetStartSha() + "..." + req.GetEndSha()
}

func newSquashWorktreePath(repoPath, squashID string) string {
	prefix := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Shuffle(len(prefix), func(i, j int) { prefix[i], prefix[j] = prefix[j], prefix[i] })

	worktreeName := squashWorktreePrefix + "-" + squashID + "-" + string(prefix[:32])
	return filepath.Join(repoPath, gitlabWorktreesSubDir, worktreeName)
}

func runCmd(ctx context.Context, repo *gitalypb.Repository, cmd string, opts []git.Option, args []string) error {
	var stderr bytes.Buffer
	safeCmd, err := git.SafeCmd(ctx, repo, nil, git.SubCmd{Name: cmd, Flags: opts, Args: args}, git.WithStderr(&stderr))
	if err != nil {
		return fmt.Errorf("create safe cmd %q: %w", cmd, gitError{ErrMsg: stderr.String(), Err: err})
	}

	if err := safeCmd.Wait(); err != nil {
		return fmt.Errorf("wait safe cmd %q: %w", cmd, gitError{ErrMsg: stderr.String(), Err: err})
	}

	return nil
}
