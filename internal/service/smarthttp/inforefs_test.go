package smarthttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/cache"
	"gitlab.com/gitlab-org/gitaly/internal/config"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/objectpool"
	"gitlab.com/gitlab-org/gitaly/internal/git/pktline"
	"gitlab.com/gitlab-org/gitaly/internal/git/stats"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/tempdir"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitaly/streamio"
	"google.golang.org/grpc/codes"
)

const (
	zeroID = "78fb81a02b03f0013360292ec5106763af32c287"
)

func TestSuccessfulInfoRefsUploadPack(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	rpcRequest := &gitalypb.InfoRefsRequest{Repository: testRepo}

	ctx, cancel := testhelper.Context()
	defer cancel()

	assertFunc := func(response []byte) {
		assertGitRefAdvertisement(t, "InfoRefsUploadPack", string(response), "001e# service=git-upload-pack", "0000", []string{
			"003ef4e6814c3e4e7a0de82a9e7cd20c626cc963a2f8 refs/tags/v1.0.0",
			"00416f6d7e7ed97bb5f0054f2b1df789b39ca89b6ff9 refs/tags/v1.0.0^{}",
		})
	}
	t.Run("raw", func(t *testing.T) {
		response, err := makeInfoRefsUploadPackRequest(ctx, t, serverSocketPath, rpcRequest)
		require.NoError(t, err)
		assertFunc(response)
	})
	t.Run("parsed", func(t *testing.T) {
		caps, refs, err := makeParsedInfoRefsUploadPackRequest(ctx, t, serverSocketPath, requestToParsedRequest(rpcRequest))
		require.NoError(t, err)
		assertFunc(unparse(t, caps, refs))
	})
}

func TestSuccessfulInfoRefsUploadWithPartialClone(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	ctx, cancel := testhelper.Context()
	defer cancel()

	testRepo := testhelper.TestRepository()

	request := &gitalypb.InfoRefsRequest{
		Repository: testRepo,
	}

	assertFunc := func(caps []string) {
		for _, c := range []string{"allow-tip-sha1-in-want", "allow-reachable-sha1-in-want", "filter"} {
			require.Contains(t, caps, c)
		}
	}
	t.Run("raw", func(t *testing.T) {
		partialResponse, err := makeInfoRefsUploadPackRequest(ctx, t, serverSocketPath, request)
		require.NoError(t, err)
		partialRefs := stats.Get{}
		err = partialRefs.Parse(bytes.NewReader(partialResponse))
		require.NoError(t, err)
		assertFunc(partialRefs.Caps)
	})
	t.Run("parsed", func(t *testing.T) {
		caps, _, err := makeParsedInfoRefsUploadPackRequest(ctx, t, serverSocketPath, requestToParsedRequest(request))
		require.NoError(t, err)
		assertFunc(caps)
	})
}

func TestSuccessfulInfoRefsUploadPackWithGitConfigOptions(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	// transfer.hideRefs=refs will hide every ref that info-refs would normally
	// output, allowing us to test that the custom configuration is respected
	rpcRequest := &gitalypb.InfoRefsRequest{
		Repository:       testRepo,
		GitConfigOptions: []string{"transfer.hideRefs=refs"},
	}

	ctx, cancel := testhelper.Context()
	defer cancel()
	assertFunc := func(response []byte) {
		assertGitRefAdvertisement(t, "InfoRefsUploadPack", string(response), "001e# service=git-upload-pack", "0000", []string{})
	}
	t.Run("raw", func(t *testing.T) {
		response, err := makeInfoRefsUploadPackRequest(ctx, t, serverSocketPath, rpcRequest)
		require.NoError(t, err)
		assertFunc(response)
	})
	t.Run("parsed", func(t *testing.T) {
		caps, refs, err := makeParsedInfoRefsUploadPackRequest(ctx, t, serverSocketPath, requestToParsedRequest(rpcRequest))
		require.NoError(t, err)
		assertFunc(unparse(t, caps, refs))
	})
}

func TestSuccessfulInfoRefsUploadPackWithGitProtocol(t *testing.T) {
	restore := testhelper.EnableGitProtocolV2Support(t)
	defer restore()

	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	rpcRequest := &gitalypb.InfoRefsRequest{
		Repository:  testRepo,
		GitProtocol: git.ProtocolV2,
	}

	client, _ := newSmartHTTPClient(t, serverSocketPath)
	ctx, cancel := testhelper.Context()
	defer cancel()

	c, err := client.InfoRefsUploadPack(ctx, rpcRequest)
	require.NoError(t, err)

	for {
		_, err := c.Recv()
		if err != nil {
			require.Equal(t, io.EOF, err)
			break
		}
	}

	envData, err := testhelper.GetGitEnvData()

	require.NoError(t, err)
	require.Contains(t, envData, fmt.Sprintf("GIT_PROTOCOL=%s\n", git.ProtocolV2))
}

