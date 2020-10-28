package diff

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/diff"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/codes"
)

func TestSuccessfulDiffTreeDiffStatsRequest(t *testing.T) {
	server, serverSocketPath := runDiffServer(t)
	defer server.Stop()

	client, conn := newDiffClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	commits := []string{"e4003da16c1c2c3fc4567700121b17bf8e591c6c", "57290e673a4c87f51294f5216672cbc58d485d25", "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"}
	rpcRequest := &gitalypb.DiffTreeDiffStatsRequest{Repository: testRepo, Commits: commits}

	ctx, cancel := testhelper.Context()
	defer cancel()

	expectedStats := []diff.NumStat{
		{
			Path:      []byte("CONTRIBUTING.md"),
			Additions: 1,
			Deletions: 1,
		},
		{
			Path:      []byte("MAINTENANCE.md"),
			Additions: 1,
			Deletions: 1,
		},
		{
			Path:      []byte("README.md"),
			Additions: 1,
			Deletions: 1,
		},
		{
			Path:      []byte("gitaly/tab\tnewline\n file"),
			Additions: 1,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/deleted-file"),
			Additions: 0,
			Deletions: 1,
		},
		{
			Path:      []byte("gitaly/file-with-multiple-chunks"),
			Additions: 28,
			Deletions: 23,
		},
		{
			Path:      []byte("gitaly/logo-white.png"),
			Additions: 0,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/mode-file"),
			Additions: 0,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/mode-file-with-mods"),
			Additions: 2,
			Deletions: 1,
		},
		{
			Path:      []byte("gitaly/named-file-with-mods"),
			Additions: 0,
			Deletions: 1,
		},
		{
			Path:      []byte("gitaly/no-newline-at-the-end"),
			Additions: 1,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/renamed-file"),
			Additions: 0,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/renamed-file-with-mods"),
			Additions: 1,
			Deletions: 0,
		},
		{
			Path:      []byte("gitaly/テスト.txt"),
			Additions: 0,
			Deletions: 0,
		},
	}

	stream, err := client.DiffTreeDiffStats(ctx, rpcRequest)
	if err != nil {
		t.Fatal(err)
	}

	for {
		fetchedStats, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)

		stats := fetchedStats.GetStats()

		for index, fetchedStat := range stats {
			expectedStat := expectedStats[index]

			require.Equal(t, expectedStat.Path, fetchedStat.Path)
			require.Equal(t, expectedStat.Additions, fetchedStat.Additions)
			require.Equal(t, expectedStat.Deletions, fetchedStat.Deletions)
		}
	}
}

func TestFailedDiffTreeDiffStatsRequest(t *testing.T) {
	server, serverSocketPath := runDiffServer(t)
	defer server.Stop()

	client, conn := newDiffClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := testhelper.Context()
	defer cancel()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	tests := []struct {
		desc    string
		repo    *gitalypb.Repository
		commits []string
		err     codes.Code
	}{
		{
			desc:    "repo not found",
			repo:    &gitalypb.Repository{StorageName: testRepo.GetStorageName(), RelativePath: "bar.git"},
			commits: []string{"e4003da16c1c2c3fc4567700121b17bf8e591c6c", "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"},
			err:     codes.NotFound,
		},
		{
			desc:    "storage not found",
			repo:    &gitalypb.Repository{StorageName: "foo", RelativePath: "bar.git"},
			commits: []string{"e4003da16c1c2c3fc4567700121b17bf8e591c6c", "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"},
			err:     codes.InvalidArgument,
		},
		{
			desc:    "must have more than 1 commit",
			repo:    testRepo,
			commits: []string{"e4003da16c1c2c3fc4567700121b17bf8e591c6c"},
			err:     codes.InvalidArgument,
		},
		{
			desc:    "commits cannot contain an empty commit",
			repo:    testRepo,
			commits: []string{"8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab", ""},
			err:     codes.InvalidArgument,
		},
		{
			desc:    "invalid commit",
			repo:    testRepo,
			commits: []string{"invalidinvalidinvalid", "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"},
			err:     codes.Unavailable,
		},
		{
			desc:    "commit not found",
			repo:    testRepo,
			commits: []string{"z4003da16c1c2c3fc4567700121b17bf8e591c6c", "8a0f2ee90d940bfb0ba1e14e8214b0649056e4ab"},
			err:     codes.Unavailable,
		},
	}

	for _, tc := range tests {
		rpcRequest := &gitalypb.DiffTreeDiffStatsRequest{Repository: tc.repo, Commits: tc.commits}
		stream, err := client.DiffTreeDiffStats(ctx, rpcRequest)
		require.NoError(t, err)

		t.Run(tc.desc, func(t *testing.T) {
			_, err := stream.Recv()

			testhelper.RequireGrpcError(t, err, tc.err)
		})
	}
}
