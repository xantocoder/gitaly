package praefect

import (
	"context"
	"errors"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/nodes"
	"google.golang.org/grpc"
)

// ErrNoSuitableNode is returned when there is not suitable node to serve a request.
var ErrNoSuitableNode = errors.New("no suitable node to serve the request")

// ErrUnknownVirtualStorage is returned when trying to access an unknown virtual storage.
var ErrUnknownVirtualStorage = errors.New("unknown virtual storage")

// Connections is a set of connections to configured storage nodes by their virtual storages.
type Connections map[string]map[string]*grpc.ClientConn

// PrimaryGetter is an interface for getting a primary of a repository.
type PrimaryGetter interface {
	// GetPrimary returns the primary storage for a given repository.
	GetPrimary(ctx context.Context, virtualStorage string, relativePath string) (string, error)
}

// Random is the interface of the Go random number generator.
type Random interface {
	// Intn returns a random integer in the range [0,n).
	Intn(n int) int
}

// RandomFunc is an adapter to turn conforming functions in to a HealthChecker.
type RandomFunc func(n int) int

func (fn RandomFunc) Intn(n int) int { return fn(n) }

// PerRepositoryRouter implements a router that routes requests respecting per repository primary nodes.
type PerRepositoryRouter struct {
	conns Connections
	pg    PrimaryGetter
	rand  Random
	hc    HealthChecker
	rs    datastore.RepositoryStore
}

// NewPerRepositoryRouter returns a new PerRepositoryRouter using the passed configuration.
func NewPerRepositoryRouter(conns Connections, pg PrimaryGetter, hc HealthChecker, rand Random, rs datastore.RepositoryStore) *PerRepositoryRouter {
	return &PerRepositoryRouter{
		conns: conns,
		pg:    pg,
		rand:  rand,
		hc:    hc,
		rs:    rs,
	}
}

func (r *PerRepositoryRouter) healthyNodes(virtualStorage string) ([]Node, error) {
	conns, ok := r.conns[virtualStorage]
	if !ok {
		return nil, ErrUnknownVirtualStorage
	}

	healthyNodes := make([]Node, 0, len(conns))
	for _, storage := range r.hc.HealthyNodes()[virtualStorage] {
		conn, ok := conns[storage]
		if !ok {
			// healthy node not in the local connection set
			continue
		}

		healthyNodes = append(healthyNodes, Node{
			Storage:    storage,
			Connection: conn,
		})
	}

	return healthyNodes, nil
}

func (r *PerRepositoryRouter) pickRandom(nodes []Node) (Node, error) {
	if len(nodes) == 0 {
		return Node{}, ErrNoSuitableNode
	}

	return nodes[r.rand.Intn(len(nodes))], nil
}

// RouteStorageAccessor returns the node which should serve the storage accessor request.
func (r *PerRepositoryRouter) RouteStorageAccessor(ctx context.Context, virtualStorage string) (Node, error) {
	healthyNodes, err := r.healthyNodes(virtualStorage)
	if err != nil {
		return Node{}, err
	}

	return r.pickRandom(healthyNodes)
}

// RouteStorageAccessor returns the primary and secondaries that should handle the storage
// mutator request.
func (r *PerRepositoryRouter) RouteStorageMutator(ctx context.Context, virtualStorage string) (StorageMutatorRoute, error) {
	return StorageMutatorRoute{}, errors.New("RouteStorageMutator is not implemented on PerRepositoryRouter")
}

// RouteRepositoryAccessor returns the node that should serve the repository accessor request.
func (r *PerRepositoryRouter) RouteRepositoryAccessor(ctx context.Context, virtualStorage, relativePath string) (Node, error) {
	healthyNodes, err := r.healthyNodes(virtualStorage)
	if err != nil {
		return Node{}, err
	}

	primary, err := r.pg.GetPrimary(ctx, virtualStorage, relativePath)
	if err != nil {
		return Node{}, fmt.Errorf("get primary: %w", err)
	}

	consistentSecondaries, err := r.rs.GetConsistentSecondaries(ctx, virtualStorage, relativePath, primary)
	if err != nil {
		// this is recoverable error - proceed with primary node
		ctxlogrus.Extract(ctx).WithError(err).Warn("get up to date secondaries")
	}

	consistentSecondaries[primary] = struct{}{}

	healthyConsistentNodes := make([]Node, 0, len(healthyNodes))
	for _, node := range healthyNodes {
		if _, ok := consistentSecondaries[node.Storage]; !ok {
			continue
		}

		healthyConsistentNodes = append(healthyConsistentNodes, node)
	}

	return r.pickRandom(healthyConsistentNodes)
}

// RouteRepositoryMutator returns the primary and secondaries that should handle the repository mutator request.
// Additionally, it returns nodes which should have the change replicated to.
func (r *PerRepositoryRouter) RouteRepositoryMutator(ctx context.Context, virtualStorage, relativePath string) (RepositoryMutatorRoute, error) {
	healthyNodes, err := r.healthyNodes(virtualStorage)
	if err != nil {
		return RepositoryMutatorRoute{}, err
	}

	primary, err := r.pg.GetPrimary(ctx, virtualStorage, relativePath)
	if err != nil {
		return RepositoryMutatorRoute{}, fmt.Errorf("get primary: %w", err)
	}

	if latest, err := r.rs.IsLatestGeneration(ctx, virtualStorage, relativePath, primary); err != nil {
		return RepositoryMutatorRoute{}, fmt.Errorf("is latest generation: %w", err)
	} else if !latest {
		return RepositoryMutatorRoute{}, ErrRepositoryReadOnly
	}

	consistentSecondaries, err := r.rs.GetConsistentSecondaries(ctx, virtualStorage, relativePath, primary)
	if err != nil {
		return RepositoryMutatorRoute{}, fmt.Errorf("consistent secondaries: %w", err)
	}

	var route RepositoryMutatorRoute
	for _, node := range healthyNodes {
		if node.Storage == primary {
			route.Primary = node
			continue
		}

		if _, ok := consistentSecondaries[node.Storage]; !ok {
			route.ReplicationTargets = append(route.ReplicationTargets, node.Storage)
			continue
		}

		route.Secondaries = append(route.Secondaries, node)
	}

	if (route.Primary == Node{}) {
		return RepositoryMutatorRoute{}, nodes.ErrPrimaryNotHealthy
	}

	return route, nil
}
