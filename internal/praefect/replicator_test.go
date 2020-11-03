package praefect

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitalyauth "gitlab.com/gitlab-org/gitaly/auth"
	"gitlab.com/gitlab-org/gitaly/client"
	"gitlab.com/gitlab-org/gitaly/internal/git/objectpool"
	gitaly_config "gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/hook"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	serverPkg "gitlab.com/gitlab-org/gitaly/internal/gitaly/server"
	objectpoolservice "gitlab.com/gitlab-org/gitaly/internal/gitaly/service/objectpool"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/ref"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/remote"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/repository"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/ssh"
	"gitlab.com/gitlab-org/gitaly/internal/middleware/metadatahandler"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/nodes"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/transactions"
	"gitlab.com/gitlab-org/gitaly/internal/storage"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper/promtest"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/labkit/correlation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func TestReplMgr_ProcessBacklog(t *testing.T) {
	backupStorageName := "backup"

	backupDir, err := ioutil.TempDir(testhelper.GitlabTestStoragePath(), backupStorageName)
	require.NoError(t, err)

	defer func() { os.RemoveAll(backupDir) }()
	defer func(oldStorages []gitaly_config.Storage) { gitaly_config.Config.Storages = oldStorages }(gitaly_config.Config.Storages)

	gitaly_config.Config.Storages = append(gitaly_config.Config.Storages,
		gitaly_config.Storage{
			Name: backupStorageName,
			Path: backupDir,
		},
		gitaly_config.Storage{
			Name: "default",
			Path: testhelper.GitlabTestStoragePath(),
		},
	)

	srvSocketPath, clean := runFullGitalyServer(t)
	defer clean()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			&config.VirtualStorage{
				Name: "default",
				Nodes: []*config.Node{
					&config.Node{
						Storage: "default",
						Address: srvSocketPath,
						Token:   gitaly_config.Config.Auth.Token,
					},
					&config.Node{
						Storage: backupStorageName,
						Address: srvSocketPath,
						Token:   gitaly_config.Config.Auth.Token,
					},
				},
			},
		},
	}

	// create object pool on the source
	objectPoolPath := testhelper.NewTestObjectPoolName(t)
	pool, err := objectpool.NewObjectPool(gitaly_config.NewLocator(gitaly_config.Config), testRepo.GetStorageName(), objectPoolPath)
	require.NoError(t, err)

	poolCtx, cancel := testhelper.Context()
	defer cancel()

	require.NoError(t, pool.Create(poolCtx, testRepo))
	require.NoError(t, pool.Link(poolCtx, testRepo))

	// replicate object pool repository to target node
	targetObjectPoolRepo := *pool.ToProto().GetRepository()
	targetObjectPoolRepo.StorageName = "backup"

	ctx, cancel := testhelper.Context()
	defer cancel()

	injectedCtx := metadata.NewOutgoingContext(ctx, testhelper.GitalyServersMetadata(t, srvSocketPath))

	repoClient, con := newRepositoryClient(t, srvSocketPath)
	defer con.Close()

	_, err = repoClient.ReplicateRepository(injectedCtx, &gitalypb.ReplicateRepositoryRequest{
		Repository: &targetObjectPoolRepo,
		Source:     pool.ToProto().GetRepository(),
	})
	require.NoError(t, err)

	entry := testhelper.DiscardTestEntry(t)

	nodeMgr, err := nodes.NewManager(entry, conf, nil, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)
	nodeMgr.Start(1*time.Millisecond, 5*time.Millisecond)

	shard, err := nodeMgr.GetShard(conf.VirtualStorages[0].Name)
	require.NoError(t, err)
	require.Len(t, shard.Secondaries, 1)

	var events []datastore.ReplicationEvent
	for _, secondary := range shard.Secondaries {
		events = append(events, datastore.ReplicationEvent{
			Job: datastore.ReplicationJob{
				VirtualStorage:    "default",
				Change:            datastore.UpdateRepo,
				TargetNodeStorage: secondary.GetStorage(),
				SourceNodeStorage: shard.Primary.GetStorage(),
				RelativePath:      testRepo.GetRelativePath(),
			},
			State:   datastore.JobStateReady,
			Attempt: 3,
			Meta:    datastore.Params{metadatahandler.CorrelationIDKey: "correlation-id"},
		})
	}
	require.Len(t, events, 1)

	commitID := testhelper.CreateCommit(t, testRepoPath, "master", &testhelper.CreateCommitOpts{
		Message: "a commit",
	})

	var mockReplicationLatencyHistogramVec promtest.MockHistogramVec
	var mockReplicationDelayHistogramVec promtest.MockHistogramVec

	logger := testhelper.DiscardTestLogger(t)
	hook := test.NewLocal(logger)

	queue := datastore.NewReplicationEventQueueInterceptor(datastore.NewMemoryReplicationEventQueue(conf))
	queue.OnAcknowledge(func(ctx context.Context, state datastore.JobState, ids []uint64, queue datastore.ReplicationEventQueue) ([]uint64, error) {
		cancel() // when it is called we know that replication is finished
		return queue.Acknowledge(ctx, state, ids)
	})

	loggerEntry := logger.WithField("test", t.Name())
	_, err = queue.Enqueue(ctx, events[0])
	require.NoError(t, err)

	store := datastore.NewMemoryRepositoryStore(conf.StorageNames())

	replMgr := NewReplMgr(
		loggerEntry,
		conf.VirtualStorageNames(),
		queue,
		store,
		nodeMgr,
		NodeSetFromNodeManager(nodeMgr),
		WithLatencyMetric(&mockReplicationLatencyHistogramVec),
		WithDelayMetric(&mockReplicationDelayHistogramVec),
	)

	replMgr.ProcessBacklog(ctx, ExpBackoffFunc(time.Hour, 0))

	<-ctx.Done()

	logEntries := hook.AllEntries()
	require.True(t, len(logEntries) > 3, "expected at least 4 log entries to be present")
	require.Equal(t,
		[]interface{}{"processing started", "default"},
		[]interface{}{logEntries[0].Message, logEntries[0].Data["virtual_storage"]},
	)

	require.Equal(t,
		[]interface{}{"replication job processing started", "default", "correlation-id"},
		[]interface{}{logEntries[1].Message, logEntries[1].Data["virtual_storage"], logEntries[1].Data[logWithCorrID]},
	)

	dequeuedEvent := logEntries[1].Data["event"].(datastore.ReplicationEvent)
	require.Equal(t, datastore.JobStateInProgress, dequeuedEvent.State)
	require.Equal(t, []string{"backup", "default"}, []string{dequeuedEvent.Job.TargetNodeStorage, dequeuedEvent.Job.SourceNodeStorage})

	require.Equal(t,
		[]interface{}{"checksum comparison completed", "default", "correlation-id"},
		[]interface{}{logEntries[2].Message, logEntries[2].Data["virtual_storage"], logEntries[2].Data[logWithCorrID]},
	)

	require.Equal(t,
		[]interface{}{"replication job processing finished", "default", datastore.JobStateCompleted, "correlation-id"},
		[]interface{}{logEntries[3].Message, logEntries[3].Data["virtual_storage"], logEntries[3].Data["new_state"], logEntries[3].Data[logWithCorrID]},
	)

	relativeRepoPath, err := filepath.Rel(testhelper.GitlabTestStoragePath(), testRepoPath)
	require.NoError(t, err)
	replicatedPath := filepath.Join(backupDir, relativeRepoPath)

	testhelper.MustRunCommand(t, nil, "git", "-C", replicatedPath, "cat-file", "-e", commitID)
	testhelper.MustRunCommand(t, nil, "git", "-C", replicatedPath, "gc")
	require.Less(t, testhelper.GetGitPackfileDirSize(t, replicatedPath), int64(100), "expect a small pack directory")

	require.Equal(t, mockReplicationLatencyHistogramVec.LabelsCalled(), [][]string{{"update"}})
	require.Equal(t, mockReplicationDelayHistogramVec.LabelsCalled(), [][]string{{"update"}})
	require.NoError(t, testutil.CollectAndCompare(replMgr, strings.NewReader(`
# HELP gitaly_praefect_replication_jobs Number of replication jobs in flight.
# TYPE gitaly_praefect_replication_jobs gauge
gitaly_praefect_replication_jobs{change_type="update",gitaly_storage="backup",virtual_storage="default"} 0
`)))
}

