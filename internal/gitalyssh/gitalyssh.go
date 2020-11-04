package gitalyssh

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	gitaly_x509 "gitlab.com/gitlab-org/gitaly/internal/x509"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"gitlab.com/gitlab-org/labkit/correlation"
	"gitlab.com/gitlab-org/labkit/tracing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// GitalyInternalURL is a special URL that indicates Gitaly wants to
	// push or fetch to another Gitaly instance
	GitalyInternalURL = "ssh://gitaly/internal.git"
)

const (
	// EnvVarUploadPackAllowAnySHA1InWant enables the capability to request
	// individual SHA1's from the remote repo
	EnvVarUploadPackAllowAnySHA1InWant = "uploadpack.allowAnySHA1InWant=true"
)

var (
	envInjector = tracing.NewEnvInjector()
)

func UploadPackEnv(ctx context.Context, req *gitalypb.SSHUploadPackRequest) ([]string, error) {
	env, err := commandEnv(ctx, req.Repository.StorageName, "upload-pack", req)
	if err != nil {
		return nil, err
	}
	return envInjector(ctx, env), nil
}

func commandEnv(ctx context.Context, storageName, command string, message proto.Message) ([]string, error) {
	var pbMarshaler jsonpb.Marshaler
	payload, err := pbMarshaler.MarshalToString(message)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "commandEnv: marshalling payload failed: %v", err)
	}

	serversInfo, err := helper.ExtractGitalyServers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "commandEnv: extracting Gitaly servers: %v", err)
	}

	storageInfo, ok := serversInfo[storageName]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "commandEnv: no storage info for %s", storageName)
	}

	if storageInfo.Address == "" {
		return nil, status.Errorf(codes.InvalidArgument, "commandEnv: empty gitaly address")
	}

	featureFlagPairs := featureflag.AllFlags(ctx)

	return []string{
		fmt.Sprintf("GITALY_PAYLOAD=%s", payload),
		fmt.Sprintf("GIT_SSH_COMMAND=%s %s", gitalySSHPath(), command),
		fmt.Sprintf("GITALY_ADDRESS=%s", storageInfo.Address),
		fmt.Sprintf("GITALY_TOKEN=%s", storageInfo.Token),
		fmt.Sprintf("GITALY_FEATUREFLAGS=%s", strings.Join(featureFlagPairs, ",")),
		fmt.Sprintf("CORRELATION_ID=%s", getCorrelationID(ctx)),
		// please see https://github.com/git/git/commit/0da0e49ba12225684b75e86a4c9344ad121652cb for mote details
		"GIT_SSH_VARIANT=simple",
		// Pass through the SSL_CERT_* variables that indicate which
		// system certs to trust
		fmt.Sprintf("%s=%s", gitaly_x509.SSLCertDir, os.Getenv(gitaly_x509.SSLCertDir)),
		fmt.Sprintf("%s=%s", gitaly_x509.SSLCertFile, os.Getenv(gitaly_x509.SSLCertFile)),
	}, nil
}

func gitalySSHPath() string {
	return filepath.Join(config.Config.BinDir, "gitaly-ssh")
}

func getCorrelationID(ctx context.Context) string {
	correlationID := correlation.ExtractFromContext(ctx)
	if correlationID != "" {
		return correlationID
	}

	return correlation.SafeRandomID()
}
