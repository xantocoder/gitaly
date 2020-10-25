package diff

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) DiffStatsCommitRange(in *gitalypb.DiffStatsCommitRangeRequest, stream gitalypb.DiffService_DiffStatsCommitRangeServer) error {
	if err := s.validateDiffStatsCommitRangeRequestParams(in); err != nil {
		return err
	}

	var batch []*gitalypb.DiffStats

	diffChunker := chunk.New(&diffStatCommitRangeSender{stream: stream})

	for _, commit := range in.GetCommits() {
		cmd, err := git.SafeCmd(stream.Context(), in.Repository, nil, git.SubCmd{
			Name:  "diff",
			Flags: []git.Option{git.Flag{Name: "--numstat"}, git.Flag{Name: "-z"}},
			Args:  []string{commit + "~", commit},
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
	}

	return diffChunker.Flush()
}

type diffStatCommitRangeSender struct {
	diffStats   []*gitalypb.DiffStats
	stream gitalypb.DiffService_DiffStatsCommitRangeServer
}

func (t *diffStatCommitRangeSender) Reset() {
	t.diffStats = nil
}

func (t *diffStatCommitRangeSender) Append(m proto.Message) {
	t.diffStats = append(t.diffStats, m.(*gitalypb.DiffStats))
}

func (t *diffStatCommitRangeSender) Send() error {
	return t.stream.Send(&gitalypb.DiffStatsCommitRangeResponse{
		Stats: t.diffStats,
	})
}

func (s *server) validateDiffStatsCommitRangeRequestParams(in *gitalypb.DiffStatsCommitRangeRequest) error {
	repo := in.GetRepository()
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	if err := validateDiffStatCommitRangeRequest(in); err != nil {
		return status.Errorf(codes.InvalidArgument, "DiffStats: %v", err)
	}

	return nil
}

func validateDiffStatCommitRangeRequest(in *gitalypb.DiffStatsCommitRangeRequest) error {
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