func TestReplicatorDowngradeAttempt(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	ctx = correlation.ContextWithCorrelation(ctx, "correlation-id")

	rs := datastore.NewMemoryRepositoryStore(nil)
	require.NoError(t, rs.SetGeneration(ctx, "virtual-storage-1", "relative-path-1", "gitaly-1", 0))
	require.NoError(t, rs.SetGeneration(ctx, "virtual-storage-1", "relative-path-1", "gitaly-2", 0))

	logger := testhelper.DiscardTestLogger(t)
	hook := test.NewLocal(logger)
	r := &defaultReplicator{rs: rs, log: logger}

	require.NoError(t, r.Replicate(ctx, datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			VirtualStorage:    "virtual-storage-1",
			RelativePath:      "relative-path-1",
			SourceNodeStorage: "gitaly-1",
			TargetNodeStorage: "gitaly-2",
		},
	}, nil, nil))

	require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	require.Equal(t, datastore.DowngradeAttemptedError{
		VirtualStorage:      "virtual-storage-1",
		RelativePath:        "relative-path-1",
		Storage:             "gitaly-2",
		CurrentGeneration:   0,
		AttemptedGeneration: 0,
	}, hook.LastEntry().Data["error"])
	require.Equal(t, "correlation-id", hook.LastEntry().Data[logWithCorrID])
	require.Equal(t, "target repository already on the same generation, skipping replication job", hook.LastEntry().Message)

	require.NoError(t, rs.SetGeneration(ctx, "virtual-storage-1", "relative-path-1", "gitaly-2", 1))
	require.NoError(t, r.Replicate(ctx, datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			VirtualStorage:    "virtual-storage-1",
			RelativePath:      "relative-path-1",
			SourceNodeStorage: "gitaly-1",
			TargetNodeStorage: "gitaly-2",
		},
	}, nil, nil))

	require.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	require.Equal(t, datastore.DowngradeAttemptedError{
		VirtualStorage:      "virtual-storage-1",
		RelativePath:        "relative-path-1",
		Storage:             "gitaly-2",
		CurrentGeneration:   1,
		AttemptedGeneration: 0,
	}, hook.LastEntry().Data["error"])
	require.Equal(t, "correlation-id", hook.LastEntry().Data[logWithCorrID])
	require.Equal(t, "repository downgrade prevented", hook.LastEntry().Message)
}

