package praefect

import (
	"fmt"

	"gitlab.com/gitlab-org/gitaly/internal/praefect/config"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/nodes"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/nodes/tracker"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/protoregistry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// NodeSet contains nodes by their virtual storage and storage names.
type NodeSet map[string]map[string]Node

func (set NodeSet) Close() {
	for _, nodes := range set {
		for _, node := range nodes {
			node.Connection.Close()
		}
	}
}

func (set NodeSet) HealthClients() nodes.HealthClients {
	clients := make(nodes.HealthClients, len(set))
	for virtualStorage, nodes := range set {
		clients[virtualStorage] = make(map[string]grpc_health_v1.HealthClient, len(nodes))
		for _, node := range nodes {
			clients[virtualStorage][node.Storage] = grpc_health_v1.NewHealthClient(node.Connection)
		}
	}

	return clients
}

func (set NodeSet) Connections() Connections {
	conns := make(Connections, len(set))
	for virtualStorage, nodes := range set {
		conns[virtualStorage] = make(map[string]*grpc.ClientConn, len(nodes))
		for _, node := range nodes {
			conns[virtualStorage][node.Storage] = node.Connection
		}
	}

	return conns
}

func NodeSetFromNodeManager(mgr nodes.Manager) NodeSet {
	nodes := mgr.Nodes()

	set := make(NodeSet, len(nodes))
	for virtualStorage, nodes := range nodes {
		set[virtualStorage] = make(map[string]Node, len(nodes))
		for _, node := range nodes {
			set[virtualStorage][node.GetStorage()] = Node{
				Storage:    node.GetStorage(),
				Address:    node.GetAddress(),
				Token:      node.GetToken(),
				Connection: node.GetConnection(),
			}
		}
	}

	return set
}

// DialNodes dials the configured storage nodes.
func DialNodes(
	virtualStorages []*config.VirtualStorage,
	registry *protoregistry.Registry,
	errorTracker tracker.ErrorTracker,
) (NodeSet, error) {
	set := make(NodeSet, len(virtualStorages))
	// We dial the same node multiple times here if the same node is configured for multiple
	// virtual storages. This is necessary as the error tracker should meter by storage rather than
	// per server.
	for _, virtualStorage := range virtualStorages {
		set[virtualStorage.Name] = make(map[string]Node, len(virtualStorage.Nodes))
		for _, node := range virtualStorage.Nodes {
			conn, err := nodes.Dial(node.Address, node.Token, node.Storage, registry, errorTracker)
			if err != nil {
				return nil, fmt.Errorf("dial %q/%q: %w", virtualStorage.Name, node.Storage, err)
			}

			set[virtualStorage.Name][node.Storage] = Node{
				Storage:    node.Storage,
				Address:    node.Address,
				Token:      node.Token,
				Connection: conn,
			}
		}
	}

	return set, nil
}
