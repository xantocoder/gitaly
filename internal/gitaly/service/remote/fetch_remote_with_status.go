package remote

import (
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

func (s *server) FetchRemoteWithStatus(in *gitalypb.FetchRemoteWithStatusRequest, stream gitalypb.RemoteService_FetchRemoteWithStatusServer) error {
	ctx := stream.Context()

	client, err := s.ruby.RepositoryServiceClient(ctx)
	if err != nil {
		return err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, in.GetRepository())
	if err != nil {
		return err
	}

	// TODO: gitaly-ruby implementation in RemoteService. Will mean we don't need
	// to do this copy. We don't want RepositoryService.FetchRemote to stream
	// a list of refs from gitaly-ruby only to be thrown away.
	newIn := &gitalypb.FetchRemoteRequest{
		Repository:   in.GetRepository(),
		Remote:       in.GetRemote(),
		Force:        in.GetForce(),
		NoTags:       in.GetNoTags(),
		Timeout:      in.GetTimeout(),
		SshKey:       in.GetSshKey(),
		KnownHosts:   in.GetKnownHosts(),
		NoPrune:      in.GetNoPrune(),
		RemoteParams: in.GetRemoteParams(),
	}

	_, err = client.FetchRemote(clientCtx, newIn)
	return err
}
