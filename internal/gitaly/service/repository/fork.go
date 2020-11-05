package repository

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitalyssh"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) CreateFork(ctx context.Context, req *gitalypb.CreateForkRequest) (*gitalypb.CreateForkResponse, error) {
	targetRepository := req.Repository
	sourceRepository := req.SourceRepository

	if sourceRepository == nil {
		return nil, status.Errorf(codes.InvalidArgument, "CreateFork: empty SourceRepository")
	}
	if targetRepository == nil {
		return nil, status.Errorf(codes.InvalidArgument, "CreateFork: empty Repository")
	}

	targetRepositoryFullPath, err := s.locator.GetPath(targetRepository)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(targetRepositoryFullPath); !os.IsNotExist(err) {
		return nil, status.Errorf(codes.InvalidArgument, "CreateFork: dest dir exists")
	}

	if err := os.MkdirAll(targetRepositoryFullPath, 0770); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFork: create dest dir: %v", err)
	}

	env, err := gitalyssh.UploadPackEnv(ctx, &gitalypb.SSHUploadPackRequest{Repository: sourceRepository})
	if err != nil {
		return nil, err
	}

	cmd, err := git.SafeBareCmd(ctx, git.CmdStream{}, env, nil,
		git.SubCmd{
			Name: "clone",
			Flags: []git.Option{
				git.Flag{Name: "--bare"},
				git.Flag{Name: "--no-local"},
			},
			PostSepArgs: []string{
				fmt.Sprintf("%s:%s", gitalyssh.GitalyInternalURL, sourceRepository.RelativePath),
				targetRepositoryFullPath,
			},
		},
		git.WithRefTxHook(ctx, req.Repository, s.cfg),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFork: clone cmd start: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFork: clone cmd wait: %v", err)
	}

	if err := s.removeOriginInRepo(ctx, targetRepository); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFork: %v", err)
	}

	// CreateRepository is harmless on existing repositories with the side effect that it creates the hook symlink.
	if _, err := s.CreateRepository(ctx, &gitalypb.CreateRepositoryRequest{Repository: targetRepository}); err != nil {
		return nil, status.Errorf(codes.Internal, "CreateFork: create hooks failed: %v", err)
	}

	return &gitalypb.CreateForkResponse{}, nil
}