func TestPropagateReplicationJob(t *testing.T) {
	primaryServer, primarySocketPath, cleanup := runMockRepositoryServer(t)
	defer cleanup()

	secondaryServer, secondarySocketPath, cleanup := runMockRepositoryServer(t)
	defer cleanup()

	primaryStorage, secondaryStorage := "internal-gitaly-0", "internal-gitaly-1"
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "default",
				Nodes: []*config.Node{
					{
						Storage: primaryStorage,
						Address: primarySocketPath,
					},
					{
						Storage: secondaryStorage,
						Address: secondarySocketPath,
					},
				},
			},
		},
	}

	queue := datastore.NewMemoryReplicationEventQueue(conf)
	logEntry := testhelper.DiscardTestEntry(t)

	nodeMgr, err := nodes.NewManager(logEntry, conf, nil, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)

	txMgr := transactions.NewManager(conf)

	rs := datastore.NewMemoryRepositoryStore(conf.StorageNames())

	coordinator := NewCoordinator(
		queue,
		rs,
		NewNodeManagerRouter(nodeMgr, rs),
		txMgr,
		conf,
		protoregistry.GitalyProtoPreregistered,
	)

	replmgr := NewReplMgr(logEntry, conf.VirtualStorageNames(), queue, rs, nodeMgr, NodeSetFromNodeManager(nodeMgr))

	prf := NewGRPCServer(conf, logEntry, protoregistry.GitalyProtoPreregistered, coordinator.StreamDirector, nodeMgr, txMgr, queue, rs)

	listener, port := listenAvailPort(t)
	ctx, cancel := testhelper.Context()
	defer cancel()

	go prf.Serve(listener)
	defer prf.Stop()

	cc := dialLocalPort(t, port, false)
	repositoryClient := gitalypb.NewRepositoryServiceClient(cc)
	defer listener.Close()
	defer cc.Close()

	repositoryRelativePath := "/path/to/repo"

	repository := &gitalypb.Repository{
		StorageName:  conf.VirtualStorages[0].Name,
		RelativePath: repositoryRelativePath,
	}

	_, err = repositoryClient.GarbageCollect(ctx, &gitalypb.GarbageCollectRequest{Repository: repository, CreateBitmap: true})
	require.NoError(t, err)

	_, err = repositoryClient.RepackFull(ctx, &gitalypb.RepackFullRequest{Repository: repository, CreateBitmap: false})
	require.NoError(t, err)

	_, err = repositoryClient.RepackIncremental(ctx, &gitalypb.RepackIncrementalRequest{Repository: repository})
	require.NoError(t, err)

	primaryRepository := &gitalypb.Repository{StorageName: primaryStorage, RelativePath: repositoryRelativePath}
	expectedPrimaryGcReq := &gitalypb.GarbageCollectRequest{
		Repository:   primaryRepository,
		CreateBitmap: true,
	}
	expectedPrimaryRepackFullReq := &gitalypb.RepackFullRequest{
		Repository:   primaryRepository,
		CreateBitmap: false,
	}
	expectedPrimaryRepackIncrementalReq := &gitalypb.RepackIncrementalRequest{
		Repository: primaryRepository,
	}

	replCtx, cancel := testhelper.Context()
	defer cancel()
	replmgr.ProcessBacklog(replCtx, noopBackoffFunc)

	// ensure primary gitaly server received the expected requests
	waitForRequest(t, primaryServer.gcChan, expectedPrimaryGcReq, 5*time.Second)
	waitForRequest(t, primaryServer.repackIncrChan, expectedPrimaryRepackIncrementalReq, 5*time.Second)
	waitForRequest(t, primaryServer.repackFullChan, expectedPrimaryRepackFullReq, 5*time.Second)

	secondaryRepository := &gitalypb.Repository{StorageName: secondaryStorage, RelativePath: repositoryRelativePath}

	expectedSecondaryGcReq := expectedPrimaryGcReq
	expectedSecondaryGcReq.Repository = secondaryRepository

	expectedSecondaryRepackFullReq := expectedPrimaryRepackFullReq
	expectedSecondaryRepackFullReq.Repository = secondaryRepository

	expectedSecondaryRepackIncrementalReq := expectedPrimaryRepackIncrementalReq
	expectedSecondaryRepackIncrementalReq.Repository = secondaryRepository

	// ensure secondary gitaly server received the expected requests
	waitForRequest(t, secondaryServer.gcChan, expectedSecondaryGcReq, 5*time.Second)
	waitForRequest(t, secondaryServer.repackIncrChan, expectedSecondaryRepackIncrementalReq, 5*time.Second)
	waitForRequest(t, secondaryServer.repackFullChan, expectedSecondaryRepackFullReq, 5*time.Second)
}

