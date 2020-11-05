package operations

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/updateref"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/hook"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/gitlabshell"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/helper/text"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/metadata"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

type preReceiveError struct {
	message string
}

func (e preReceiveError) Error() string {
	return e.message
}

type updateRefError struct {
	reference string
}

func (e updateRefError) Error() string {
	return fmt.Sprintf("Could not update %s. Please refresh and try again.", e.reference)
}

func (s *server) UserMergeBranch(bidi gitalypb.OperationService_UserMergeBranchServer) error {
	ctx := bidi.Context()

	if featureflag.IsEnabled(ctx, featureflag.GoUserMergeBranch) {
		return s.userMergeBranch(bidi)
	}

	firstRequest, err := bidi.Recv()
	if err != nil {
		return err
	}

	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, firstRequest.GetRepository())
	if err != nil {
		return err
	}

	rubyBidi, err := client.UserMergeBranch(clientCtx)
	if err != nil {
		return err
	}

	if err := rubyBidi.Send(firstRequest); err != nil {
		return err
	}

	return rubyserver.ProxyBidi(
		func() error {
			request, err := bidi.Recv()
			if err != nil {
				return err
			}

			return rubyBidi.Send(request)
		},
		rubyBidi,
		func() error {
			response, err := rubyBidi.Recv()
			if err != nil {
				return err
			}

			return bidi.Send(response)
		},
	)
}

func validateMergeBranchRequest(request *gitalypb.UserMergeBranchRequest) error {
	if request.User == nil {
		return fmt.Errorf("empty user")
	}

	if len(request.Branch) == 0 {
		return fmt.Errorf("empty branch name")
	}

	if request.CommitId == "" {
		return fmt.Errorf("empty commit ID")
	}

	if len(request.Message) == 0 {
		return fmt.Errorf("empty message")
	}

	return nil
}

func (s *server) updateReferenceWithHooks(ctx context.Context, repo *gitalypb.Repository, user *gitalypb.User, reference, newrev, oldrev string) error {
	gitlabshellEnv, err := gitlabshell.Env()
	if err != nil {
		return err
	}

	env := append([]string{
		fmt.Sprintf("GL_ID=%s", user.GetGlId()),
		fmt.Sprintf("GL_USERNAME=%s", user.GetGlUsername()),
		fmt.Sprintf("GL_REPOSITORY=%s", repo.GetGlRepository()),
		fmt.Sprintf("GL_PROJECT_PATH=%s", repo.GetGlProjectPath()),
		fmt.Sprintf("GITALY_SOCKET=" + config.GitalyInternalSocketPath()),
		fmt.Sprintf("GITALY_REPO=%s", repo),
		fmt.Sprintf("GITALY_TOKEN=%s", s.cfg.Auth.Token),
	}, gitlabshellEnv...)

	transaction, err := metadata.TransactionFromContext(ctx)
	if err != nil {
		if err != metadata.ErrTransactionNotFound {
			return err
		}
	}

	if err == nil {
		praefect, err := metadata.PraefectFromContext(ctx)
		if err != nil {
			return err
		}

		transactionEnv, err := transaction.Env()
		if err != nil {
			return err
		}

		praefectEnv, err := praefect.Env()
		if err != nil {
			return err
		}

		env = append(env, transactionEnv, praefectEnv)
	}

	changes := fmt.Sprintf("%s %s %s\n", oldrev, newrev, reference)
	var stdout, stderr bytes.Buffer

	if err := s.hookManager.PreReceiveHook(ctx, repo, env, strings.NewReader(changes), &stdout, &stderr); err != nil {
		return preReceiveError{message: stdout.String()}
	}
	if err := s.hookManager.UpdateHook(ctx, repo, reference, oldrev, newrev, env, &stdout, &stderr); err != nil {
		return preReceiveError{message: stdout.String()}
	}

	// For backwards compatibility with Ruby, we need to only call the reference-transaction
	// hook if the corresponding Ruby feature flag is set.
	if featureflag.IsEnabled(ctx, featureflag.RubyReferenceTransactionHook) {
		if err := s.hookManager.ReferenceTransactionHook(ctx, hook.ReferenceTransactionPrepared, env, strings.NewReader(changes)); err != nil {
			return preReceiveError{message: stdout.String()}
		}
	}

	updater, err := updateref.New(ctx, repo)
	if err != nil {
		return err
	}

	if err := updater.Update(reference, newrev, oldrev); err != nil {
		return err
	}

	if err := updater.Wait(); err != nil {
		return updateRefError{reference: reference}
	}

	if err := s.hookManager.PostReceiveHook(ctx, repo, nil, env, strings.NewReader(changes), &stdout, &stderr); err != nil {
		return err
	}

	return nil
}

