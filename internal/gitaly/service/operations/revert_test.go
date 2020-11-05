package operations

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/git/log"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
)

func TestServer_UserRevert_successful(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertSuccessful)
}

func testServerUserRevertSuccessful(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	destinationBranch := "revert-dst"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", destinationBranch, "master")

	masterHeadCommit, err := log.GetCommit(ctxOuter, testRepo, "master")
	require.NoError(t, err)

	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, "d59c60028b053793cecfb4022de34602e1a9218e")
	require.NoError(t, err)

	testRepoCopy, testRepoCopyPath, cleanup := testhelper.NewTestRepo(t) // read-only repo
	defer cleanup()

	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoCopyPath, "branch", destinationBranch, "master")

	testCases := []struct {
		desc         string
		request      *gitalypb.UserRevertRequest
		branchUpdate *gitalypb.OperationBranchUpdate
	}{
		{
			desc: "branch exists",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepo,
				User:       testhelper.TestUser,
				Commit:     revertedCommit,
				BranchName: []byte(destinationBranch),
				Message:    []byte("Reverting " + revertedCommit.Id),
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{},
		},
		{
			desc: "nonexistent branch + start_repository == repository",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepo,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-1"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartBranchName: []byte("master"),
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
		{
			desc: "nonexistent branch + start_repository != repository",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepo,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-2"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartRepository: testRepoCopy,
				StartBranchName: []byte("master"),
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
		{
			desc: "nonexistent branch + empty start_repository",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepo,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-3"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartBranchName: []byte("master"),
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
		{
			desc: "branch exists with dry run",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepoCopy,
				User:       testhelper.TestUser,
				Commit:     revertedCommit,
				BranchName: []byte(destinationBranch),
				Message:    []byte("Reverting " + revertedCommit.Id),
				DryRun:     true,
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{},
		},
		{
			desc: "nonexistent branch + start_repository == repository with dry run",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepoCopy,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-1"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartBranchName: []byte("master"),
				DryRun:          true,
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
		{
			desc: "nonexistent branch + start_repository != repository with dry run",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepoCopy,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-2"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartRepository: testRepoCopy,
				StartBranchName: []byte("master"),
				DryRun:          true,
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
		{
			desc: "nonexistent branch + empty start_repository with dry run",
			request: &gitalypb.UserRevertRequest{
				Repository:      testRepoCopy,
				User:            testhelper.TestUser,
				Commit:          revertedCommit,
				BranchName:      []byte("to-be-reverted-into-3"),
				Message:         []byte("Reverting " + revertedCommit.Id),
				StartBranchName: []byte("master"),
				DryRun:          true,
			},
			branchUpdate: &gitalypb.OperationBranchUpdate{BranchCreated: true},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			md := testhelper.GitalyServersMetadata(t, serverSocketPath)
			ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

			response, err := client.UserRevert(ctx, testCase.request)
			require.NoError(t, err)

			headCommit, err := log.GetCommit(ctx, testCase.request.Repository, string(testCase.request.BranchName))
			require.NoError(t, err)

			expectedBranchUpdate := testCase.branchUpdate
			expectedBranchUpdate.CommitId = headCommit.Id

			require.Equal(t, expectedBranchUpdate, response.BranchUpdate)
			require.Empty(t, response.CreateTreeError)
			require.Empty(t, response.CreateTreeErrorCode)

			if testCase.request.DryRun {
				require.Equal(t, masterHeadCommit.Subject, headCommit.Subject)
				require.Equal(t, masterHeadCommit.Id, headCommit.Id)
			} else {
				require.Equal(t, testCase.request.Message, headCommit.Subject)
				require.Equal(t, masterHeadCommit.Id, headCommit.ParentIds[0])
			}
		})
	}
}

func TestServer_UserRevert_successful_git_hooks(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertSuccessfulGitHooks)
}

func testServerUserRevertSuccessfulGitHooks(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	destinationBranch := "revert-dst"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", destinationBranch, "master")

	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, "d59c60028b053793cecfb4022de34602e1a9218e")
	require.NoError(t, err)

	request := &gitalypb.UserRevertRequest{
		Repository: testRepo,
		User:       testhelper.TestUser,
		Commit:     revertedCommit,
		BranchName: []byte(destinationBranch),
		Message:    []byte("Reverting " + revertedCommit.Id),
	}

	var hookOutputFiles []string
	for _, hookName := range GitlabHooks {
		hookOutputTempPath, cleanup := testhelper.WriteEnvToCustomHook(t, testRepoPath, hookName)
		defer cleanup()
		hookOutputFiles = append(hookOutputFiles, hookOutputTempPath)
	}

	md := testhelper.GitalyServersMetadata(t, serverSocketPath)
	ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

	response, err := client.UserRevert(ctx, request)
	require.NoError(t, err)
	require.Empty(t, response.PreReceiveError)

	for _, file := range hookOutputFiles {
		output := string(testhelper.MustReadFile(t, file))
		require.Contains(t, output, "GL_USERNAME="+testhelper.TestUser.GlUsername)
	}
}

func TestServer_UserRevert_failued_due_to_validations(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertFailuedDueToValidations)
}

