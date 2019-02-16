package commit

import (
	"gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
)

type batchFindCommitSender struct {
	stream  gitalypb.CommitService_BatchFindCommitServer
	commits []*gitalypb.GitCommit
}

func (sender *batchFindCommitSender) Reset() { sender.commits = nil }
func (sender *batchFindCommitSender) Append(it chunk.Item) {
	sender.commits = append(sender.commits, it.(*gitalypb.GitCommit))
}
func (sender *batchFindCommitSender) Send() error {
	return sender.stream.Send(&gitalypb.BatchFindCommitResponse{Commits: sender.commits})
}

func (s *server) BatchFindCommit(in *gitalypb.BatchFindCommitRequest, stream gitalypb.CommitService_BatchFindCommitServer) error {
	repo := in.GetRepository()
	ctx := stream.Context()
	sender := &batchFindCommitSender{stream: stream}
	chunker := chunk.New(sender)

	for _, revision := range in.Revisions {
		if err := git.ValidateRevision(revision); err != nil {
			if err := chunker.Send(gitalypb.GitCommit{}); err != nil {
				return err
			}

			continue
		}

		commit, err := log.GetCommit(ctx, repo, string(revision))
		if log.IsNotFound(err) {
			if err := chunker.Send(gitalypb.GitCommit{}); err != nil {
				return err
			}
		} else {
			if err := chunker.Send(commit); err != nil {
				return err
			}
		}
	}

	if err := chunker.Flush(); err != nil {
		return err
	}

	return nil
}
