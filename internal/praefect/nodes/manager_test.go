package nodes

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/client"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper/promtest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type nodeAssertion struct {
	Storage string
	Address string
}

type shardAssertion struct {
	Primary     *nodeAssertion
	Secondaries []nodeAssertion
}

func toNodeAssertion(n Node) *nodeAssertion {
	if n == nil {
		return nil
	}

	return &nodeAssertion{
		Storage: n.GetStorage(),
		Address: n.GetAddress(),
	}
}

func assertShard(t *testing.T, exp shardAssertion, act Shard) {
	t.Helper()

	actSecondaries := make([]nodeAssertion, 0, len(act.Secondaries))
	for _, n := range act.Secondaries {
		actSecondaries = append(actSecondaries, *toNodeAssertion(n))
	}

	require.Equal(t, exp, shardAssertion{
		Primary:     toNodeAssertion(act.Primary),
		Secondaries: actSecondaries,
	})
}

func TestNodeStatus(t *testing.T) {
	socket := testhelper.GetTemporaryGitalySocketFileName()
	svr, healthSvr := testhelper.NewServerWithHealth(t, socket)
	defer svr.Stop()

	cc, err := grpc.Dial(
		"unix://"+socket,
		grpc.WithInsecure(),
	)

	require.NoError(t, err)

	mockHistogramVec := promtest.NewMockHistogramVec()

	storageName := "default"
	cs := newConnectionStatus(config.Node{Storage: storageName}, cc, testhelper.DiscardTestEntry(t), mockHistogramVec, nil)

	ctx, cancel := testhelper.Context()
	defer cancel()

	var expectedLabels [][]string
	for i := 0; i < healthcheckThreshold; i++ {
		status, err := cs.CheckHealth(ctx)

		require.NoError(t, err)
		require.True(t, status)
		expectedLabels = append(expectedLabels, []string{storageName})
	}

	require.Equal(t, expectedLabels, mockHistogramVec.LabelsCalled())
	require.Len(t, mockHistogramVec.Observer().Observed(), healthcheckThreshold)

	healthSvr.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	status, err := cs.CheckHealth(ctx)
	require.NoError(t, err)
	require.False(t, status)
}

func TestManagerFailoverDisabledElectionStrategySQL(t *testing.T) {
	const virtualStorageName = "virtual-storage-0"
	const primaryStorage = "praefect-internal-0"
	socket0, socket1 := testhelper.GetTemporaryGitalySocketFileName(), testhelper.GetTemporaryGitalySocketFileName()
	virtualStorage := &config.VirtualStorage{
		Name: virtualStorageName,
		Nodes: []*config.Node{
			{
				Storage: primaryStorage,
				Address: "unix://" + socket0,
			},
			{
				Storage: "praefect-internal-1",
				Address: "unix://" + socket1,
			},
		},
	}

	srv0, healthSrv := testhelper.NewServerWithHealth(t, socket0)
	defer srv0.Stop()

	srv1, _ := testhelper.NewServerWithHealth(t, socket1)
	defer srv1.Stop()

	conf := config.Config{
		Failover:        config.Failover{Enabled: false, ElectionStrategy: "sql"},
		VirtualStorages: []*config.VirtualStorage{virtualStorage},
	}
	nm, err := NewManager(testhelper.DiscardTestEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)

	nm.Start(time.Millisecond, time.Millisecond)

	shard, err := nm.GetShard(virtualStorageName)
	require.NoError(t, err)
	require.Equal(t, primaryStorage, shard.Primary.GetStorage())

	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
	nm.checkShards()

	shard, err = nm.GetShard(virtualStorageName)
	require.Error(t, err)
	require.Equal(t, ErrPrimaryNotHealthy, err)
}