type mockRepositoryServer struct {
	gcChan, repackFullChan, repackIncrChan chan proto.Message

	gitalypb.UnimplementedRepositoryServiceServer
}

func newMockRepositoryServer() *mockRepositoryServer {
	return &mockRepositoryServer{
		gcChan:         make(chan proto.Message),
		repackFullChan: make(chan proto.Message),
		repackIncrChan: make(chan proto.Message),
	}
}

func (m *mockRepositoryServer) GarbageCollect(ctx context.Context, in *gitalypb.GarbageCollectRequest) (*gitalypb.GarbageCollectResponse, error) {
	go func() {
		m.gcChan <- in
	}()
	return &gitalypb.GarbageCollectResponse{}, nil
}

func (m *mockRepositoryServer) RepackFull(ctx context.Context, in *gitalypb.RepackFullRequest) (*gitalypb.RepackFullResponse, error) {
	go func() {
		m.repackFullChan <- in
	}()
	return &gitalypb.RepackFullResponse{}, nil
}

func (m *mockRepositoryServer) RepackIncremental(ctx context.Context, in *gitalypb.RepackIncrementalRequest) (*gitalypb.RepackIncrementalResponse, error) {
	go func() {
		m.repackIncrChan <- in
	}()
	return &gitalypb.RepackIncrementalResponse{}, nil
}

func runMockRepositoryServer(t *testing.T) (*mockRepositoryServer, string, func()) {
	server := testhelper.NewTestGrpcServer(t, nil, nil)
	serverSocketPath := testhelper.GetTemporaryGitalySocketFileName()

	listener, err := net.Listen("unix", serverSocketPath)
	require.NoError(t, err)

	mockServer := newMockRepositoryServer()

	gitalypb.RegisterRepositoryServiceServer(server, mockServer)
	healthpb.RegisterHealthServer(server, health.NewServer())
	reflection.Register(server)

	go server.Serve(listener)

	return mockServer, "unix://" + serverSocketPath, server.Stop
}

func waitForRequest(t *testing.T, ch chan proto.Message, expected proto.Message, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case req := <-ch:
		testhelper.ProtoEqual(t, expected, req)
		close(ch)
	case <-timer.C:
		t.Fatal("timed out")
	}
}

func TestConfirmReplication(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	srvSocketPath, clean := runFullGitalyServer(t)
	defer clean()

	testRepoA, testRepoAPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	testRepoB, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(gitalyauth.RPCCredentialsV2(testhelper.RepositoryAuthToken)),
	}
	conn, err := grpc.Dial(srvSocketPath, connOpts...)
	require.NoError(t, err)

	var replicator defaultReplicator

	equal, err := replicator.confirmChecksums(ctx, testhelper.DiscardTestLogger(t), gitalypb.NewRepositoryServiceClient(conn), gitalypb.NewRepositoryServiceClient(conn), testRepoA, testRepoB)
	require.NoError(t, err)
	require.True(t, equal)

	testhelper.CreateCommit(t, testRepoAPath, "master", &testhelper.CreateCommitOpts{
		Message: "a commit",
	})

	equal, err = replicator.confirmChecksums(ctx, testhelper.DiscardTestLogger(t), gitalypb.NewRepositoryServiceClient(conn), gitalypb.NewRepositoryServiceClient(conn), testRepoA, testRepoB)
	require.NoError(t, err)
	require.False(t, equal)
}

