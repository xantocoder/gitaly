package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/git-lfs/git-lfs/lfs"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/hook"
	"gitlab.com/gitlab-org/labkit/log"
	"gitlab.com/gitlab-org/labkit/tracing"
)

type configProvider interface {
	Get(key string) string
}

func smudge(to io.Writer, from io.Reader, cfgProvider configProvider) error {
	output, err := handleSmudge(to, from, cfgProvider)
	if err != nil {
		log.WithError(err).Error(err)
		return err
	}

	_, copyErr := io.Copy(to, output)
	if copyErr != nil {
		log.WithError(err).Error(copyErr)
		return copyErr
	}

	return nil
}

func handleSmudge(to io.Writer, from io.Reader, config configProvider) (io.Reader, error) {
	ptr, contents, err := lfs.DecodeFrom(from)
	if err != nil {
		// This isn't a valid LFS pointer. Just copy the existing pointer data.
		return contents, nil
	}

	log.WithField("oid", ptr.Oid).Debug("decoded LFS OID")

	cfg, glRepository, err := loadConfig(config)
	if err != nil {
		return contents, err
	}

	log.WithField("gitlab_config", cfg).Debug("loaded GitLab API config")

	client, err := hook.NewGitlabNetClient(cfg)
	if err != nil {
		return contents, err
	}

	// Since the environment is sanitized at the moment, we're only
	// using this to extract the correlation ID. The finished() call
	// to clean up the tracing will be a NOP here.
	ctx, finished := tracing.ExtractFromEnv(context.Background())
	defer finished()

	url := fmt.Sprintf("/lfs?oid=%s&gl_repository=%s", ptr.Oid, glRepository)
	response, err := client.Get(ctx, url)
	if err != nil {
		return contents, fmt.Errorf("error loading LFS object: %v", err)
	}

	// This cannot go to STDOUT or it will corrupt the stream
	log.WithFields(logrus.Fields{
		"status_code": response.StatusCode,
		"url":         url,
	}).Info("completed HTTP request")

	if response.StatusCode == 200 {
		return response.Body, nil
	}

	return contents, nil
}

func loadConfig(cfgProvider configProvider) (config.Gitlab, string, error) {
	var cfg config.Gitlab

	glRepository := cfgProvider.Get("GL_REPOSITORY")
	if glRepository == "" {
		return cfg, "", fmt.Errorf("error loading project: GL_REPOSITORY is not defined")
	}

	u := cfgProvider.Get("GL_INTERNAL_CONFIG")
	if u == "" {
		return cfg, glRepository, fmt.Errorf("unable to retrieve GL_INTERNAL_CONFIG")
	}

	if err := json.Unmarshal([]byte(u), &cfg); err != nil {
		return cfg, glRepository, fmt.Errorf("unable to unmarshal GL_INTERNAL_CONFIG: %v", err)
	}

	return cfg, glRepository, nil
}
