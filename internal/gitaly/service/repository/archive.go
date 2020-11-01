package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/commit"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/storage"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/streamio"
	"gitlab.com/gitlab-org/labkit/correlation"
)

type archiveParams struct {
	ctx         context.Context
	writer      io.Writer
	in          *gitalypb.GetArchiveRequest
	compressCmd *exec.Cmd
	format      string
	archivePath string
	exclude     []string
	internalCfg []byte
	binDir      string
	loggingDir  string
}

func (s *server) GetArchive(in *gitalypb.GetArchiveRequest, stream gitalypb.RepositoryService_GetArchiveServer) error {
	ctx := stream.Context()
	compressCmd, format := parseArchiveFormat(in.GetFormat())

	repoRoot, err := s.locator.GetRepoPath(in.GetRepository())
	if err != nil {
		return err
	}

	path, err := storage.ValidateRelativePath(repoRoot, string(in.GetPath()))
	if err != nil {
		return helper.ErrInvalidArgument(err)
	}

	exclude := make([]string, len(in.GetExclude()))
	for i, ex := range in.GetExclude() {
		exclude[i], err = storage.ValidateRelativePath(repoRoot, string(ex))
		if err != nil {
			return helper.ErrInvalidArgument(err)
		}
	}

	if err := validateGetArchiveRequest(in, format, path); err != nil {
		return err
	}

	if err := validateGetArchivePrecondition(ctx, in, path, exclude); err != nil {
		return err
	}

	if in.GetElidePath() {
		// `git archive <commit ID>:<path>` expects exclusions to be relative to path
		pathSlash := path + string(os.PathSeparator)
		for i := range exclude {
			if !strings.HasPrefix(exclude[i], pathSlash) {
				return helper.ErrInvalidArgumentf("invalid exclude: %q is not a subdirectory of %q", exclude[i], path)
			}

			exclude[i] = exclude[i][len(pathSlash):]
		}
	}

	writer := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&gitalypb.GetArchiveResponse{Data: p})
	})

	gitlabConfig, err := json.Marshal(s.cfg)
	if err != nil {
		return err
	}

	return handleArchive(archiveParams{
		ctx:         ctx,
		writer:      writer,
		in:          in,
		compressCmd: compressCmd,
		format:      format,
		archivePath: path,
		exclude:     exclude,
		internalCfg: gitlabConfig,
		binDir:      s.binDir,
		loggingDir:  s.loggingCfg.Dir,
	})
}

func parseArchiveFormat(format gitalypb.GetArchiveRequest_Format) (*exec.Cmd, string) {
	switch format {
	case gitalypb.GetArchiveRequest_TAR:
		return nil, "tar"
	case gitalypb.GetArchiveRequest_TAR_GZ:
		return exec.Command("gzip", "-c", "-n"), "tar"
	case gitalypb.GetArchiveRequest_TAR_BZ2:
		return exec.Command("bzip2", "-c"), "tar"
	case gitalypb.GetArchiveRequest_ZIP:
		return nil, "zip"
	}

	return nil, ""
}

func validateGetArchiveRequest(in *gitalypb.GetArchiveRequest, format string, path string) error {
	if err := git.ValidateRevision([]byte(in.GetCommitId())); err != nil {
		return helper.ErrInvalidArgumentf("invalid commitId: %v", err)
	}

	if len(format) == 0 {
		return helper.ErrInvalidArgumentf("invalid format")
	}

	return nil
}

func validateGetArchivePrecondition(ctx context.Context, in *gitalypb.GetArchiveRequest, path string, exclude []string) error {
	c, err := catfile.New(ctx, in.GetRepository())
	if err != nil {
		return err
	}

	f := commit.NewTreeEntryFinder(c)
	if path != "." {
		if ok, err := findGetArchivePath(f, in.GetCommitId(), path); err != nil {
			return err
		} else if !ok {
			return helper.ErrPreconditionFailedf("path doesn't exist")
		}
	}

	for i, exclude := range exclude {
		if ok, err := findGetArchivePath(f, in.GetCommitId(), exclude); err != nil {
			return err
		} else if !ok {
			return helper.ErrPreconditionFailedf("exclude[%d] doesn't exist", i)
		}
	}

	return nil
}

func findGetArchivePath(f *commit.TreeEntryFinder, commitID, path string) (ok bool, err error) {
	treeEntry, err := f.FindByRevisionAndPath(commitID, path)
	if err != nil {
		return false, err
	}

	if treeEntry == nil || len(treeEntry.Oid) == 0 {
		return false, nil
	}
	return true, nil
}

func handleArchive(p archiveParams) error {
	var args []string
	pathspecs := make([]string, 0, len(p.exclude)+1)
	if !p.in.GetElidePath() {
		// git archive [options] <commit ID> -- <path> [exclude*]
		args = []string{p.in.GetCommitId()}
		pathspecs = append(pathspecs, p.archivePath)
	} else if p.archivePath != "." {
		// git archive [options] <commit ID>:<path> -- [exclude*]
		args = []string{p.in.GetCommitId() + ":" + p.archivePath}
	} else {
		// git archive [options] <commit ID> -- [exclude*]
		args = []string{p.in.GetCommitId()}
	}

	for _, exclude := range p.exclude {
		pathspecs = append(pathspecs, ":(exclude)"+exclude)
	}

	env := []string{
		fmt.Sprintf("GL_REPOSITORY=%s", p.in.GetRepository().GetGlRepository()),
		fmt.Sprintf("GL_PROJECT_PATH=%s", p.in.GetRepository().GetGlProjectPath()),
		fmt.Sprintf("GL_INTERNAL_CONFIG=%s", p.internalCfg),
		fmt.Sprintf("CORRELATION_ID=%s", correlation.ExtractFromContext(p.ctx)),
		fmt.Sprintf("GITALY_LOG_DIR=%s", p.loggingDir),
	}

	var globals []git.Option

	if p.in.GetIncludeLfsBlobs() {
		binary := filepath.Join(p.binDir, "gitaly-lfs-smudge")
		globals = append(globals, git.ValueFlag{"-c", fmt.Sprintf("filter.lfs.smudge=%s", binary)})
	}

	archiveCommand, err := git.SafeCmdWithEnv(p.ctx, env, p.in.GetRepository(), globals, git.SubCmd{
		Name:        "archive",
		Flags:       []git.Option{git.ValueFlag{"--format", p.format}, git.ValueFlag{"--prefix", p.in.GetPrefix() + "/"}},
		Args:        args,
		PostSepArgs: pathspecs,
	})
	if err != nil {
		return err
	}

	if p.compressCmd != nil {
		command, err := command.New(p.ctx, p.compressCmd, archiveCommand, p.writer, nil)
		if err != nil {
			return err
		}

		if err := command.Wait(); err != nil {
			return err
		}
	} else if _, err = io.Copy(p.writer, archiveCommand); err != nil {
		return err
	}

	return archiveCommand.Wait()
}