func TestProcessBacklog_FailedJobs(t *testing.T) {
	backupStorageName := "backup"

	backupDir, err := ioutil.TempDir(testhelper.GitlabTestStoragePath(), backupStorageName)
	require.NoError(t, err)
	defer os.RemoveAll(backupDir)
	defer func(oldStorages []gitaly_config.Storage) { gitaly_config.Config.Storages = oldStorages }(gitaly_config.Config.Storages)

	gitaly_config.Config.Storages = append(gitaly_config.Config.Storages, gitaly_config.Storage{
		Name: backupStorageName,
		Path: backupDir,
	})

	require.Len(t, gitaly_config.Config.Storages, 2, "expected 'default' storage and a new one")

	primarySvr, primarySocket := newReplicationService(t)
	defer primarySvr.Stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	primary := config.Node{
		Storage: "default",
		Address: "unix://" + primarySocket,
	}

	secondary := config.Node{
		Storage: backupStorageName,
		Address: "unix://" + primarySocket,
	}

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "praefect",
				Nodes: []*config.Node{
					&primary,
					&secondary,
				},
			},
		},
	}

	ctx, cancel := testhelper.Context()
	defer cancel()

	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewMemoryReplicationEventQueue(conf))
	processed := make(chan struct{})

	dequeues := 0
	queueInterceptor.OnDequeue(func(ctx context.Context, virtual, target string, count int, queue datastore.ReplicationEventQueue) ([]datastore.ReplicationEvent, error) {
		events, err := queue.Dequeue(ctx, virtual, target, count)
		if len(events) > 0 {
			dequeues++
		}
		return events, err
	})

	completedAcks := 0
	failedAcks := 0
	deadAcks := 0

	queueInterceptor.OnAcknowledge(func(ctx context.Context, state datastore.JobState, ids []uint64, queue datastore.ReplicationEventQueue) ([]uint64, error) {
		switch state {
		case datastore.JobStateCompleted:
			require.Equal(t, []uint64{1}, ids)
			completedAcks++
		case datastore.JobStateFailed:
			require.Equal(t, []uint64{2}, ids)
			failedAcks++
		case datastore.JobStateDead:
			require.Equal(t, []uint64{2}, ids)
			deadAcks++
		default:
			require.FailNow(t, "acknowledge is not expected", state)
		}
		ackIDs, err := queue.Acknowledge(ctx, state, ids)
		if completedAcks+failedAcks+deadAcks == 4 {
			close(processed)
		}
		return ackIDs, err
	})

	// this job exists to verify that replication works
	okJob := datastore.ReplicationJob{
		Change:            datastore.UpdateRepo,
		RelativePath:      testRepo.RelativePath,
		TargetNodeStorage: secondary.Storage,
		SourceNodeStorage: primary.Storage,
		VirtualStorage:    "praefect",
	}
	event1, err := queueInterceptor.Enqueue(ctx, datastore.ReplicationEvent{Job: okJob})
	require.NoError(t, err)
	require.Equal(t, uint64(1), event1.ID)

	// this job checks flow for replication event that fails
	failJob := okJob
	failJob.RelativePath = "invalid path to fail the job"
	event2, err := queueInterceptor.Enqueue(ctx, datastore.ReplicationEvent{Job: failJob})
	require.NoError(t, err)
	require.Equal(t, uint64(2), event2.ID)

	logEntry := testhelper.DiscardTestEntry(t)

	nodeMgr, err := nodes.NewManager(logEntry, conf, nil, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)

	replMgr := NewReplMgr(
		logEntry,
		conf.VirtualStorageNames(),
		queueInterceptor,
		datastore.NewMemoryRepositoryStore(conf.StorageNames()),
		nodeMgr,
		NodeSetFromNodeManager(nodeMgr),
	)
	replMgr.ProcessBacklog(ctx, noopBackoffFunc)

	select {
	case <-processed:
	case <-time.After(60 * time.Second):
		// strongly depends on the processing capacity
		t.Fatal("time limit expired for job to complete")
	}

	require.Equal(t, 3, dequeues, "expected 1 deque to get [okJob, failJob] and 2 more for [failJob] only")
	require.Equal(t, 2, failedAcks)
	require.Equal(t, 1, deadAcks)
	require.Equal(t, 1, completedAcks)
}