func makeInfoRefsUploadPackRequest(ctx context.Context, t *testing.T, serverSocketPath string, rpcRequest *gitalypb.InfoRefsRequest) ([]byte, error) {
	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, err := client.InfoRefsUploadPack(ctx, rpcRequest)
	require.NoError(t, err)

	response, err := ioutil.ReadAll(streamio.NewReader(func() ([]byte, error) {
		resp, err := c.Recv()
		return resp.GetData(), err
	}))

	return response, err
}

func makeParsedInfoRefsUploadPackRequest(ctx context.Context, t *testing.T, serverSocketPath string, rpcRequest *gitalypb.ParsedInfoRefsRequest) ([]string, []*gitalypb.Ref, error) {
	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	c, err := client.ParsedInfoRefsUploadPack(ctx, rpcRequest)
	if err != nil {
		return nil, nil, err
	}

	var caps []string
	var refs []*gitalypb.Ref
	firstMsg := true

	for {
		resp, err := c.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		if firstMsg {
			caps = resp.Capabilities
			firstMsg = false
		} else {
			assert.Empty(t, resp.Capabilities)
		}
		refs = append(refs, resp.Refs...)
	}

	return caps, refs, err
}

func requestToParsedRequest(req *gitalypb.InfoRefsRequest) *gitalypb.ParsedInfoRefsRequest {
	return &gitalypb.ParsedInfoRefsRequest{
		Repository:       req.Repository,
		GitConfigOptions: req.GitConfigOptions,
		GitProtocol:      req.GitProtocol,
	}
}

func unparse(t *testing.T, caps []string, refs []*gitalypb.Ref) []byte {
	var buf bytes.Buffer

	_, err := pktline.WriteString(&buf, "# service=git-upload-pack\n")
	require.NoError(t, err)

	err = pktline.WriteFlush(&buf)
	require.NoError(t, err)

	if len(refs) == 0 {
		refs = []*gitalypb.Ref{
			{
				Name:     []byte("capabilities^{}"),
				CommitId: zeroID,
			},
		}
	}
	first := true
	for _, ref := range refs {
		var s string
		if first {
			s = fmt.Sprintf("%s %s\x00%s\n", ref.CommitId, ref.Name, strings.Join(caps, " "))
			first = false
		} else {
			s = fmt.Sprintf("%s %s\n", ref.CommitId, ref.Name)
		}
		_, err = pktline.WriteString(&buf, s)
		require.NoError(t, err)
	}
	err = pktline.WriteFlush(&buf)
	require.NoError(t, err)

	return buf.Bytes()
}

func TestSuccessfulInfoRefsReceivePack(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	rpcRequest := &gitalypb.InfoRefsRequest{Repository: testRepo}

	ctx, cancel := testhelper.Context()
	defer cancel()
	c, err := client.InfoRefsReceivePack(ctx, rpcRequest)
	if err != nil {
		t.Fatal(err)
	}

	response, err := ioutil.ReadAll(streamio.NewReader(func() ([]byte, error) {
		resp, err := c.Recv()
		return resp.GetData(), err
	}))
	if err != nil {
		t.Fatal(err)
	}

	assertGitRefAdvertisement(t, "InfoRefsReceivePack", string(response), "001f# service=git-receive-pack", "0000", []string{
		"003ef4e6814c3e4e7a0de82a9e7cd20c626cc963a2f8 refs/tags/v1.0.0",
		"003e8a2a6eb295bb170b34c24c76c49ed0e9b2eaf34b refs/tags/v1.1.0",
	})
}

