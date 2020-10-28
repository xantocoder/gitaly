package diff

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) DiffTreeDiffStats(in *gitalypb.DiffTreeDiffStatsRequest, stream gitalypb.DiffService_DiffTreeDiffStatsServer) error {
	if err := s.validateDiffTreeDiffStatsRequestParams(in); err != nil {
		return err
	}

	var batch []*gitalypb.DiffStats

	diffChunker := chunk.New(&diffTreeDiffStatsSender{stream: stream})

	cmd, err := git.SafeCmd(stream.Context(), in.Repository, nil, git.SubCmd{
		Name:  "diff-tree",
		Flags: []git.Option{git.Flag{Name: "--numstat"}, git.Flag{Name: "-z"}, git.Flag{Name: "--stdin"}, git.Flag{Name: "-c"}, git.Flag{Name: "-m"}, git.Flag{Name: "--no-commit-id"}},
		Args:  in.GetCommits(),
	})

	if err != nil {
		if _, ok := status.FromError(err); ok {
			return err
		}
		return status.Errorf(codes.Internal, "%s: cmd: %v", "DiffStats", err)
	}

	if err := diff.ParseNumStats(batch, cmd, diffChunker); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return status.Errorf(codes.Unavailable, "%s: %v", "DiffStats", err)
	}

	return diffChunker.Flush()
}

type diffTreeDiffStatsSender struct {
	diffStats []*gitalypb.DiffStats
	stream    gitalypb.DiffService_DiffTreeDiffStatsServer
}

func (t *diffTreeDiffStatsSender) Reset() {
	t.diffStats = nil
}

func (t *diffTreeDiffStatsSender) Append(m proto.Message) {
	t.diffStats = append(t.diffStats, m.(*gitalypb.DiffStats))
}

func (t *diffTreeDiffStatsSender) Send() error {
	return t.stream.Send(&gitalypb.DiffTreeDiffStatsResponse{
		Stats: t.diffStats,
	})
}

func (s *server) validateDiffTreeDiffStatsRequestParams(in *gitalypb.DiffTreeDiffStatsRequest) error {
	repo := in.GetRepository()
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	if err := validateDiffStatCommitRangeRequest(in); err != nil {
		return status.Errorf(codes.InvalidArgument, "DiffStats: %v", err)
	}

	return nil
}

func validateDiffStatCommitRangeRequest(in *gitalypb.DiffTreeDiffStatsRequest) error {
	if len(in.GetCommits()) <= 1 {
		return fmt.Errorf("must have more than 1 commit")
	}

	for _, commit := range in.GetCommits() {
		if commit == "" {
			return fmt.Errorf("commits cannot contain an empty commit")
		}
	}

	return nil
}
