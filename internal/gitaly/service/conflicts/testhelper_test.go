package conflicts

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/git/hooks"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestMain(m *testing.M) {
	testhelper.Configure()
	os.Exit(testMain(m))
}

var RubyServer = &rubyserver.Server{}

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	testhelper.ConfigureGitalyGit2Go()

	tempDir, err := ioutil.TempDir("", "gitaly")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	hooks.Override = tempDir + "/hooks"
	config.Config.InternalSocketDir = tempDir + "/sock"

	if err := RubyServer.Start(); err != nil {
		log.Fatal(err)
	}
	defer RubyServer.Stop()

	return m.Run()
}

func runConflictsServer(t *testing.T) (string, func()) {
	srv := testhelper.NewServer(t, nil, nil)
	locator := config.NewLocator(config.Config)

	gitalypb.RegisterConflictsServiceServer(srv.GrpcServer(), NewServer(RubyServer, config.Config, locator))
	reflection.Register(srv.GrpcServer())

	require.NoError(t, srv.Start())

	return "unix://" + srv.Socket(), srv.Stop
}

func NewConflictsClient(t *testing.T, serverSocketPath string) (gitalypb.ConflictsServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return gitalypb.NewConflictsServiceClient(conn), conn
}
