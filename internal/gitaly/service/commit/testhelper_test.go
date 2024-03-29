package commit

import (
	"net"
	"os"
	"testing"

	"github.com/golang/protobuf/ptypes/timestamp"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestMain(m *testing.M) {
	testhelper.Configure()
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	return m.Run()
}

func startTestServices(t testing.TB) (*grpc.Server, string) {
	server := testhelper.NewTestGrpcServer(t, nil, nil)
	serverSocketPath := testhelper.GetTemporaryGitalySocketFileName()

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal("failed to start server")
	}

	gitalypb.RegisterCommitServiceServer(server, NewServer(config.NewLocator(config.Config)))
	reflection.Register(server)

	go server.Serve(listener)
	return server, "unix://" + serverSocketPath
}

func newCommitServiceClient(t testing.TB, serviceSocketPath string) (gitalypb.CommitServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(serviceSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return gitalypb.NewCommitServiceClient(conn), conn
}

func dummyCommitAuthor(ts int64) *gitalypb.CommitAuthor {
	return &gitalypb.CommitAuthor{
		Name:     []byte("Ahmad Sherif"),
		Email:    []byte("ahmad+gitlab-test@gitlab.com"),
		Date:     &timestamp.Timestamp{Seconds: ts},
		Timezone: []byte("+0200"),
	}
}