func (s *server) userMergeBranch(stream gitalypb.OperationService_UserMergeBranchServer) error {
	ctx := stream.Context()

	firstRequest, err := stream.Recv()
	if err != nil {
		return err
	}

	if err := validateMergeBranchRequest(firstRequest); err != nil {
		return helper.ErrInvalidArgument(err)
	}

	repo := firstRequest.Repository
	repoPath, err := s.locator.GetPath(repo)
	if err != nil {
		return err
	}

	revision, err := git.NewRepository(repo).ResolveRefish(ctx, string(firstRequest.Branch))
	if err != nil {
		return err
	}

	merge, err := git2go.MergeCommand{
		Repository: repoPath,
		AuthorName: string(firstRequest.User.Name),
		AuthorMail: string(firstRequest.User.Email),
		Message:    string(firstRequest.Message),
		Ours:       firstRequest.CommitId,
		Theirs:     revision,
	}.Run(ctx, s.cfg)
	if err != nil {
		if errors.Is(err, git2go.ErrInvalidArgument) {
			return helper.ErrInvalidArgument(err)
		}
		return err
	}

	if err := stream.Send(&gitalypb.UserMergeBranchResponse{
		CommitId: merge.CommitID,
	}); err != nil {
		return err
	}

	secondRequest, err := stream.Recv()
	if err != nil {
		return err
	}
	if !secondRequest.Apply {
		return helper.ErrPreconditionFailedf("merge aborted by client")
	}

	branch := "refs/heads/" + text.ChompBytes(firstRequest.Branch)
	if err := s.updateReferenceWithHooks(ctx, firstRequest.Repository, firstRequest.User, branch, merge.CommitID, revision); err != nil {
		var preReceiveError preReceiveError
		var updateRefError updateRefError

		if errors.As(err, &preReceiveError) {
			err = stream.Send(&gitalypb.UserMergeBranchResponse{
				PreReceiveError: preReceiveError.message,
			})
		} else if errors.As(err, &updateRefError) {
			// When an error happens updating the reference, e.g. because of a race
			// with another update, then Ruby code didn't send an error but just an
			// empty response.
			err = stream.Send(&gitalypb.UserMergeBranchResponse{})
		}

		return err
	}

	if err := stream.Send(&gitalypb.UserMergeBranchResponse{
		BranchUpdate: &gitalypb.OperationBranchUpdate{
			CommitId:      merge.CommitID,
			RepoCreated:   false,
			BranchCreated: false,
		},
	}); err != nil {
		return err
	}

	return nil
}

func validateFFRequest(in *gitalypb.UserFFBranchRequest) error {
	if len(in.Branch) == 0 {
		return fmt.Errorf("empty branch name")
	}

	if in.User == nil {
		return fmt.Errorf("empty user")
	}

	if in.CommitId == "" {
		return fmt.Errorf("empty commit id")
	}

	return nil
}

func (s *server) UserFFBranch(ctx context.Context, in *gitalypb.UserFFBranchRequest) (*gitalypb.UserFFBranchResponse, error) {
	if err := validateFFRequest(in); err != nil {
		return nil, helper.ErrInvalidArgument(err)
	}

	if featureflag.IsDisabled(ctx, featureflag.GoUserFFBranch) {
		return s.userFFBranchRuby(ctx, in)
	}

	revision, err := git.NewRepository(in.Repository).ResolveRefish(ctx, string(in.Branch))
	if err != nil {
		return nil, helper.ErrInvalidArgument(err)
	}

	cmd, err := git.SafeCmd(ctx, in.Repository, nil, git.SubCmd{
		Name:  "merge-base",
		Flags: []git.Option{git.Flag{Name: "--is-ancestor"}},
		Args:  []string{revision, in.CommitId},
	})
	if err != nil {
		return nil, helper.ErrInternal(err)
	}
	if err := cmd.Wait(); err != nil {
		status, ok := command.ExitStatus(err)
		if !ok {
			return nil, helper.ErrInternal(err)
		}
		// --is-ancestor errors are signaled by a non-zero status that is not 1.
		// https://git-scm.com/docs/git-merge-base#Documentation/git-merge-base.txt---is-ancestor
		if status != 1 {
			return nil, helper.ErrInvalidArgument(err)
		}
		return nil, helper.ErrPreconditionFailedf("not fast forward")
	}

	branch := fmt.Sprintf("refs/heads/%s", in.Branch)
	if err := s.updateReferenceWithHooks(ctx, in.Repository, in.User, branch, in.CommitId, revision); err != nil {
		var preReceiveError preReceiveError
		if errors.As(err, &preReceiveError) {
			return &gitalypb.UserFFBranchResponse{
				PreReceiveError: preReceiveError.message,
			}, nil
		}

		var updateRefError updateRefError
		if errors.As(err, &updateRefError) {
			// When an error happens updating the reference, e.g. because of a race
			// with another update, then Ruby code didn't send an error but just an
			// empty response.
			return &gitalypb.UserFFBranchResponse{}, nil
		}

		return nil, err
	}

	return &gitalypb.UserFFBranchResponse{
		BranchUpdate: &gitalypb.OperationBranchUpdate{
			CommitId: in.CommitId,
		},
	}, nil
}