func TestDialWithUnhealthyNode(t *testing.T) {
	primaryLn, err := net.Listen("unix", testhelper.GetTemporaryGitalySocketFileName())
	require.NoError(t, err)

	primaryAddress := "unix://" + primaryLn.Addr().String()
	const secondaryAddress = "unix://does-not-exist"
	const storageName = "default"

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: storageName,
				Nodes: []*config.Node{
					{
						Storage: "starts",
						Address: primaryAddress,
					},
					{
						Storage: "never-starts",
						Address: secondaryAddress,
					},
				},
			},
		},
	}

	srv, _ := testhelper.NewHealthServerWithListener(t, primaryLn)
	defer srv.Stop()

	mgr, err := NewManager(testhelper.DiscardTestEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)

	mgr.Start(1*time.Millisecond, 1*time.Millisecond)

	shard, err := mgr.GetShard(storageName)
	require.NoError(t, err)
	assertShard(t, shardAssertion{
		Primary:     &nodeAssertion{Storage: "starts", Address: primaryAddress},
		Secondaries: []nodeAssertion{{Storage: "never-starts", Address: secondaryAddress}},
	}, shard)
}

func TestGetPrimaries(t *testing.T) {
	primarySocket := testhelper.GetTemporaryGitalySocketFileName()

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "healthy-primary",
				Nodes: []*config.Node{
					{
						Storage: "primary",
						Address: "unix://" + primarySocket,
					},
				},
			},
			{
				Name: "unhealthy-primary",
				Nodes: []*config.Node{
					{
						Storage: "primary",
						Address: "unix://does-not-exist",
					},
				},
			},
		},
		Failover: config.Failover{Enabled: true},
	}

	primaryListener, err := net.Listen("unix", primarySocket)
	require.NoError(t, err)

	primarySrv, _ := testhelper.NewHealthServerWithListener(t, primaryListener)
	defer primarySrv.Stop()

	mgr, err := NewManager(testhelper.DiscardTestEntry(t), conf, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)

	for i := 0; i < healthcheckThreshold; i++ {
		mgr.checkShards()
	}

	shard, err := mgr.GetShard("healthy-primary")
	require.NoError(t, err)
	assertShard(t, shardAssertion{
		Primary:     &nodeAssertion{Storage: "primary", Address: "unix://" + primaryListener.Addr().String()},
		Secondaries: []nodeAssertion{},
	}, shard)

	_, err = mgr.GetShard("unhealthy-primary")
	require.Equal(t, ErrPrimaryNotHealthy, err)

	primaries, err := mgr.GetPrimaries(context.Background())
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"healthy-primary":   "primary",
		"unhealthy-primary": "",
	}, primaries)
}

