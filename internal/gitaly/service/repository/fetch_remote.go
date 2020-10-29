package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/rubyserver"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/metadata/featureflag"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/status"
)

var (
	fetchRemoteImplCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gitaly_fetch_remote_counter",
			Help: "Number of calls to FetchRemote rpc for each implementation (ruby/go)",
		},
		[]string{"impl"},
	)
)

func init() {
	prometheus.MustRegister(fetchRemoteImplCounter)
}

func (s *server) FetchRemote(ctx context.Context, req *gitalypb.FetchRemoteRequest) (*gitalypb.FetchRemoteResponse, error) {
	if featureflag.IsDisabled(ctx, featureflag.GoFetchRemote) {
		fetchRemoteImplCounter.WithLabelValues("ruby").Inc()

		client, err := s.ruby.RepositoryServiceClient(ctx)
		if err != nil {
			return nil, err
		}

		clientCtx, err := rubyserver.SetHeaders(ctx, req.GetRepository())
		if err != nil {
			return nil, err
		}

		return client.FetchRemote(clientCtx, req)
	}

	fetchRemoteImplCounter.WithLabelValues("go").Inc()

	if err := s.validateFetchRemoteRequest(req); err != nil {
		return nil, err
	}

	var stderr bytes.Buffer
	opts := git.FetchOpts{Stderr: &stderr, Force: req.Force, Prune: true, Tags: git.FetchOptsTagsAll}

	if req.GetNoTags() {
		opts.Tags = git.FetchOptsTagsNone
	}

	repo := git.NewRepository(req.GetRepository())
	params := req.GetRemoteParams()
	remoteName := req.GetRemote()

	if params != nil {
		remoteName = params.GetName()
		remoteURL := params.GetUrl()
		refspecs := s.getRefspecs(params.GetMirrorRefmaps())

		if err := s.setRemote(ctx, repo, remoteName, remoteURL); err != nil {
			return nil, fmt.Errorf("set remote: %w", err)
		}

		defer func(ctx context.Context) {
			// we pass context as it may be overridden in case timeout is set for the call
			if err := s.removeRemote(ctx, repo, remoteName); err != nil {
				ctxlogrus.Extract(ctx).WithError(err).WithFields(logrus.Fields{
					"remote":  remoteName,
					"storage": req.GetRepository().GetStorageName(),
					"path":    req.GetRepository().GetRelativePath(),
				}).Error("removal of remote failed")
			}
		}(ctx)

		if err := s.configureRemote(ctx, repo, remoteName, refspecs); err != nil {
			return nil, fmt.Errorf("set configs of the remote: %w", err)
		}

		opts.Global = append(opts.Global,
			git.ValueFlag{Name: "-c", Value: "http.followRedirects=false"},
			git.ValueFlag{Name: "-c", Value: "http." + remoteURL + ".extraHeader=Authorization: " + params.GetHttpAuthorizationHeader()},
		)
	} else {
		envGitSSHCommand, cleanup, err := s.configureSSH(ctx, req.GetSshKey(), req.GetKnownHosts())
		if err != nil {
			return nil, err
		}
		defer cleanup()

		opts.Env = append(opts.Env, envGitSSHCommand)
	}

	if req.GetTimeout() > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.GetTimeout())*time.Second)
		defer cancel()
	}

	if err := repo.FetchRemote(ctx, remoteName, opts); err != nil {
		if _, ok := status.FromError(err); ok {
			// this check is used because of internal call to alternates.PathAndEnv
			// which may return gRPC status as an error result
			return nil, err
		}

		errMsg := stderr.String()
		if errMsg != "" {
			return nil, fmt.Errorf("fetch remote: %q: %w", errMsg, err)
		}

		return nil, fmt.Errorf("fetch remote: %w", err)
	}

	return &gitalypb.FetchRemoteResponse{}, nil
}

