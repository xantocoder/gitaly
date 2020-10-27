package diff

import (
	"github.com/golang/protobuf/proto"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) DiffStats(in *gitalypb.DiffStatsRequest, stream gitalypb.DiffService_DiffStatsServer) error {
	if err := s.validateDiffStatsRequestParams(in); err != nil {
		return err
	}

	var batch []*gitalypb.DiffStats

	diffChunker := chunk.New(&diffStatSender{stream: stream})

	cmd, err := git.SafeCmd(stream.Context(), in.Repository, nil, git.SubCmd{
		Name:  "diff",
		Flags: []git.Option{git.Flag{Name: "--numstat"}, git.Flag{Name: "-z"}},
		Args:  []string{in.LeftCommitId, in.RightCommitId},
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

type diffStatSender struct {
	diffStats []*gitalypb.DiffStats
	stream    gitalypb.DiffService_DiffStatsServer
}

func (t *diffStatSender) Reset() {
	t.diffStats = nil
}

func (t *diffStatSender) Append(m proto.Message) {
	t.diffStats = append(t.diffStats, m.(*gitalypb.DiffStats))
}

func (t *diffStatSender) Send() error {
	return t.stream.Send(&gitalypb.DiffStatsResponse{
		Stats: t.diffStats,
	})
}

func (s *server) validateDiffStatsRequestParams(in *gitalypb.DiffStatsRequest) error {
	repo := in.GetRepository()
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	if err := validateRequest(in); err != nil {
		return status.Errorf(codes.InvalidArgument, "DiffStats: %v", err)
	}

	return nil
}
