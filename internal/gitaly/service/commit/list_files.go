package commit

import (
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/lstree"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
)

func (s *server) ListFiles(in *gitalypb.ListFilesRequest, stream gitalypb.CommitService_ListFilesServer) error {
	ctxlogrus.Extract(stream.Context()).WithFields(log.Fields{
		"Revision": in.GetRevision(),
	}).Debug("ListFiles")

	if err := validateListFilesRequest(in); err != nil {
		return err
	}

	repo := in.Repository
	if _, err := s.locator.GetRepoPath(repo); err != nil {
		return err
	}

	revision := string(in.GetRevision())
	if len(revision) == 0 {
		defaultBranch, err := defaultBranchName(stream.Context(), repo)
		if err != nil {
			return helper.DecorateError(codes.NotFound, fmt.Errorf("revision not found %q", revision))
		}

		revision = string(defaultBranch)
	}

	containsRef, err := git.NewRepository(in.Repository).ContainsRef(stream.Context(), revision)
	if err != nil {
		return helper.ErrInternal(err)
	}

	if !containsRef {
		return stream.Send(&gitalypb.ListFilesResponse{})
	}

	if err := listFiles(repo, revision, stream); err != nil {
		return helper.ErrInternal(err)
	}

	return nil
}

func validateListFilesRequest(in *gitalypb.ListFilesRequest) error {
	if err := git.ValidateRevisionAllowEmpty(in.Revision); err != nil {
		return helper.ErrInvalidArgument(err)
	}
	return nil
}

func listFiles(repo *gitalypb.Repository, revision string, stream gitalypb.CommitService_ListFilesServer) error {
	cmd, err := git.SafeCmd(stream.Context(), repo, nil, git.SubCmd{Name: "ls-tree",
		Flags:       []git.Option{git.Flag{Name: "-z"}, git.Flag{Name: "-r"}, git.Flag{Name: "--full-tree"}, git.Flag{Name: "--full-name"}},
		PostSepArgs: []string{revision},
	})
	if err != nil {
		return err
	}

	sender := chunk.New(&listFilesSender{stream: stream})

	for parser := lstree.NewParser(cmd); ; {
		entry, err := parser.NextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if entry.Type != lstree.Blob {
			continue
		}

		if err := sender.Send(&gitalypb.ListFilesResponse{Paths: [][]byte{[]byte(entry.Path)}}); err != nil {
			return err
		}
	}

	return sender.Flush()
}

type listFilesSender struct {
	stream   gitalypb.CommitService_ListFilesServer
	response *gitalypb.ListFilesResponse
}

func (s *listFilesSender) Reset()      { s.response = &gitalypb.ListFilesResponse{} }
func (s *listFilesSender) Send() error { return s.stream.Send(s.response) }
func (s *listFilesSender) Append(m proto.Message) {
	s.response.Paths = append(s.response.Paths, m.(*gitalypb.ListFilesResponse).Paths...)
}
