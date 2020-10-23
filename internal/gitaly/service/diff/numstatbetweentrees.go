package diff

import (
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) DiffStatsBetweenTrees(in *gitalypb.DiffStatsBetweenTreesRequest, stream gitalypb.DiffService_DiffStatsBetweenTreesServer) error {
	if err := s.validateDiffStatsBetweenTreesRequestParams(in); err != nil {
		return err
	}

	var batch []*gitalypb.DiffStats
	for i := 0 ; i < len(in.Trees)  - 1 ; i ++ {
		cmd, err := git.SafeCmd(stream.Context(), in.Repository, nil, git.SubCmd{
			Name:  "diff",
			Flags: []git.Option{git.Flag{Name: "--numstat"}, git.Flag{Name: "-z"}},
			Args:  []string{in.Trees[i + 1], in.Trees[i]},
		})

		if err != nil {
			if _, ok := status.FromError(err); ok {
				return err
			}
			return status.Errorf(codes.Internal, "%s: cmd: %v", "DiffStats", err)
		}

		diff.ParseNumStats(batch, cmd)

		if err := cmd.Wait(); err != nil {
			return status.Errorf(codes.Unavailable, "%s: %v", "DiffStats", err)
		}
	}

	return sendStats(batch, stream)
}

func (s *server) validateDiffStatsBetweenTreesRequestParams(in *gitalypb.DiffStatsBetweenTreesRequest) error {
	repo := in.GetRepository()
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	// if err := validateRequest(in); err != nil {
	// 	return status.Errorf(codes.InvalidArgument, "DiffStats: %v", err)
	// }

	return nil
}