func TestProcessBacklog_Success(t *testing.T) {
	defer func(oldStorages []gitaly_config.Storage) { gitaly_config.Config.Storages = oldStorages }(gitaly_config.Config.Storages)

	backupDir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(backupDir)

	gitaly_config.Config.Storages = append(gitaly_config.Config.Storages, gitaly_config.Storage{
		Name: "backup",
		Path: backupDir,
	})
	require.Len(t, gitaly_config.Config.Storages, 2, "expected 'default' storage and a new one")

	primarySvr, primarySocket := newReplicationService(t)
	defer primarySvr.Stop()

	testRepo, _, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	primary := config.Node{
		Storage: "default",
		Address: "unix://" + primarySocket,
	}

	secondary := config.Node{
		Storage: "backup",
		Address: "unix://" + primarySocket,
	}

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "default",
				Nodes: []*config.Node{
					&primary,
					&secondary,
				},
			},
		},
	}

	ctx, cancel := testhelper.Context()
	defer cancel()

	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewMemoryReplicationEventQueue(conf))

	processed := make(chan struct{})
	queueInterceptor.OnAcknowledge(func(ctx context.Context, state datastore.JobState, ids []uint64, queue datastore.ReplicationEventQueue) ([]uint64, error) {
		ackIDs, err := queue.Acknowledge(ctx, state, ids)
		if len(ids) > 0 {
			require.Equal(t, datastore.JobStateCompleted, state, "no fails expected")
			require.Equal(t, []uint64{1, 3, 4}, ids, "all jobs must be processed at once")
			close(processed)
		}
		return ackIDs, err
	})

	var healthUpdated int32
	queueInterceptor.OnStartHealthUpdate(func(ctx context.Context, trigger <-chan time.Time, events []datastore.ReplicationEvent) error {
		require.Len(t, events, 3)
		atomic.AddInt32(&healthUpdated, 1)
		return nil
	})

	// Update replication job
	eventType1 := datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			Change:            datastore.UpdateRepo,
			RelativePath:      testRepo.GetRelativePath(),
			TargetNodeStorage: secondary.Storage,
			SourceNodeStorage: primary.Storage,
			VirtualStorage:    conf.VirtualStorages[0].Name,
		},
	}

	_, err = queueInterceptor.Enqueue(ctx, eventType1)
	require.NoError(t, err)

	_, err = queueInterceptor.Enqueue(ctx, eventType1)
	require.NoError(t, err)

	renameTo1 := filepath.Join(testRepo.GetRelativePath(), "..", filepath.Base(testRepo.GetRelativePath())+"-mv1")
	fullNewPath1 := filepath.Join(backupDir, renameTo1)

	renameTo2 := filepath.Join(renameTo1, "..", filepath.Base(testRepo.GetRelativePath())+"-mv2")
	fullNewPath2 := filepath.Join(backupDir, renameTo2)

	// Rename replication job
	eventType2 := datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			Change:            datastore.RenameRepo,
			RelativePath:      testRepo.GetRelativePath(),
			TargetNodeStorage: secondary.Storage,
			SourceNodeStorage: primary.Storage,
			VirtualStorage:    conf.VirtualStorages[0].Name,
			Params:            datastore.Params{"RelativePath": renameTo1},
		},
	}

	_, err = queueInterceptor.Enqueue(ctx, eventType2)
	require.NoError(t, err)

	// Rename replication job
	eventType3 := datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			Change:            datastore.RenameRepo,
			RelativePath:      renameTo1,
			TargetNodeStorage: secondary.Storage,
			SourceNodeStorage: primary.Storage,
			VirtualStorage:    conf.VirtualStorages[0].Name,
			Params:            datastore.Params{"RelativePath": renameTo2},
		},
	}
	require.NoError(t, err)

	_, err = queueInterceptor.Enqueue(ctx, eventType3)
	require.NoError(t, err)

	logEntry := testhelper.DiscardTestEntry(t)

	nodeMgr, err := nodes.NewManager(logEntry, conf, nil, nil, nil, promtest.NewMockHistogramVec(), protoregistry.GitalyProtoPreregistered, nil)
	require.NoError(t, err)
	nodeMgr.Start(0, time.Hour)

	replMgr := NewReplMgr(
		logEntry,
		conf.VirtualStorageNames(),
		queueInterceptor,
		datastore.NewMemoryRepositoryStore(conf.StorageNames()),
		nodeMgr,
		NodeSetFromNodeManager(nodeMgr),
	)
	replMgr.ProcessBacklog(ctx, noopBackoffFunc)

	select {
	case <-processed:
		require.EqualValues(t, 1, atomic.LoadInt32(&healthUpdated), "health update should be called")
	case <-time.After(30 * time.Second):
		// strongly depends on the processing capacity
		t.Fatal("time limit expired for job to complete")
	}

	_, serr := os.Stat(fullNewPath1)
	require.True(t, os.IsNotExist(serr), "repository must be moved from %q to the new location", fullNewPath1)
	require.True(t, storage.IsGitDirectory(fullNewPath2), "repository must exist at new last RenameRepository location")
}

func TestReplMgrProcessBacklog_OnlyHealthyNodes(t *testing.T) {
	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: "default",
				Nodes: []*config.Node{
					{Storage: "node-1"},
					{Storage: "node-2"},
					{Storage: "node-3"},
				},
			},
		},
	}

	ctx, cancel := testhelper.Context()

	first := true
	queueInterceptor := datastore.NewReplicationEventQueueInterceptor(datastore.NewMemoryReplicationEventQueue(conf))
	queueInterceptor.OnDequeue(func(_ context.Context, virtualStorageName string, storageName string, _ int, _ datastore.ReplicationEventQueue) ([]datastore.ReplicationEvent, error) {
		select {
		case <-ctx.Done():
			return nil, nil
		default:
			if first {
				first = false
				require.Equal(t, conf.VirtualStorages[0].Name, virtualStorageName)
				require.Equal(t, conf.VirtualStorages[0].Nodes[0].Storage, storageName)
				return nil, nil
			}

			assert.Equal(t, conf.VirtualStorages[0].Name, virtualStorageName)
			assert.Equal(t, conf.VirtualStorages[0].Nodes[2].Storage, storageName)
			cancel()
			return nil, nil
		}
	})

	virtualStorage := conf.VirtualStorages[0].Name
	node1 := Node{Storage: conf.VirtualStorages[0].Nodes[0].Storage}
	node2 := Node{Storage: conf.VirtualStorages[0].Nodes[1].Storage}
	node3 := Node{Storage: conf.VirtualStorages[0].Nodes[2].Storage}

	replMgr := NewReplMgr(
		testhelper.DiscardTestEntry(t),
		conf.VirtualStorageNames(),
		queueInterceptor,
		nil,
		StaticHealthChecker{virtualStorage: {node1.Storage, node3.Storage}},
		NodeSet{
			virtualStorage: {
				node1.Storage: node1,
				node2.Storage: node2,
				node3.Storage: node3,
			},
		},
	)
	replMgr.ProcessBacklog(ctx, noopBackoffFunc)

	select {
	case <-ctx.Done():
		// completed by scenario
	case <-time.After(30 * time.Second):
		// strongly depends on the processing capacity
		t.Fatal("time limit expired for job to complete")
	}
}