func testServerUserRevertFailuedDueToValidations(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, "d59c60028b053793cecfb4022de34602e1a9218e")
	require.NoError(t, err)

	destinationBranch := "revert-dst"

	testCases := []struct {
		desc    string
		request *gitalypb.UserRevertRequest
		code    codes.Code
	}{
		{
			desc: "empty user",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepo,
				User:       nil,
				Commit:     revertedCommit,
				BranchName: []byte(destinationBranch),
				Message:    []byte("Reverting " + revertedCommit.Id),
			},
			code: codes.InvalidArgument,
		},
		{
			desc: "empty commit",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepo,
				User:       testhelper.TestUser,
				Commit:     nil,
				BranchName: []byte(destinationBranch),
				Message:    []byte("Reverting " + revertedCommit.Id),
			},
			code: codes.InvalidArgument,
		},
		{
			desc: "empty branch name",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepo,
				User:       testhelper.TestUser,
				Commit:     revertedCommit,
				BranchName: nil,
				Message:    []byte("Reverting " + revertedCommit.Id),
			},
			code: codes.InvalidArgument,
		},
		{
			desc: "empty message",
			request: &gitalypb.UserRevertRequest{
				Repository: testRepo,
				User:       testhelper.TestUser,
				Commit:     revertedCommit,
				BranchName: []byte(destinationBranch),
				Message:    nil,
			},
			code: codes.InvalidArgument,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			md := testhelper.GitalyServersMetadata(t, serverSocketPath)
			ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

			_, err := client.UserRevert(ctx, testCase.request)
			testhelper.RequireGrpcError(t, err, testCase.code)
		})
	}
}

func TestServer_UserRevert_failed_due_to_pre_receive_error(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertFailedDueToPreReceiveError)
}

func testServerUserRevertFailedDueToPreReceiveError(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	destinationBranch := "revert-dst"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", destinationBranch, "master")

	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, "d59c60028b053793cecfb4022de34602e1a9218e")
	require.NoError(t, err)

	request := &gitalypb.UserRevertRequest{
		Repository: testRepo,
		User:       testhelper.TestUser,
		Commit:     revertedCommit,
		BranchName: []byte(destinationBranch),
		Message:    []byte("Reverting " + revertedCommit.Id),
	}

	hookContent := []byte("#!/bin/sh\necho GL_ID=$GL_ID\nexit 1")

	for _, hookName := range GitlabPreHooks {
		t.Run(hookName, func(t *testing.T) {
			remove, err := testhelper.WriteCustomHook(testRepoPath, hookName, hookContent)
			require.NoError(t, err)
			defer remove()

			md := testhelper.GitalyServersMetadata(t, serverSocketPath)
			ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

			response, err := client.UserRevert(ctx, request)
			require.NoError(t, err)
			require.Contains(t, response.PreReceiveError, "GL_ID="+testhelper.TestUser.GlId)
		})
	}
}

func TestServer_UserRevert_failed_due_to_create_tree_error(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertFailedDueToCreateTreeError)
}

func testServerUserRevertFailedDueToCreateTreeError(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	destinationBranch := "revert-dst"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", destinationBranch, "master")

	// This revert patch of the following commit cannot be applied to the destinationBranch above
	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, "372ab6950519549b14d220271ee2322caa44d4eb")
	require.NoError(t, err)

	request := &gitalypb.UserRevertRequest{
		Repository: testRepo,
		User:       testhelper.TestUser,
		Commit:     revertedCommit,
		BranchName: []byte(destinationBranch),
		Message:    []byte("Reverting " + revertedCommit.Id),
	}

	md := testhelper.GitalyServersMetadata(t, serverSocketPath)
	ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

	response, err := client.UserRevert(ctx, request)
	require.NoError(t, err)
	require.NotEmpty(t, response.CreateTreeError)
	require.Equal(t, gitalypb.UserRevertResponse_CONFLICT, response.CreateTreeErrorCode)
}

func TestServer_UserRevert_failed_due_to_commit_error(t *testing.T) {
	testWithFeature(t, featureflag.GoUserRevert, testServerUserRevertFailedDueToCommitError)
}

func testServerUserRevertFailedDueToCommitError(t *testing.T, ctxOuter context.Context) {
	serverSocketPath, stop := runOperationServiceServer(t)
	defer stop()

	client, conn := newOperationClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, testRepoPath, cleanup := testhelper.NewTestRepo(t)
	defer cleanup()

	sourceBranch := "revert-src"
	destinationBranch := "revert-dst"
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", destinationBranch, "master")
	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "branch", sourceBranch, "a5391128b0ef5d21df5dd23d98557f4ef12fae20")

	revertedCommit, err := log.GetCommit(ctxOuter, testRepo, sourceBranch)
	require.NoError(t, err)

	request := &gitalypb.UserRevertRequest{
		Repository:      testRepo,
		User:            testhelper.TestUser,
		Commit:          revertedCommit,
		BranchName:      []byte(destinationBranch),
		Message:         []byte("Reverting " + revertedCommit.Id),
		StartBranchName: []byte(sourceBranch),
	}

	md := testhelper.GitalyServersMetadata(t, serverSocketPath)
	ctx := testhelper.MergeOutgoingMetadata(ctxOuter, md)

	response, err := client.UserRevert(ctx, request)
	require.NoError(t, err)
	require.Equal(t, "Branch diverged", response.CommitError)
}