func TestObjectPoolRefAdvertisementHiding(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()

	repo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	ctx, cancel := testhelper.Context()
	defer cancel()

	pool, err := objectpool.NewObjectPool(config.NewLocator(config.Config), repo.GetStorageName(), testhelper.NewTestObjectPoolName(t))
	require.NoError(t, err)

	require.NoError(t, pool.Create(ctx, repo))
	defer pool.Remove(ctx)

	commitID := testhelper.CreateCommit(t, pool.FullPath(), t.Name(), nil)

	require.NoError(t, pool.Link(ctx, repo))

	rpcRequest := &gitalypb.InfoRefsRequest{Repository: repo}

	c, err := client.InfoRefsReceivePack(ctx, rpcRequest)
	require.NoError(t, err)

	response, err := ioutil.ReadAll(streamio.NewReader(func() ([]byte, error) {
		resp, err := c.Recv()
		return resp.GetData(), err
	}))

	require.NoError(t, err)
	require.NotContains(t, string(response), commitID+" .have")
}

func TestFailureRepoNotFoundInfoRefsReceivePack(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()
	repo := &gitalypb.Repository{StorageName: "default", RelativePath: "testdata/scratch/another_repo"}
	rpcRequest := &gitalypb.InfoRefsRequest{Repository: repo}

	ctx, cancel := testhelper.Context()
	defer cancel()
	c, err := client.InfoRefsReceivePack(ctx, rpcRequest)
	if err != nil {
		t.Fatal(err)
	}

	for err == nil {
		_, err = c.Recv()
	}
	testhelper.RequireGrpcError(t, err, codes.NotFound)
}

func TestFailureRepoNotSetInfoRefsReceivePack(t *testing.T) {
	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	client, conn := newSmartHTTPClient(t, serverSocketPath)
	defer conn.Close()
	rpcRequest := &gitalypb.InfoRefsRequest{}

	ctx, cancel := testhelper.Context()
	defer cancel()
	c, err := client.InfoRefsReceivePack(ctx, rpcRequest)
	if err != nil {
		t.Fatal(err)
	}

	for err == nil {
		_, err = c.Recv()
	}
	testhelper.RequireGrpcError(t, err, codes.InvalidArgument)
}

func assertGitRefAdvertisement(t *testing.T, rpc, responseBody string, firstLine, lastLine string, middleLines []string) {
	responseLines := strings.Split(responseBody, "\n")

	if responseLines[0] != firstLine {
		t.Errorf("%q: expected response first line to be %q, found %q", rpc, firstLine, responseLines[0])
	}

	lastIndex := len(responseLines) - 1
	if responseLines[lastIndex] != lastLine {
		t.Errorf("%q: expected response last line to be %q, found %q", rpc, lastLine, responseLines[lastIndex])
	}

	for _, ref := range middleLines {
		if !strings.Contains(responseBody, ref) {
			t.Errorf("%q: expected response to contain %q, found none", rpc, ref)
		}
	}
}

type mockStreamer struct {
	streamer
	putStream func(context.Context, *gitalypb.Repository, proto.Message, io.Reader) error
}

func (ms mockStreamer) PutStream(ctx context.Context, repo *gitalypb.Repository, req proto.Message, src io.Reader) error {
	if ms.putStream != nil {
		return ms.putStream(ctx, repo, req, src)
	}
	return ms.streamer.PutStream(ctx, repo, req, src)
}

func TestCacheInfoRefsUploadPack(t *testing.T) {
	t.Run("raw", func(t *testing.T) {
		testCacheInfoRefsUploadPack(t, "/gitaly.SmartHTTPService/InfoRefsUploadPack", makeInfoRefsUploadPackRequest)
	})
	t.Run("parsed", func(t *testing.T) {
		testCacheInfoRefsUploadPack(t, "/gitaly.SmartHTTPService/ParsedInfoRefsUploadPack", func(ctx context.Context, t *testing.T, serverSocketPath string, rpcRequest *gitalypb.InfoRefsRequest) ([]byte, error) {
			caps, refs, err := makeParsedInfoRefsUploadPackRequest(ctx, t, serverSocketPath, requestToParsedRequest(rpcRequest))
			if err != nil {
				return nil, err
			}
			return unparse(t, caps, refs), nil
		})
	})
}