type mockReplicator struct {
	Replicator
	ReplicateFunc func(ctx context.Context, event datastore.ReplicationEvent, source, target *grpc.ClientConn) error
}

func (m mockReplicator) Replicate(ctx context.Context, event datastore.ReplicationEvent, source, target *grpc.ClientConn) error {
	return m.ReplicateFunc(ctx, event, source, target)
}

func TestProcessBacklog_ReplicatesToReadOnlyPrimary(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	const virtualStorage = "virtal-storage"
	const primaryStorage = "storage-1"
	const secondaryStorage = "storage-2"

	primaryConn := &grpc.ClientConn{}
	secondaryConn := &grpc.ClientConn{}

	conf := config.Config{
		VirtualStorages: []*config.VirtualStorage{
			{
				Name: virtualStorage,
				Nodes: []*config.Node{
					{Storage: primaryStorage},
					{Storage: secondaryStorage},
				},
			},
		},
	}

	queue := datastore.NewMemoryReplicationEventQueue(conf)
	_, err := queue.Enqueue(ctx, datastore.ReplicationEvent{
		Job: datastore.ReplicationJob{
			Change:            datastore.UpdateRepo,
			RelativePath:      "ignored",
			TargetNodeStorage: primaryStorage,
			SourceNodeStorage: secondaryStorage,
			VirtualStorage:    virtualStorage,
		},
	})
	require.NoError(t, err)

	replMgr := NewReplMgr(
		testhelper.DiscardTestEntry(t),
		conf.VirtualStorageNames(),
		queue,
		datastore.NewMemoryRepositoryStore(conf.StorageNames()),
		StaticHealthChecker{virtualStorage: {primaryStorage, secondaryStorage}},
		NodeSet{virtualStorage: {
			primaryStorage:   {Storage: primaryStorage, Connection: primaryConn},
			secondaryStorage: {Storage: secondaryStorage, Connection: secondaryConn},
		}},
	)

	processed := make(chan struct{})
	replMgr.replicator = mockReplicator{
		ReplicateFunc: func(ctx context.Context, event datastore.ReplicationEvent, source, target *grpc.ClientConn) error {
			require.True(t, primaryConn == target)
			require.True(t, secondaryConn == source)
			close(processed)
			return nil
		},
	}
	replMgr.ProcessBacklog(ctx, noopBackoffFunc)
	select {
	case <-processed:
	case <-time.After(5 * time.Second):
		t.Fatalf("replication job targeting read-only primary was not processed before timeout")
	}
}

func TestBackoff(t *testing.T) {
	start := 1 * time.Microsecond
	max := 6 * time.Microsecond
	expectedBackoffs := []time.Duration{
		1 * time.Microsecond,
		2 * time.Microsecond,
		4 * time.Microsecond,
		6 * time.Microsecond,
		6 * time.Microsecond,
		6 * time.Microsecond,
	}
	b, reset := ExpBackoffFunc(start, max)()
	for _, expectedBackoff := range expectedBackoffs {
		require.Equal(t, expectedBackoff, b())
	}

	reset()
	require.Equal(t, start, b())
}

func runFullGitalyServer(t *testing.T) (string, func()) {
	conns := client.NewPool()
	server := serverPkg.NewInsecure(RubyServer, hook.NewManager(hook.GitlabAPIStub, gitaly_config.Config), gitaly_config.Config, conns)

	serverSocketPath := testhelper.GetTemporaryGitalySocketFileName()

	listener, err := net.Listen("unix", serverSocketPath)
	if err != nil {
		t.Fatal(err)
	}
	//listen on internal socket
	internalListener, err := net.Listen("unix", gitaly_config.GitalyInternalSocketPath())
	require.NoError(t, err)

	go server.Serve(listener)
	go server.Serve(internalListener)

	return "unix://" + serverSocketPath, func() {
		conns.Close()
		server.Stop()
	}
}