func TestNodeManager(t *testing.T) {
	internalSocket0, internalSocket1 := testhelper.GetTemporaryGitalySocketFileName(), testhelper.GetTemporaryGitalySocketFileName()
	srv0, healthSrv0 := testhelper.NewServerWithHealth(t, internalSocket0)
	defer srv0.Stop()

	srv1, healthSrv1 := testhelper.NewServerWithHealth(t, internalSocket1)
	defer srv1.Stop()

	node1 := &config.Node{
		Storage: "praefect-internal-0",
		Address: "unix://" + internalSocket0,
	}

	node2 := &config.Node{
		Storage: "praefect-internal-1",
		Address: "unix://" + internalSocket1,
	}

	virtualStorages := []*config.VirtualStorage{
		{
			Name:  "virtual-storage-0",
			Nodes: []*config.Node{node1, node2},
		},
	}

	confWithFailover := config.Config{
		VirtualStorages: virtualStorages,
		Failover:        config.Failover{Enabled: true},
	}
	confWithoutFailover := config.Config{
		VirtualStorages: virtualStorages,
		Failover:        config.Failover{Enabled: false},
	}

	mockHistogram := promtest.NewMockHistogramVec()
	nm, err := NewManager(testhelper.DiscardTestEntry(t), confWithFailover, nil, nil, mockHistogram, protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)

	nmWithoutFailover, err := NewManager(testhelper.DiscardTestEntry(t), confWithoutFailover, nil, nil, mockHistogram, protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)

	// monitoring period set to 1 hour as we execute health checks by hands in this test
	nm.Start(0, time.Hour)
	nmWithoutFailover.Start(0, time.Hour)

	shardWithoutFailover, err := nmWithoutFailover.GetShard("virtual-storage-0")
	require.NoError(t, err)

	shard, err := nm.GetShard("virtual-storage-0")
	require.NoError(t, err)

	// shard without failover and shard with failover should be the same
	initialState := shardAssertion{
		Primary:     &nodeAssertion{node1.Storage, node1.Address},
		Secondaries: []nodeAssertion{{node2.Storage, node2.Address}},
	}
	assertShard(t, initialState, shard)
	assertShard(t, initialState, shardWithoutFailover)

	const unhealthyCheckCount = 1
	const healthyCheckCount = healthcheckThreshold
	checkShards := func(count int) {
		for i := 0; i < count; i++ {
			nm.checkShards()
			nmWithoutFailover.checkShards()
		}
	}

	healthSrv0.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	checkShards(unhealthyCheckCount)

	labelsCalled := mockHistogram.LabelsCalled()
	for _, node := range virtualStorages[0].Nodes {
		require.Contains(t, labelsCalled, []string{node.Storage})
	}

	// since the failover is disabled the attempt to get a shard with unhealthy primary fails
	_, err = nmWithoutFailover.GetShard("virtual-storage-0")
	require.Error(t, err)
	require.Equal(t, ErrPrimaryNotHealthy, err)

	// since the primary is unhealthy, we expect checkShards to demote primary to secondary, and promote the healthy
	// secondary to primary
	shard, err = nm.GetShard("virtual-storage-0")
	require.NoError(t, err)
	// shard with failover should have promoted a secondary to primary and demoted the primary to a secondary
	assertShard(t, shardAssertion{
		Primary:     &nodeAssertion{node2.Storage, node2.Address},
		Secondaries: []nodeAssertion{{node1.Storage, node1.Address}},
	}, shard)

	// failing back to the original primary
	healthSrv0.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthSrv1.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	checkShards(healthyCheckCount)

	shard, err = nm.GetShard("virtual-storage-0")
	require.NoError(t, err)

	assertShard(t, shardAssertion{
		Primary:     &nodeAssertion{node1.Storage, node1.Address},
		Secondaries: []nodeAssertion{{node2.Storage, node2.Address}},
	}, shard)

	healthSrv0.SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
	healthSrv1.SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
	checkShards(unhealthyCheckCount)

	_, err = nm.GetShard("virtual-storage-0")
	require.Error(t, err, "should return error since no nodes are healthy")
}

