package commit

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type commitsBetweenSender struct {
	stream  gitalypb.CommitService_CommitsBetweenServer
	commits []*gitalypb.GitCommit
}

func (sender *commitsBetweenSender) Reset() { sender.commits = nil }
func (sender *commitsBetweenSender) Append(m proto.Message) {
	sender.commits = append(sender.commits, m.(*gitalypb.GitCommit))
}

func (sender *commitsBetweenSender) Send() error {
	return sender.stream.Send(&gitalypb.CommitsBetweenResponse{Commits: sender.commits})
}

func (s *server) CommitsBetween(in *gitalypb.CommitsBetweenRequest, stream gitalypb.CommitService_CommitsBetweenServer) error {
	if err := validateCommitsBetween(in); err != nil {
		return helper.ErrInvalidArgument(err)
	}

	sender := &commitsBetweenSender{stream: stream}
	from := in.GetFrom()
	var limit int32 = 2147483647

	if in.PaginationParams != nil {
		from = []byte(in.PaginationParams.GetPageToken())
		limit = in.PaginationParams.GetLimit()
	}

	revisionRange := fmt.Sprintf("%s..%s", from, in.GetTo())

	if err := sendCommits(stream.Context(), sender, in.GetRepository(),
		[]string{revisionRange}, nil, nil,
		git.Flag{Name: "--reverse"},
		git.ValueFlag{Name: "--max-count", Value: string(limit)}); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func validateCommitsBetween(in *gitalypb.CommitsBetweenRequest) error {
	if err := git.ValidateRevision(in.GetFrom()); err != nil {
		return fmt.Errorf("from: %v", err)
	}

	if err := git.ValidateRevision(in.GetTo()); err != nil {
		return fmt.Errorf("to: %v", err)
	}

	return nil
}
