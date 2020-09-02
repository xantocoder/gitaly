package smarthttp

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/pktline"
	"gitlab.com/gitlab-org/gitaly/internal/git/stats"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/helper/chunk"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/streamio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	uploadPackSvc  = "upload-pack"
	receivePackSvc = "receive-pack"
)

func (s *server) InfoRefsUploadPack(in *gitalypb.InfoRefsRequest, stream gitalypb.SmartHTTPService_InfoRefsUploadPackServer) error {
	w := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&gitalypb.InfoRefsResponse{Data: p})
	})

	return tryCache(stream.Context(), in, w, func(w io.Writer) error {
		return handleInfoRefs(stream.Context(), uploadPackSvc, in, w)
	})
}

func (s *server) ParsedInfoRefsUploadPack(in *gitalypb.ParsedInfoRefsRequest, stream gitalypb.SmartHTTPService_ParsedInfoRefsUploadPackServer) error {
	pr, pw := io.Pipe()

	// if the current goroutine doesn't fully consume pr for any reason, this line will ensure that
	// the started goroutine is unblocked as it will error out trying to write to pw.
	defer pr.Close()
	go func() {
		req := &gitalypb.InfoRefsRequest{
			Repository:       in.Repository,
			GitConfigOptions: in.GitConfigOptions,
			GitProtocol:      in.GitProtocol,
		}
		err := tryCache(stream.Context(), req, pw, func(w io.Writer) error {
			return handleInfoRefs(stream.Context(), uploadPackSvc, req, w)
		})
		_ = pw.CloseWithError(err) // always returns nil
	}()
	// NB: this is not really streaming because ParseReferenceDiscovery fully parses the stream into memory
	rd, err := stats.ParseReferenceDiscovery(pr)
	if err != nil {
		if _, isStatus := status.FromError(err); isStatus {
			return err
		}
		return status.Errorf(codes.Internal, "ParseReferenceDiscovery: %v", err)
	}
	chunker := chunk.New(&parsedInfoRefsUploadPackSender{
		stream:       stream,
		capabilities: rd.Caps,
	})
	for _, ref := range rd.Refs {
		err = chunker.Send(&gitalypb.Ref{
			Name:     []byte(ref.Name),
			CommitId: ref.Oid,
		})
		if err != nil {
			return err
		}
	}
	return chunker.Flush()
}

func (s *server) InfoRefsReceivePack(in *gitalypb.InfoRefsRequest, stream gitalypb.SmartHTTPService_InfoRefsReceivePackServer) error {
	w := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&gitalypb.InfoRefsResponse{Data: p})
	})
	return handleInfoRefs(stream.Context(), receivePackSvc, in, w)
}

func handleInfoRefs(ctx context.Context, service string, req *gitalypb.InfoRefsRequest, w io.Writer) error {
	ctxlogrus.Extract(ctx).WithFields(log.Fields{
		"service": service,
	}).Debug("handleInfoRefs")

	env := git.AddGitProtocolEnv(ctx, req, []string{})

	repoPath, err := helper.GetRepoPath(req.Repository)
	if err != nil {
		return err
	}

	var globalOpts []git.Option
	if service == "receive-pack" {
		globalOpts = append(globalOpts, git.ReceivePackConfig()...)
	}

	if service == "upload-pack" {
		globalOpts = append(globalOpts, git.UploadPackFilterConfig()...)
	}

	for _, o := range req.GitConfigOptions {
		globalOpts = append(globalOpts, git.ValueFlag{"-c", o})
	}

	cmd, err := git.SafeBareCmd(ctx, git.CmdStream{}, env, globalOpts, git.SubCmd{
		Name:  service,
		Flags: []git.Option{git.Flag{Name: "--stateless-rpc"}, git.Flag{Name: "--advertise-refs"}},
		Args:  []string{repoPath},
	})

	if err != nil {
		if _, ok := status.FromError(err); ok {
			return err
		}
		return status.Errorf(codes.Internal, "GetInfoRefs: cmd: %v", err)
	}

	if _, err := pktline.WriteString(w, fmt.Sprintf("# service=git-%s\n", service)); err != nil {
		return status.Errorf(codes.Internal, "GetInfoRefs: pktLine: %v", err)
	}

	if err := pktline.WriteFlush(w); err != nil {
		return status.Errorf(codes.Internal, "GetInfoRefs: pktFlush: %v", err)
	}

	if _, err := io.Copy(w, cmd); err != nil {
		return status.Errorf(codes.Internal, "GetInfoRefs: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		return status.Errorf(codes.Internal, "GetInfoRefs: %v", err)
	}

	return nil
}

type parsedInfoRefsUploadPackSender struct {
	stream gitalypb.SmartHTTPService_ParsedInfoRefsUploadPackServer
	// capabilities is sent in the first message only.
	capabilities []string
	refs         []*gitalypb.Ref
}

func (sender *parsedInfoRefsUploadPackSender) Reset() {
	sender.refs = nil
}

func (sender *parsedInfoRefsUploadPackSender) Append(m proto.Message) {
	sender.refs = append(sender.refs, m.(*gitalypb.Ref))
}

func (sender *parsedInfoRefsUploadPackSender) Send() error {
	// Reset the field on Send(). Cannot do this in Reset() because it's called
	// before accumulation starts.
	capabilities := sender.capabilities
	sender.capabilities = nil
	return sender.stream.Send(&gitalypb.ParsedInfoRefsResponse{
		Capabilities: capabilities,
		Refs:         sender.refs,
	})
}
