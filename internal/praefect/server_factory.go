package praefect

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/grpc-proxy/proxy"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/nodes"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/protoregistry"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/transactions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewServerFactory returns factory object for initialization of praefect gRPC servers.
func NewServerFactory(
	conf config.Config,
	logger *logrus.Entry,
	director proxy.StreamDirector,
	nodeMgr nodes.Manager,
	txMgr *transactions.Manager,
	queue datastore.ReplicationEventQueue,
	rs datastore.RepositoryStore,
	registry *protoregistry.Registry,
) *ServerFactory {
	return &ServerFactory{
		conf:     conf,
		logger:   logger,
		director: director,
		nodeMgr:  nodeMgr,
		txMgr:    txMgr,
		queue:    queue,
		rs:       rs,
		registry: registry,
	}
}

// ServerFactory is a factory of praefect grpc servers
type ServerFactory struct {
	mtx              sync.Mutex
	conf             config.Config
	logger           *logrus.Entry
	director         proxy.StreamDirector
	nodeMgr          nodes.Manager
	txMgr            *transactions.Manager
	queue            datastore.ReplicationEventQueue
	rs               datastore.RepositoryStore
	registry         *protoregistry.Registry
	secure, insecure []*grpc.Server
}

// Serve starts serving on the provided listener with newly created grpc.Server
func (s *ServerFactory) Serve(l net.Listener, secure bool) error {
	srv, err := s.create(secure)
	if err != nil {
		return err
	}

	return srv.Serve(l)
}

// Stop stops all servers created by the factory.
func (s *ServerFactory) Stop() {
	for _, srv := range s.all() {
		srv.Stop()
	}
}

// GracefulStop stops both the secure and insecure servers gracefully.
func (s *ServerFactory) GracefulStop() {
	wg := sync.WaitGroup{}

	for _, srv := range s.all() {
		wg.Add(1)

		go func(s *grpc.Server) {
			s.GracefulStop()
			wg.Done()
		}(srv)
	}

	wg.Wait()
}

func (s *ServerFactory) create(secure bool) (*grpc.Server, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if !secure {
		s.insecure = append(s.insecure, s.createGRPC())
		return s.insecure[len(s.insecure)-1], nil
	}

	cert, err := tls.LoadX509KeyPair(s.conf.TLS.CertPath, s.conf.TLS.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load certificate key pair: %w", err)
	}

	s.secure = append(s.secure, s.createGRPC(grpc.Creds(credentials.NewServerTLSFromCert(&cert))))

	return s.secure[len(s.secure)-1], nil
}

func (s *ServerFactory) createGRPC(grpcOpts ...grpc.ServerOption) *grpc.Server {
	return NewGRPCServer(
		s.conf,
		s.logger,
		s.registry,
		s.director,
		s.nodeMgr,
		s.txMgr,
		s.queue,
		s.rs,
		grpcOpts...,
	)
}

func (s *ServerFactory) all() []*grpc.Server {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	servers := make([]*grpc.Server, 0, len(s.secure)+len(s.insecure))
	servers = append(servers, s.secure...)
	servers = append(servers, s.insecure...)
	return servers
}