// newReplicationService is a grpc service that has the SSH, Repository, Remote and ObjectPool services, which
// are the only ones needed for replication
func newReplicationService(tb testing.TB) (*grpc.Server, string) {
	socketName := testhelper.GetTemporaryGitalySocketFileName()
	internalSocketName := gitaly_config.GitalyInternalSocketPath()
	require.NoError(tb, os.RemoveAll(internalSocketName))

	svr := testhelper.NewTestGrpcServer(tb, nil, nil)

	locator := gitaly_config.NewLocator(gitaly_config.Config)
	gitalypb.RegisterRepositoryServiceServer(svr, repository.NewServer(gitaly_config.Config, RubyServer, locator, internalSocketName))
	gitalypb.RegisterObjectPoolServiceServer(svr, objectpoolservice.NewServer(locator))
	gitalypb.RegisterRemoteServiceServer(svr, remote.NewServer(RubyServer, locator))
	gitalypb.RegisterSSHServiceServer(svr, ssh.NewServer(locator))
	gitalypb.RegisterRefServiceServer(svr, ref.NewServer(locator))
	healthpb.RegisterHealthServer(svr, health.NewServer())
	reflection.Register(svr)

	listener, err := net.Listen("unix", socketName)
	require.NoError(tb, err)

	internalListener, err := net.Listen("unix", internalSocketName)
	require.NoError(tb, err)

	go svr.Serve(listener)         // listens for incoming requests
	go svr.Serve(internalListener) // listens for internal requests (service need to access another service on same server)

	return svr, socketName
}

func newRepositoryClient(t *testing.T, serverSocketPath string) (gitalypb.RepositoryServiceClient, *grpc.ClientConn) {
	connOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithPerRPCCredentials(gitalyauth.RPCCredentialsV2(testhelper.RepositoryAuthToken)),
	}
	conn, err := grpc.Dial(serverSocketPath, connOpts...)
	if err != nil {
		t.Fatal(err)
	}

	return gitalypb.NewRepositoryServiceClient(conn), conn
}

var RubyServer = &rubyserver.Server{}

func TestMain(m *testing.M) {
	testhelper.Configure()
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	gitaly_config.Config.Auth.Token = testhelper.RepositoryAuthToken

	var err error
	gitaly_config.Config.GitlabShell.Dir, err = filepath.Abs("testdata/gitlab-shell")
	if err != nil {
		log.Fatal(err)
	}

	testhelper.ConfigureGitalySSH()

	if err := RubyServer.Start(); err != nil {
		log.Fatal(err)
	}
	defer RubyServer.Stop()

	return m.Run()
}

func TestSubtractUint64(t *testing.T) {
	testCases := []struct {
		desc  string
		left  []uint64
		right []uint64
		exp   []uint64
	}{
		{desc: "empty left", left: nil, right: []uint64{1, 2}, exp: nil},
		{desc: "empty right", left: []uint64{1, 2}, right: []uint64{}, exp: []uint64{1, 2}},
		{desc: "some exists", left: []uint64{1, 2, 3, 4, 5}, right: []uint64{2, 4, 5}, exp: []uint64{1, 3}},
		{desc: "nothing exists", left: []uint64{10, 20}, right: []uint64{100, 200}, exp: []uint64{10, 20}},
		{desc: "duplicates exists", left: []uint64{1, 1, 2, 3, 3, 4, 4, 5}, right: []uint64{3, 4, 4, 5}, exp: []uint64{1, 1, 2}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			require.Equal(t, testCase.exp, subtractUint64(testCase.left, testCase.right))
		})
	}
}

func TestReplMgr_ProcessStale(t *testing.T) {
	logger := testhelper.DiscardTestLogger(t)
	hook := test.NewLocal(logger)

	queue := datastore.NewReplicationEventQueueInterceptor(nil)
	mgr := NewReplMgr(logger.WithField("test", t.Name()), nil, queue, datastore.NewMemoryRepositoryStore(nil), nil, nil)

	var counter int32
	queue.OnAcknowledgeStale(func(ctx context.Context, duration time.Duration) error {
		counter++
		if counter > 2 {
			return assert.AnError
		}
		return nil
	})

	ctx, cancel := testhelper.Context()
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 350*time.Millisecond)
	defer cancel()

	done := mgr.ProcessStale(ctx, 100*time.Millisecond, time.Second)

	select {
	case <-time.After(time.Second):
		require.FailNow(t, "execution had stuck")
	case <-done:
	}

	require.Equal(t, int32(3), counter)
	require.Len(t, hook.Entries, 1)
	require.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	require.Equal(t, "background periodical acknowledgement for stale replication jobs", hook.LastEntry().Message)
	require.Equal(t, "replication_manager", hook.LastEntry().Data["component"])
	require.Equal(t, assert.AnError, hook.LastEntry().Data["error"])
}