func (s *server) userFFBranchRuby(ctx context.Context, in *gitalypb.UserFFBranchRequest) (*gitalypb.UserFFBranchResponse, error) {
	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, err
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, in.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserFFBranch(clientCtx, in)
}

func validateUserMergeToRefRequest(in *gitalypb.UserMergeToRefRequest) error {
	if len(in.FirstParentRef) == 0 && len(in.Branch) == 0 {
		return fmt.Errorf("empty first parent ref and branch name")
	}

	if in.User == nil {
		return fmt.Errorf("empty user")
	}

	if in.SourceSha == "" {
		return fmt.Errorf("empty source SHA")
	}

	if len(in.TargetRef) == 0 {
		return fmt.Errorf("empty target ref")
	}

	if !strings.HasPrefix(string(in.TargetRef), "refs/merge-requests") {
		return fmt.Errorf("invalid target ref")
	}

	return nil
}

// userMergeToRef overwrites the given TargetRef to point to either Branch or
// FirstParentRef. Afterwards, it performs a merge of SourceSHA with either
// Branch or FirstParentRef and updates TargetRef to the merge commit.
func (s *server) userMergeToRef(ctx context.Context, request *gitalypb.UserMergeToRefRequest) (*gitalypb.UserMergeToRefResponse, error) {
	repoPath, err := s.locator.GetPath(request.Repository)
	if err != nil {
		return nil, err
	}

	repo := git.NewRepository(request.Repository)

	refName := string(request.Branch)
	if request.FirstParentRef != nil {
		refName = string(request.FirstParentRef)
	}

	ref, err := repo.ResolveRefish(ctx, refName)
	if err != nil {
		//nolint:stylecheck
		return nil, helper.ErrInvalidArgument(errors.New("Invalid merge source"))
	}

	sourceRef, err := repo.ResolveRefish(ctx, request.SourceSha)
	if err != nil {
		//nolint:stylecheck
		return nil, helper.ErrInvalidArgument(errors.New("Invalid merge source"))
	}

	// First, overwrite the reference with the target reference.
	if err := repo.UpdateRef(ctx, string(request.TargetRef), ref, ""); err != nil {
		return nil, updateRefError{reference: string(request.TargetRef)}
	}

	// Now, we create the merge commit...
	merge, err := git2go.MergeCommand{
		Repository:     repoPath,
		AuthorName:     string(request.User.Name),
		AuthorMail:     string(request.User.Email),
		Message:        string(request.Message),
		Ours:           ref,
		Theirs:         sourceRef,
		AllowConflicts: request.AllowConflicts,
	}.Run(ctx, s.cfg)
	if err != nil {
		if errors.Is(err, git2go.ErrInvalidArgument) {
			return nil, helper.ErrInvalidArgument(err)
		}
		//nolint:stylecheck
		return nil, helper.ErrPreconditionFailed(fmt.Errorf("Failed to create merge commit for source_sha %s and target_sha %s at %s", sourceRef, string(request.TargetRef), refName))
	}

	// ... and move branch from target ref to the merge commit. The Ruby
	// implementation doesn't invoke hooks, so we don't either.
	if err := repo.UpdateRef(ctx, string(request.TargetRef), merge.CommitID, ref); err != nil {
		//nolint:stylecheck
		return nil, helper.ErrPreconditionFailed(fmt.Errorf("Could not update %s. Please refresh and try again", string(request.TargetRef)))
	}

	return &gitalypb.UserMergeToRefResponse{
		CommitId: merge.CommitID,
	}, nil
}

func (s *server) UserMergeToRef(ctx context.Context, in *gitalypb.UserMergeToRefRequest) (*gitalypb.UserMergeToRefResponse, error) {
	if err := validateUserMergeToRefRequest(in); err != nil {
		return nil, helper.ErrInvalidArgument(err)
	}

	if featureflag.IsEnabled(ctx, featureflag.GoUserMergeToRef) {
		return s.userMergeToRef(ctx, in)
	}

	client, err := s.ruby.OperationServiceClient(ctx)
	if err != nil {
		return nil, helper.ErrInternal(err)
	}

	clientCtx, err := rubyserver.SetHeaders(ctx, in.GetRepository())
	if err != nil {
		return nil, err
	}

	return client.UserMergeToRef(clientCtx, in)
}
