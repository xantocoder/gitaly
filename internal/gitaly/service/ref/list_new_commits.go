package ref

import (
	"bufio"

	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

func (s *server) ListNewCommits(in *gitalypb.ListNewCommitsRequest, stream gitalypb.RefService_ListNewCommitsServer) error {
	oid := in.GetCommitId()
	if err := git.ValidateCommitID(oid); err != nil {
		return helper.ErrInvalidArgument(err)
	}

	if err := listNewCommits(in, stream, oid); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func listNewCommits(in *gitalypb.ListNewCommitsRequest, stream gitalypb.RefService_ListNewCommitsServer, oid string) error {
	ctx := stream.Context()

	revList, err := git.SafeCmd(ctx, in.GetRepository(), nil, git.SubCmd{
		Name:  "rev-list",
		Flags: []git.Option{git.Flag{Name: "--not"}, git.Flag{Name: "--all"}},
		Args:  []string{"^" + oid}, // the added ^ is to negate the oid since there is a --not option that comes earlier in the arg list
	})
	if err != nil {
		return err
	}

	batch, err := catfile.New(ctx, in.GetRepository())
	if err != nil {
		return err
	}

	commits := []*gitalypb.GitCommit{}
	scanner := bufio.NewScanner(revList)
	for scanner.Scan() {
		line := scanner.Text()

		commit, err := log.GetCommitCatfile(ctx, batch, line)
		if err != nil {
			return err
		}
		commits = append(commits, commit)

		if len(commits) >= 10 {
			response := &gitalypb.ListNewCommitsResponse{Commits: commits}
			if err := stream.Send(response); err != nil {
				return err
			}

			commits = commits[:0]
		}
	}

	if len(commits) > 0 {
		response := &gitalypb.ListNewCommitsResponse{Commits: commits}
		if err := stream.Send(response); err != nil {
			return err
		}
	}

	return revList.Wait()
}