func TestMgr_GetSyncedNode(t *testing.T) {
	const count = 3
	const virtualStorage = "virtual-storage-0"
	const repoPath = "path/1"

	var srvs [count]*grpc.Server
	var healthSrvs [count]*health.Server
	var nodes [count]*config.Node
	for i := 0; i < count; i++ {
		socket := testhelper.GetTemporaryGitalySocketFileName()
		srvs[i], healthSrvs[i] = testhelper.NewServerWithHealth(t, socket)
		defer srvs[i].Stop()
		nodes[i] = &config.Node{Storage: fmt.Sprintf("gitaly-%d", i), Address: "unix://" + socket}
	}

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{{Name: virtualStorage, Nodes: nodes[:]}},
	}

	ctx, cancel := testhelper.Context()
	defer cancel()

	verify := func(failover bool, scenario func(t *testing.T, nm Manager, rs datastore.RepositoryStore)) func(*testing.T) {
		conf.Failover.Enabled = failover
		rs := datastore.NewMemoryRepositoryStore(conf.StorageNames())

		nm, err := NewManager(testhelper.DiscardTestEntry(t), conf, nil, rs, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
		require.NoError(t, err)

		for i := range healthSrvs {
			healthSrvs[i].SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		}
		nm.Start(0, time.Hour)

		return func(t *testing.T) { scenario(t, nm, rs) }
	}

	t.Run("unknown virtual storage", verify(true, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		_, err := nm.GetSyncedNode(ctx, "virtual-storage-unknown", "stub")
		require.True(t, errors.Is(err, ErrVirtualStorageNotExist))
	}))

	t.Run("state is undefined", verify(true, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		node, err := nm.GetSyncedNode(ctx, virtualStorage, "no/matter")
		require.NoError(t, err)
		require.Equal(t, conf.VirtualStorages[0].Nodes[0].Address, node.GetAddress(), "")
	}))

	t.Run("multiple storages up to date", verify(true, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		require.NoError(t, rs.IncrementGeneration(ctx, virtualStorage, repoPath, "gitaly-0", nil))
		generation, err := rs.GetGeneration(ctx, virtualStorage, repoPath, "gitaly-0")
		require.NoError(t, err)
		require.NoError(t, rs.SetGeneration(ctx, virtualStorage, repoPath, "gitaly-1", generation))
		require.NoError(t, rs.SetGeneration(ctx, virtualStorage, repoPath, "gitaly-2", generation))

		chosen := map[Node]struct{}{}
		for i := 0; i < 1000 && len(chosen) < 2; i++ {
			node, err := nm.GetSyncedNode(ctx, virtualStorage, repoPath)
			require.NoError(t, err)
			chosen[node] = struct{}{}
		}
		if len(chosen) < 2 {
			require.FailNow(t, "no distribution in too many attempts")
		}
	}))

	t.Run("single secondary storage up to date but unhealthy", verify(true, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		require.NoError(t, rs.IncrementGeneration(ctx, virtualStorage, repoPath, "gitaly-0", nil))
		generation, err := rs.GetGeneration(ctx, virtualStorage, repoPath, "gitaly-0")
		require.NoError(t, err)
		require.NoError(t, rs.SetGeneration(ctx, virtualStorage, repoPath, "gitaly-1", generation))

		healthSrvs[1].SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)

		shard, err := nm.GetShard(virtualStorage)
		require.NoError(t, err)

		gitaly1, err := shard.GetNode("gitaly-1")
		require.NoError(t, err)

		ok, err := gitaly1.CheckHealth(ctx)
		require.NoError(t, err)
		require.False(t, ok)

		node, err := nm.GetSyncedNode(ctx, virtualStorage, repoPath)
		require.NoError(t, err)
		require.Equal(t, conf.VirtualStorages[0].Nodes[0].Address, node.GetAddress(), "secondary shouldn't be chosen as it is unhealthy")
	}))

	t.Run("no healthy storages", verify(true, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		require.NoError(t, rs.IncrementGeneration(ctx, virtualStorage, repoPath, "gitaly-0", nil))
		generation, err := rs.GetGeneration(ctx, virtualStorage, repoPath, "gitaly-0")
		require.NoError(t, err)
		require.NoError(t, rs.SetGeneration(ctx, virtualStorage, repoPath, "gitaly-1", generation))

		healthSrvs[0].SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
		healthSrvs[1].SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)

		shard, err := nm.GetShard(virtualStorage)
		require.NoError(t, err)

		gitaly0, err := shard.GetNode("gitaly-0")
		require.NoError(t, err)

		gitaly0OK, err := gitaly0.CheckHealth(ctx)
		require.NoError(t, err)
		require.False(t, gitaly0OK)

		gitaly1, err := shard.GetNode("gitaly-1")
		require.NoError(t, err)

		gitaly1OK, err := gitaly1.CheckHealth(ctx)
		require.NoError(t, err)
		require.False(t, gitaly1OK)

		_, err = nm.GetSyncedNode(ctx, virtualStorage, repoPath)
		require.True(t, errors.Is(err, ErrPrimaryNotHealthy))
	}))

	t.Run("disabled failover doesn't disable health state", verify(false, func(t *testing.T, nm Manager, rs datastore.RepositoryStore) {
		require.NoError(t, rs.IncrementGeneration(ctx, virtualStorage, repoPath, "gitaly-0", nil))
		generation, err := rs.GetGeneration(ctx, virtualStorage, repoPath, "gitaly-0")
		require.NoError(t, err)
		require.NoError(t, rs.SetGeneration(ctx, virtualStorage, repoPath, "gitaly-1", generation))

		shard, err := nm.GetShard(virtualStorage)
		require.NoError(t, err)

		gitaly0, err := shard.GetNode("gitaly-0")
		require.NoError(t, err)

		require.Equal(t, shard.Primary, gitaly0)

		gitaly0OK, err := gitaly0.CheckHealth(ctx)
		require.NoError(t, err)
		require.True(t, gitaly0OK)

		gitaly1, err := shard.GetNode("gitaly-1")
		require.NoError(t, err)

		gitaly1OK, err := gitaly1.CheckHealth(ctx)
		require.NoError(t, err)
		require.True(t, gitaly1OK)

		healthSrvs[0].SetServingStatus("", grpc_health_v1.HealthCheckResponse_UNKNOWN)
		gitaly0OK, err = gitaly0.CheckHealth(ctx)
		require.NoError(t, err)
		require.False(t, gitaly0OK, "primary should be unhealthy")

		_, err = nm.GetSyncedNode(ctx, virtualStorage, repoPath)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrPrimaryNotHealthy))
	}))
}

