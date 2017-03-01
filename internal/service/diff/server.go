package diff

import (
	pb "gitlab.com/gitlab-org/gitaly-proto/go"
)

type server struct{}

// NewServer creates a new instance of a gRPC DiffServer
func NewServer() pb.DiffServer {
	return &server{}
}
