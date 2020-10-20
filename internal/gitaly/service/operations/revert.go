package operations

import (
	"context"
	"errors"
	"fmt"

	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/repository"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/gitalyssh"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const gitalyInternalURL = "ssh://gitaly/internal.git"

func (s *server) UserRevert(ctx context.Context, req *gitalypb.UserRevertRequest) (*gitalypb.UserRevertResponse, error) {
	if err := validateCherryPickOrRevertRequest(req); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "UserRevert: %v", err)
	}
	if featureflag.IsDisabled(ctx, featureflag.GoUserRevert) {
		return s.rubyUserRevert(ctx, req)
	}

	startBranchName := req.StartBranchName
	if len(startBranchName) == 0 {
		startBranchName = req.BranchName
	}

	startRepository := req.StartRepository
	if startRepository == nil {
		startRepository = req.Repository
	}

	startRevision, err := git.NewRepository(startRepository).ResolveRefish(ctx, string(startBranchName))
	if err != nil {
		return nil, helper.ErrInvalidArgument(fmt.Errorf("resolve start ref: %w", err))
	}

	if req.StartRepository != nil {
		if err := fetchOid(ctx, req.Repository, req.StartRepository, startRevision); err != nil {
			return nil, fmt.Errorf("fetch OID: %w", err)
		}
	}

	repoPath, err := s.locator.GetPath(req.Repository)
	if err != nil {
		return nil, fmt.Errorf("get path: %w", err)
	}

	revert, err := git2go.RevertCommand{
		Repository: repoPath,
		AuthorName: string(req.User.Name),
		AuthorMail: string(req.User.Email),
		Message:    string(req.Message),
		Ours:       startRevision,
		Revert:     req.Commit.Id,
	}.Run(ctx, s.cfg)
	if err != nil {
		if errors.Is(err, git2go.ErrInvalidArgument) {
			return nil, helper.ErrInvalidArgument(err)
		}
		return nil, fmt.Errorf("revert command: %w", err)
	}

	branchCreated := false
	revision, err := git.NewRepository(req.Repository).ResolveRefish(ctx, string(req.BranchName))
	if errors.Is(err, git.ErrReferenceNotFound) {
		branchCreated = true
	} else if err != nil {
		return nil, helper.ErrInvalidArgument(fmt.Errorf("resolve ref: %w", err))
	}

	if req.DryRun {
		revert.CommitID = startRevision
	}

	branch := fmt.Sprintf("refs/heads/%s", req.BranchName)
	if err := s.updateReferenceWithHooks(ctx, req.Repository, req.User, branch, revert.CommitID, revision); err != nil {
		var preReceiveError preReceiveError
		if errors.As(err, &preReceiveError) {
			return &gitalypb.UserRevertResponse{
				PreReceiveError: preReceiveError.message,
			}, nil
		}

		var updateRefError updateRefError
		if errors.As(err, &updateRefError) {
			// When an error happens updating the reference, e.g. because of a race
			// with another update, then Ruby code didn't send an error but just an
			// empty response.
			return &gitalypb.UserRevertResponse{}, nil
		}

		return nil, fmt.Errorf("update reference with hooks: %w", err)
	}

	return &gitalypb.UserRevertResponse{
		BranchUpdate: &gitalypb.OperationBranchUpdate{
			CommitId:      revert.CommitID,
			BranchCreated: branchCreated,
		},
	}, nil
}

func (s *server) rubyUserRevert(ctx context.Context, req *gitalypb.UserRevertRequest) (*gitalypb.UserRevertResponse, error) {
	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, req.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserRevert(clientCtx, req)
}

func fetchOid(ctx context.Context, dst repository.GitRepo, src *gitalypb.Repository, oid string) error {
	env, err := gitalyssh.UploadPackEnv(ctx, &gitalypb.SSHUploadPackRequest{Repository: src})
	if err != nil {
		return err
	}
	cmd, err := git.SafeCmdWithEnv(ctx, env, dst, nil,
		git.SubCmd{
			Name:  "fetch",
			Flags: []git.Option{git.Flag{Name: "--no-tags"}},
			Args:  []string{gitalyInternalURL, oid},
		},
	)
	if err != nil {
		return err
	}
	return cmd.Wait()
}
