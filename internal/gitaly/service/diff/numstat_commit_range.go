package diff

import (
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) DiffStatsCommitRange(in *gitalypb.DiffStatsCommitRangeRequest, stream gitalypb.DiffService_DiffStatsCommitRangeServer) error {
	if err := s.validateDiffStatsCommitRangeRequestParams(in); err != nil {
		return err
	}

	var batch []*gitalypb.DiffStats
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

		if err := diff.ParseNumStats(batch, cmd); err != nil {
			return err
		}

		if err := cmd.Wait(); err != nil {
			return status.Errorf(codes.Unavailable, "%s: %v", "DiffStats", err)
		}
	}

	return sendCommitRangeStats(batch, stream)
}

func sendCommitRangeStats(batch []*gitalypb.DiffStats, stream gitalypb.DiffService_DiffStatsCommitRangeServer) error {
	if len(batch) == 0 {
		return nil
	}

	if err := stream.Send(&gitalypb.DiffStatsCommitRangeResponse{Stats: batch}); err != nil {
		return status.Errorf(codes.Unavailable, "DiffStats: send: %v", err)
	}

	return nil
}

func (s *server) validateDiffStatsCommitRangeRequestParams(in *gitalypb.DiffStatsCommitRangeRequest) error {
	repo := in.GetRepository()
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	// if err := validateRequest(in); err != nil {
	// 	return status.Errorf(codes.InvalidArgument, "DiffStats: %v", err)
	// }

	return nil
}