func testCacheInfoRefsUploadPack(t *testing.T, method string, rawInfoRefsUploadPackRequest func(ctx context.Context, t *testing.T, serverSocketPath string, rpcRequest *gitalypb.InfoRefsRequest) ([]byte, error)) {
	clearCache(t)

	serverSocketPath, stop := runSmartHTTPServer(t)
	defer stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	rpcRequest := &gitalypb.InfoRefsRequest{Repository: testRepo}

	ctx, cancel := testhelper.Context()
	defer cancel()

	assertNormalResponse := func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		response, err := rawInfoRefsUploadPackRequest(ctx, t, serverSocketPath, rpcRequest)
		require.NoError(t, err)

		assertGitRefAdvertisement(t, "InfoRefsUploadPack", string(response),
			"001e# service=git-upload-pack", "0000",
			[]string{
				"003ef4e6814c3e4e7a0de82a9e7cd20c626cc963a2f8 refs/tags/v1.0.0",
				"00416f6d7e7ed97bb5f0054f2b1df789b39ca89b6ff9 refs/tags/v1.0.0^{}",
			},
		)
	}

	assertNormalResponse()
	require.FileExists(t, pathToCachedResponse(t, ctx, rpcRequest, method))

	replacedContents := []string{
		"001e# service=git-upload-pack",
		`00000148` + zeroID + ` HEAD` + "\x00" + `multi_ack thin-pack side-band side-band-64k ofs-delta shallow deepen-since deepen-not deepen-relative no-progress include-tag multi_ack_detailed allow-tip-sha1-in-want allow-reachable-sha1-in-want no-done symref=HEAD:refs/heads/master filter object-format=sha1 agent=git/2.28.0`,
		`003f` + zeroID + ` refs/heads/master`,
		"0000",
	}

	// replace cached response file to prove the info-ref uses the cache
	replaceCachedResponse(t, ctx, rpcRequest, method, strings.Join(replacedContents, "\n"))
	response, err := rawInfoRefsUploadPackRequest(ctx, t, serverSocketPath, rpcRequest)
	require.NoError(t, err)
	assertGitRefAdvertisement(t, "InfoRefsUploadPack", string(response),
		replacedContents[0], replacedContents[3], replacedContents[1:3],
	)

	invalidateCacheForRepo := func() {
		ender, err := cache.LeaseKeyer{}.StartLease(rpcRequest.Repository)
		require.NoError(t, err)
		require.NoError(t, ender.EndLease(setInfoRefsUploadPackMethod(ctx, method)))
	}

	invalidateCacheForRepo()

	// replaced cache response is no longer valid
	assertNormalResponse()

	// failed requests should not cache response
	invalidReq := &gitalypb.InfoRefsRequest{
		Repository: &gitalypb.Repository{
			RelativePath: "fake_repo",
			StorageName:  testRepo.StorageName,
		},
	} // invalid request because repo is empty
	invalidRepoCleanup := createInvalidRepo(t, invalidReq.Repository)
	defer invalidRepoCleanup()

	_, err = rawInfoRefsUploadPackRequest(ctx, t, serverSocketPath, invalidReq)
	testhelper.RequireGrpcError(t, err, codes.Internal)
	testhelper.AssertPathNotExists(t, pathToCachedResponse(t, ctx, invalidReq, method))

	// if an error occurs while putting stream, it should not interrupt
	// request from being served
	happened := false
	defer func(old streamer) { infoRefCache = old }(infoRefCache)
	infoRefCache = mockStreamer{
		streamer: infoRefCache,
		putStream: func(context.Context, *gitalypb.Repository, proto.Message, io.Reader) error {
			happened = true
			return errors.New("oh nos!")
		},
	}
	invalidateCacheForRepo()
	assertNormalResponse()
	require.True(t, happened)
}

func createInvalidRepo(t testing.TB, repo *gitalypb.Repository) func() {
	repoDir, err := helper.GetPath(repo)
	require.NoError(t, err)
	for _, subDir := range []string{"objects", "refs", "HEAD"} {
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, subDir), 0755))
	}
	return func() { require.NoError(t, os.RemoveAll(repoDir)) }
}

func replaceCachedResponse(t testing.TB, ctx context.Context, req *gitalypb.InfoRefsRequest, method, newContents string) {
	path := pathToCachedResponse(t, ctx, req, method)
	require.NoError(t, ioutil.WriteFile(path, []byte(newContents), 0644))
}

func clearCache(t testing.TB) {
	for _, storage := range config.Config.Storages {
		require.NoError(t, os.RemoveAll(tempdir.CacheDir(storage)))
	}
}

func setInfoRefsUploadPackMethod(ctx context.Context, method string) context.Context {
	return testhelper.SetCtxGrpcMethod(ctx, method)
}

func pathToCachedResponse(t testing.TB, ctx context.Context, req *gitalypb.InfoRefsRequest, method string) string {
	ctx = setInfoRefsUploadPackMethod(ctx, method)
	path, err := cache.LeaseKeyer{}.KeyPath(ctx, req.GetRepository(), req)
	require.NoError(t, err)
	return path
}
