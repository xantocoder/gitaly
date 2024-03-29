package stats

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestRepositoryProfile(t *testing.T) {
	testRepo, testRepoPath, cleanup := testhelper.InitBareRepo(t)
	defer cleanup()

	ctx, cancel := testhelper.Context()
	defer cancel()

	hasBitmap, err := HasBitmap(testRepo)
	require.NoError(t, err)
	require.False(t, hasBitmap, "repository should not have a bitmap initially")
	unpackedObjects, err := UnpackedObjects(testRepo)
	require.NoError(t, err)
	require.Zero(t, unpackedObjects)
	packfiles, err := GetPackfiles(testRepoPath)
	require.NoError(t, err)
	require.Empty(t, packfiles)
	packfilesCount, err := PackfilesCount(testRepo)
	require.NoError(t, err)
	require.Zero(t, packfilesCount)

	blobs := 10
	blobIDs := testhelper.WriteBlobs(t, testRepoPath, blobs)

	unpackedObjects, err = UnpackedObjects(testRepo)
	require.NoError(t, err)
	require.Equal(t, int64(blobs), unpackedObjects)

	looseObjects, err := LooseObjects(ctx, testRepo)
	require.NoError(t, err)
	require.Equal(t, int64(blobs), looseObjects)

	for _, blobID := range blobIDs {
		commitID := testhelper.CommitBlobWithName(t, testRepoPath, blobID, blobID, "adding another blob....")
		testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "update-ref", "refs/heads/"+blobID, commitID)
	}

	// write a loose object
	testhelper.WriteBlobs(t, testRepoPath, 1)

	testhelper.MustRunCommand(t, nil, "git", "-C", testRepoPath, "repack", "-A", "-b", "-d")

	unpackedObjects, err = UnpackedObjects(testRepo)
	require.NoError(t, err)
	require.Zero(t, unpackedObjects)
	looseObjects, err = LooseObjects(ctx, testRepo)
	require.NoError(t, err)
	require.Equal(t, int64(1), looseObjects)

	// let a ms elapse for the OS to recognize the blobs have been written after the packfile
	time.Sleep(1 * time.Millisecond)

	// write another loose object
	blobID := testhelper.WriteBlobs(t, testRepoPath, 1)[0]

	// due to OS semantics, ensure that the blob has a timestamp that is after the packfile
	theFuture := time.Now().Add(10 * time.Minute)
	require.NoError(t, os.Chtimes(filepath.Join(testRepoPath, "objects", blobID[0:2], blobID[2:]), theFuture, theFuture))

	unpackedObjects, err = UnpackedObjects(testRepo)
	require.NoError(t, err)
	require.Equal(t, int64(1), unpackedObjects)

	looseObjects, err = LooseObjects(ctx, testRepo)
	require.NoError(t, err)
	require.Equal(t, int64(2), looseObjects)
}