func (s *server) validateFetchRemoteRequest(req *gitalypb.FetchRemoteRequest) error {
	if req.GetRepository() == nil {
		return helper.ErrInvalidArgument(errors.New(`no "repository"`))
	}

	params := req.GetRemoteParams()
	if params == nil {
		return nil
	}

	remoteURL, err := url.ParseRequestURI(params.GetUrl())
	if err != nil {
		return helper.ErrInvalidArgument(fmt.Errorf(`invalid "remote_params.url": %q: %w`, params.GetUrl(), err))
	}

	if remoteURL.Host == "" {
		return helper.ErrInvalidArgumentf(`invalid "remote_params.url": %q: no host`, params.GetUrl())
	}

	remote := params.GetName()
	if strings.TrimSpace(remote) == "" {
		return helper.ErrInvalidArgument(fmt.Errorf(`blank or empty "remote_params.name": %q`, params.Name))
	}

	return nil
}

func (s *server) getRefspecs(refmaps []string) []string {
	refspecs := make([]string, 0, len(refmaps))

	for _, refmap := range refmaps {
		switch refmap {
		case "all_refs":
			// with `all_refs`, the repository is equivalent to the result of `git clone --mirror`
			refspecs = append(refspecs, "+refs/*:refs/*")
		case "heads":
			refspecs = append(refspecs, "+refs/heads/*:refs/heads/*")
		case "tags":
			refspecs = append(refspecs, "+refs/tags/*:refs/tags/*")
		default:
			refspecs = append(refspecs, refmap)
		}
	}
	return refspecs
}

func (s *server) setRemote(ctx context.Context, repo git.Repository, name, url string) error {
	if err := repo.Remote().Remove(ctx, name); err != nil {
		if err != git.ErrNotFound {
			return fmt.Errorf("remove remote: %w", err)
		}
	}

	if err := repo.Remote().Add(ctx, name, url, git.RemoteAddOpts{}); err != nil {
		return fmt.Errorf("add remote: %w", err)
	}

	return nil
}

func (s *server) removeRemote(ctx context.Context, repo git.Repository, name string) error {
	if err := repo.Remote().Remove(ctx, name); err != nil {
		if err != git.ErrNotFound {
			return fmt.Errorf("remove remote: %w", err)
		}
	}

	return nil
}

func (s *server) configureRemote(ctx context.Context, repo git.Repository, name string, refspecs []string) error {
	for _, refspec := range refspecs {
		if err := repo.Config().Add(ctx, "remote."+name+".fetch", refspec, git.ConfigAddOpts{}); err != nil {
			return fmt.Errorf("config set fetch for remote: %q: %w", name, err)
		}
	}

	if err := repo.Config().Add(ctx, "remote."+name+".mirror", "true", git.ConfigAddOpts{}); err != nil {
		return fmt.Errorf("config set mirror for remote: %q: %w", name, err)
	}

	if err := repo.Config().Add(ctx, "remote."+name+".prune", "true", git.ConfigAddOpts{}); err != nil {
		return fmt.Errorf("config set prune for remote: %q: %w", name, err)
	}

	return nil
}

func (s *server) configureSSH(ctx context.Context, sshKey, knownHosts string) (string, func(), error) {
	if sshKey == "" && knownHosts == "" {
		return "", func() {}, nil
	}

	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		if err := os.RemoveAll(tmpdir); err != nil {
			ctxlogrus.Extract(ctx).WithError(err).Error("failed to remove tmp directory with ssh key/config")
		}
	}

	var conf []string

	if sshKey != "" {
		identityFilePath := filepath.Join(tmpdir, "gitlab-shell-key-file")

		if err := ioutil.WriteFile(identityFilePath, []byte(sshKey), 0400); err != nil {
			cleanup()
			return "", nil, err
		}
		conf = append(conf, "IdentitiesOnly=yes", "IdentityFile="+identityFilePath)
	}

	if knownHosts != "" {
		hostsFilePath := filepath.Join(tmpdir, "gitlab-shell-known-hosts")

		if err := ioutil.WriteFile(hostsFilePath, []byte(knownHosts), 0400); err != nil {
			cleanup()
			return "", nil, err
		}
		conf = append(conf, "StrictHostKeyChecking=yes", "UserKnownHostsFile="+hostsFilePath)
	}

	return "GIT_SSH_COMMAND=ssh " + strings.Join(conf, " "), cleanup, nil
}