func TestNodeStatus_IsHealthy(t *testing.T) {
	checkNTimes := func(ctx context.Context, t *testing.T, ns *nodeStatus, n int) {
		for i := 0; i < n; i++ {
			_, err := ns.CheckHealth(ctx)
			require.NoError(t, err)
		}
	}

	socket := testhelper.GetTemporaryGitalySocketFileName()
	address := "unix://" + socket

	srv, healthSrv := testhelper.NewServerWithHealth(t, socket)
	defer srv.Stop()

	clientConn, err := client.Dial(address, nil)
	require.NoError(t, err)
	defer func() { require.NoError(t, clientConn.Close()) }()

	node := config.Node{Storage: "gitaly-0", Address: address}

	ctx, cancel := testhelper.Context()
	defer cancel()

	logger := testhelper.DiscardTestLogger(t)
	latencyHistMock := &promtest.MockHistogramVec{}

	t.Run("unchecked node is unhealthy", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		require.False(t, ns.IsHealthy())
	})

	t.Run("not enough check to consider it healthy", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		checkNTimes(ctx, t, ns, healthcheckThreshold-1)

		require.False(t, ns.IsHealthy())
	})

	t.Run("healthy", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		checkNTimes(ctx, t, ns, healthcheckThreshold)

		require.True(t, ns.IsHealthy())
	})

	t.Run("healthy turns into unhealthy after single failed check", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		checkNTimes(ctx, t, ns, healthcheckThreshold)

		require.True(t, ns.IsHealthy(), "node must be turned into healthy state")

		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		checkNTimes(ctx, t, ns, 1)

		require.False(t, ns.IsHealthy(), "node must be turned into unhealthy state")
	})

	t.Run("unhealthy turns into healthy after pre-define threshold of checks", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		checkNTimes(ctx, t, ns, healthcheckThreshold)

		require.True(t, ns.IsHealthy(), "node must be turned into healthy state")

		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		checkNTimes(ctx, t, ns, 1)

		require.False(t, ns.IsHealthy(), "node must be turned into unhealthy state")

		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		for i := 1; i < healthcheckThreshold; i++ {
			checkNTimes(ctx, t, ns, 1)
			require.False(t, ns.IsHealthy(), "node must be unhealthy until defined threshold of checks complete positively")
		}
		checkNTimes(ctx, t, ns, 1) // the last check that must turn it into healthy state

		require.True(t, ns.IsHealthy(), "node should be healthy again")
	})

	t.Run("concurrent access has no races", func(t *testing.T) {
		ns := newConnectionStatus(node, clientConn, logger, latencyHistMock, nil)
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

		t.Run("continuously does health checks - 1", func(t *testing.T) {
			t.Parallel()
			checkNTimes(ctx, t, ns, healthcheckThreshold)
		})

		t.Run("continuously checks health - 1", func(t *testing.T) {
			t.Parallel()
			ns.IsHealthy()
		})

		t.Run("continuously does health checks - 2", func(t *testing.T) {
			t.Parallel()
			checkNTimes(ctx, t, ns, healthcheckThreshold)
		})

		t.Run("continuously checks health - 2", func(t *testing.T) {
			t.Parallel()
			ns.IsHealthy()
		})
	})
}
