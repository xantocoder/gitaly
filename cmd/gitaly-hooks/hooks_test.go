package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	gitalyhook "gitlab.com/gitlab-org/gitaly/internal/gitaly/hook"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/service/hook"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/metadata"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
	"google.golang.org/grpc/reflection"
)

func TestMain(m *testing.M) {
	testhelper.Configure()
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	defer testhelper.MustHaveNoChildProcess()

	defer func(rubyDir string) {
		config.Config.Ruby.Dir = rubyDir
	}(config.Config.Ruby.Dir)

	rubyDir, err := filepath.Abs("../../ruby")
	if err != nil {
		log.Fatal(err)
	}

	config.Config.Ruby.Dir = rubyDir

	testhelper.ConfigureGitalyHooksBinary()
	testhelper.ConfigureGitalySSH()

	return m.Run()
}

func TestHooksPrePostWithSymlinkedStoragePath(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	tempDir, cleanup := testhelper.TempDir(t)
	defer cleanup()

	originalStoragePath := config.Config.Storages[0].Path
	symlinkedStoragePath := filepath.Join(tempDir, "storage")
	require.NoError(t, os.Symlink(originalStoragePath, symlinkedStoragePath))

	defer func() {
		config.Config.Storages[0].Path = originalStoragePath
	}()
	config.Config.Storages[0].Path = symlinkedStoragePath

	testHooksPrePostReceive(t)
}

func TestHooksPrePostReceive(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	testHooksPrePostReceive(t)
}

func testHooksPrePostReceive(t *testing.T) {
	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	secretToken := "secret token"
	glID := "key-1234"
	glUsername := "iamgitlab"
	glProtocol := "ssh"
	glRepository := "some_repo"

	tempGitlabShellDir, cleanup := testhelper.CreateTemporaryGitlabShellDir(t)
	defer cleanup()

	changes := "abc"

	gitPushOptions := []string{"gitpushoption1", "gitpushoption2"}
	gitObjectDir := filepath.Join(testRepoPath, "objects", "temp")
	gitAlternateObjectDirs := []string{(filepath.Join(testRepoPath, "objects"))}

	gitlabUser, gitlabPassword := "gitlab_user-1234", "gitlabsecret9887"
	httpProxy, httpsProxy, noProxy := "http://test.example.com:8080", "https://test.example.com:8080", "*"

	c := testhelper.GitlabTestServerOptions{
		User:                        gitlabUser,
		Password:                    gitlabPassword,
		SecretToken:                 secretToken,
		GLID:                        glID,
		GLRepository:                glRepository,
		Changes:                     changes,
		PostReceiveCounterDecreased: true,
		Protocol:                    "ssh",
		GitPushOptions:              gitPushOptions,
		GitObjectDir:                gitObjectDir,
		GitAlternateObjectDirs:      gitAlternateObjectDirs,
		RepoPath:                    testRepoPath,
	}

	serverURL, cleanup := testhelper.NewGitlabTestServer(t, c)
	defer cleanup()

	config.Config.GitlabShell.Dir = tempGitlabShellDir

	testhelper.WriteTemporaryGitlabShellConfigFile(t,
		tempGitlabShellDir,
		testhelper.GitlabShellConfig{
			GitlabURL: serverURL,
			HTTPSettings: testhelper.HTTPSettings{
				User:     gitlabUser,
				Password: gitlabPassword,
			},
		})

	testhelper.WriteShellSecretFile(t, tempGitlabShellDir, secretToken)

	config.Config.Gitlab.URL = serverURL
	config.Config.Gitlab.SecretFile = filepath.Join(tempGitlabShellDir, ".gitlab_shell_secret")
	config.Config.Gitlab.HTTPSettings.User = gitlabUser
	config.Config.Gitlab.HTTPSettings.Password = gitlabPassword

	gitObjectDirRegex := regexp.MustCompile(`(?m)^GIT_OBJECT_DIRECTORY=(.*)$`)
	gitAlternateObjectDirRegex := regexp.MustCompile(`(?m)^GIT_ALTERNATE_OBJECT_DIRECTORIES=(.*)$`)
	token := "abc123"

	hookNames := []string{"pre-receive", "post-receive"}

	for _, hookName := range hookNames {
		t.Run(fmt.Sprintf("hookName: %s", hookName), func(t *testing.T) {
			customHookOutputPath, cleanup := testhelper.WriteEnvToCustomHook(t, testRepoPath, hookName)
			defer cleanup()

			config.Config.Gitlab.URL = serverURL
			config.Config.Gitlab.SecretFile = filepath.Join(tempGitlabShellDir, ".gitlab_shell_secret")
			config.Config.Gitlab.HTTPSettings.User = gitlabUser
			config.Config.Gitlab.HTTPSettings.Password = gitlabPassword

			gitlabAPI, err := gitalyhook.NewGitlabAPI(config.Config.Gitlab)
			require.NoError(t, err)

			socket, stop := runHookServiceServerWithAPI(t, token, gitlabAPI)
			defer stop()

			var stderr, stdout bytes.Buffer
			stdin := bytes.NewBuffer([]byte(changes))
			hookPath, err := filepath.Abs(fmt.Sprintf("../../ruby/git-hooks/%s", hookName))
			require.NoError(t, err)
			cmd := exec.Command(hookPath)
			cmd.Stderr = &stderr
			cmd.Stdout = &stdout
			cmd.Stdin = stdin
			cmd.Env = testhelper.EnvForHooks(
				t,
				tempGitlabShellDir,
				socket,
				token,
				testRepo,
				testhelper.GlHookValues{
					GLID:                   glID,
					GLUsername:             glUsername,
					GLRepo:                 glRepository,
					GLProtocol:             glProtocol,
					GitObjectDir:           c.GitObjectDir,
					GitAlternateObjectDirs: c.GitAlternateObjectDirs,
				},
				testhelper.ProxyValues{
					HTTPProxy:  httpProxy,
					HTTPSProxy: httpsProxy,
					NoProxy:    noProxy,
				},
				gitPushOptions...,
			)

			cmd.Dir = testRepoPath

			require.NoError(t, cmd.Run())
			require.Empty(t, stderr.String())
			require.Empty(t, stdout.String())

			output := string(testhelper.MustReadFile(t, customHookOutputPath))
			requireContainsOnce(t, output, "GL_USERNAME="+glUsername)
			requireContainsOnce(t, output, "GL_ID="+glID)
			requireContainsOnce(t, output, "GL_REPOSITORY="+glRepository)
			requireContainsOnce(t, output, "HTTP_PROXY="+httpProxy)
			requireContainsOnce(t, output, "http_proxy="+httpProxy)
			requireContainsOnce(t, output, "HTTPS_PROXY="+httpsProxy)
			requireContainsOnce(t, output, "https_proxy="+httpsProxy)
			requireContainsOnce(t, output, "no_proxy="+noProxy)
			requireContainsOnce(t, output, "NO_PROXY="+noProxy)

			if hookName == "pre-receive" {
				gitObjectDirMatches := gitObjectDirRegex.FindStringSubmatch(output)
				require.Len(t, gitObjectDirMatches, 2)
				require.Equal(t, gitObjectDir, gitObjectDirMatches[1])

				gitAlternateObjectDirMatches := gitAlternateObjectDirRegex.FindStringSubmatch(output)
				require.Len(t, gitAlternateObjectDirMatches, 2)
				require.Equal(t, strings.Join(gitAlternateObjectDirs, ":"), gitAlternateObjectDirMatches[1])
			} else {
				require.Contains(t, output, "GL_PROTOCOL="+glProtocol)
			}
		})
	}
}

func TestHooksUpdate(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	glID := "key-1234"
	glUsername := "iamgitlab"
	glProtocol := "ssh"
	glRepository := "some_repo"

	tempGitlabShellDir, cleanup := testhelper.CreateTemporaryGitlabShellDir(t)
	defer cleanup()

	customHooksDir, cleanup := testhelper.TempDir(t)
	defer cleanup()

	testhelper.WriteTemporaryGitlabShellConfigFile(t, tempGitlabShellDir, testhelper.GitlabShellConfig{GitlabURL: "http://www.example.com", CustomHooksDir: customHooksDir})

	os.Symlink(filepath.Join(config.Config.GitlabShell.Dir, "config.yml"), filepath.Join(tempGitlabShellDir, "config.yml"))

	testhelper.WriteShellSecretFile(t, tempGitlabShellDir, "the wrong token")

	config.Config.GitlabShell.Dir = tempGitlabShellDir

	token := "abc123"
	socket, stop := runHookServiceServer(t, token)
	defer stop()

	config.Config.Hooks.CustomHooksDir = customHooksDir

	testHooksUpdate(t, tempGitlabShellDir, socket, token, testhelper.GlHookValues{
		GLID:       glID,
		GLUsername: glUsername,
		GLRepo:     glRepository,
		GLProtocol: glProtocol,
	})
}

func testHooksUpdate(t *testing.T, gitlabShellDir, socket, token string, glValues testhelper.GlHookValues) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	refval, oldval, newval := "refval", "oldval", "newval"
	updateHookPath, err := filepath.Abs("../../ruby/git-hooks/update")
	require.NoError(t, err)
	cmd := exec.Command(updateHookPath, refval, oldval, newval)
	cmd.Env = testhelper.EnvForHooks(t, gitlabShellDir, socket, token, testRepo, glValues, testhelper.ProxyValues{})
	cmd.Dir = testRepoPath

	tempDir, cleanup := testhelper.TempDir(t)
	defer cleanup()

	customHookArgsPath := filepath.Join(tempDir, "containsarguments")
	dumpArgsToTempfileScript := fmt.Sprintf(`#!/usr/bin/env ruby
require 'json'
open('%s', 'w') { |f| f.puts(JSON.dump(ARGV)) }
`, customHookArgsPath)
	// write a custom hook to path/to/repo.git/custom_hooks/update.d/dumpargsscript which dumps the args into a tempfile
	cleanup, err = testhelper.WriteExecutable(filepath.Join(testRepoPath, "custom_hooks", "update.d", "dumpargsscript"), []byte(dumpArgsToTempfileScript))
	require.NoError(t, err)
	defer cleanup()

	// write a custom hook to path/to/repo.git/custom_hooks/update which dumps the env into a tempfile
	customHookOutputPath, cleanup := testhelper.WriteEnvToCustomHook(t, testRepoPath, "update")
	defer cleanup()

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = testRepoPath

	require.NoError(t, cmd.Run())
	require.Empty(t, stdout.String())
	require.Empty(t, stderr.String())

	require.FileExists(t, customHookArgsPath)

	var inputs []string

	b, err := ioutil.ReadFile(customHookArgsPath)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, &inputs))
	require.Equal(t, []string{refval, oldval, newval}, inputs)

	output := string(testhelper.MustReadFile(t, customHookOutputPath))
	require.Contains(t, output, "GL_USERNAME="+glValues.GLUsername)
	require.Contains(t, output, "GL_ID="+glValues.GLID)
	require.Contains(t, output, "GL_REPOSITORY="+glValues.GLRepo)
	require.Contains(t, output, "GL_PROTOCOL="+glValues.GLProtocol)
}

func TestHooksPostReceiveFailed(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)
	secretToken := "secret token"
	glID := "key-1234"
	glUsername := "iamgitlab"
	glProtocol := "ssh"
	glRepository := "some_repo"
	changes := "oldhead newhead"

	tempGitlabShellDir, cleanup := testhelper.CreateTemporaryGitlabShellDir(t)
	defer cleanup()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	// By setting the last parameter to false, the post-receive API call will
	// send back {"reference_counter_increased": false}, indicating something went wrong
	// with the call

	c := testhelper.GitlabTestServerOptions{
		User:                        "",
		Password:                    "",
		SecretToken:                 secretToken,
		Changes:                     changes,
		GLID:                        glID,
		GLRepository:                glRepository,
		PostReceiveCounterDecreased: false,
		Protocol:                    "ssh",
	}
	serverURL, cleanup := testhelper.NewGitlabTestServer(t, c)
	defer cleanup()

	testhelper.WriteShellSecretFile(t, tempGitlabShellDir, secretToken)

	config.Config.GitlabShell.Dir = tempGitlabShellDir
	config.Config.Gitlab.URL = serverURL
	config.Config.Gitlab.SecretFile = filepath.Join(tempGitlabShellDir, ".gitlab_shell_secret")

	token := "abc123"

	gitlabAPI, err := gitalyhook.NewGitlabAPI(config.Config.Gitlab)
	require.NoError(t, err)

	socket, stop := runHookServiceServerWithAPI(t, token, gitlabAPI)
	defer stop()

	customHookOutputPath, cleanup := testhelper.WriteEnvToCustomHook(t, testRepoPath, "post-receive")
	defer cleanup()

	var stdout, stderr bytes.Buffer

	postReceiveHookPath, err := filepath.Abs("../../ruby/git-hooks/post-receive")
	require.NoError(t, err)

	testcases := []struct {
		desc    string
		primary bool
		verify  func(*testing.T, *exec.Cmd, *bytes.Buffer, *bytes.Buffer)
	}{
		{
			desc:    "Primary calls out to post_receive endpoint",
			primary: true,
			verify: func(t *testing.T, cmd *exec.Cmd, stdout, stderr *bytes.Buffer) {
				err = cmd.Run()
				code, ok := command.ExitStatus(err)
				require.True(t, ok, "expect exit status in %v", err)

				require.Equal(t, 1, code, "exit status")
				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())

				output := string(testhelper.MustReadFile(t, customHookOutputPath))
				require.Empty(t, output, "custom hook should not have run")
			},
		},
		{
			desc:    "Secondary does not call out to post_receive endpoint",
			primary: false,
			verify: func(t *testing.T, cmd *exec.Cmd, stdout, stderr *bytes.Buffer) {
				err = cmd.Run()
				require.NoError(t, err)

				require.Empty(t, stdout.String())
				require.Empty(t, stderr.String())

				output := string(testhelper.MustReadFile(t, customHookOutputPath))
				require.Empty(t, output, "custom hook should not have run")
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			transactionEnv, err := metadata.Transaction{
				ID:      1,
				Node:    "node",
				Primary: tc.primary,
			}.Env()
			require.NoError(t, err)

			env := testhelper.EnvForHooks(t, tempGitlabShellDir, socket, token, testRepo,
				testhelper.GlHookValues{
					GLID:       glID,
					GLUsername: glUsername,
					GLRepo:     glRepository,
					GLProtocol: glProtocol,
				},
				testhelper.ProxyValues{},
			)
			env = append(env, transactionEnv)

			cmd := exec.Command(postReceiveHookPath)
			cmd.Env = env
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			cmd.Stdin = bytes.NewBuffer([]byte(changes))
			cmd.Dir = testRepoPath

			tc.verify(t, cmd, &stdout, &stderr)
		})
	}
}

func TestHooksNotAllowed(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	secretToken := "secret token"
	glID := "key-1234"
	glUsername := "iamgitlab"
	glProtocol := "ssh"
	glRepository := "some_repo"
	changes := "oldhead newhead"

	tempGitlabShellDir, cleanup := testhelper.CreateTemporaryGitlabShellDir(t)
	defer cleanup()

	testRepo, testRepoPath, cleanupFn := testhelper.NewTestRepo(t)
	defer cleanupFn()

	c := testhelper.GitlabTestServerOptions{
		User:                        "",
		Password:                    "",
		SecretToken:                 secretToken,
		GLID:                        glID,
		GLRepository:                glRepository,
		Changes:                     changes,
		PostReceiveCounterDecreased: true,
		Protocol:                    "ssh",
	}
	serverURL, cleanup := testhelper.NewGitlabTestServer(t, c)
	defer cleanup()

	testhelper.WriteTemporaryGitlabShellConfigFile(t, tempGitlabShellDir, testhelper.GitlabShellConfig{GitlabURL: serverURL})
	testhelper.WriteShellSecretFile(t, tempGitlabShellDir, "the wrong token")

	config.Config.GitlabShell.Dir = tempGitlabShellDir
	config.Config.Gitlab.URL = serverURL
	config.Config.Gitlab.SecretFile = filepath.Join(tempGitlabShellDir, ".gitlab_shell_secret")

	customHookOutputPath, cleanup := testhelper.WriteEnvToCustomHook(t, testRepoPath, "post-receive")
	defer cleanup()

	token := "abc123"

	gitlabAPI, err := gitalyhook.NewGitlabAPI(config.Config.Gitlab)
	require.NoError(t, err)

	socket, stop := runHookServiceServerWithAPI(t, token, gitlabAPI)
	defer stop()

	var stderr, stdout bytes.Buffer

	preReceiveHookPath, err := filepath.Abs("../../ruby/git-hooks/pre-receive")
	require.NoError(t, err)
	cmd := exec.Command(preReceiveHookPath)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Stdin = strings.NewReader(changes)
	cmd.Env = testhelper.EnvForHooks(t, tempGitlabShellDir, socket, token, testRepo,
		testhelper.GlHookValues{
			GLID:       glID,
			GLUsername: glUsername,
			GLRepo:     glRepository,
			GLProtocol: glProtocol,
		},
		testhelper.ProxyValues{})
	cmd.Dir = testRepoPath

	require.Error(t, cmd.Run())
	require.Equal(t, "GitLab: 401 Unauthorized\n", stderr.String())
	require.Equal(t, "", stdout.String())

	output := string(testhelper.MustReadFile(t, customHookOutputPath))
	require.Empty(t, output, "custom hook should not have run")
}

func TestCheckOK(t *testing.T) {
	user, password := "user123", "password321"

	c := testhelper.GitlabTestServerOptions{
		User:                        user,
		Password:                    password,
		SecretToken:                 "",
		GLRepository:                "",
		Changes:                     "",
		PostReceiveCounterDecreased: false,
		Protocol:                    "ssh",
	}
	serverURL, cleanup := testhelper.NewGitlabTestServer(t, c)
	defer cleanup()

	tempDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(tempDir)
	}()

	gitlabShellDir := filepath.Join(tempDir, "gitlab-shell")
	require.NoError(t, os.MkdirAll(gitlabShellDir, 0755))

	testhelper.WriteShellSecretFile(t, gitlabShellDir, "the secret")
	configPath, cleanup := testhelper.WriteTemporaryGitalyConfigFile(t, tempDir, serverURL, user, password, path.Join(gitlabShellDir, ".gitlab_shell_secret"))
	defer cleanup()

	cmd := exec.Command(filepath.Join(config.Config.BinDir, "gitaly-hooks"), "check", configPath)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err = cmd.Run()
	require.NoError(t, err)
	require.Empty(t, stderr.String())

	output := stdout.String()
	require.Contains(t, output, "Checking GitLab API access: OK")
	require.Contains(t, output, "Redis reachable for GitLab: true")
}

func TestCheckBadCreds(t *testing.T) {
	defer func(cfg config.Cfg) {
		config.Config = cfg
	}(config.Config)

	user, password := "user123", "password321"

	c := testhelper.GitlabTestServerOptions{
		User:                        user,
		Password:                    password,
		SecretToken:                 "",
		GLRepository:                "",
		Changes:                     "",
		PostReceiveCounterDecreased: false,
		Protocol:                    "ssh",
		GitPushOptions:              nil,
	}
	serverURL, cleanup := testhelper.NewGitlabTestServer(t, c)
	defer cleanup()

	tempDir, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(tempDir)
	}()

	gitlabShellDir := filepath.Join(tempDir, "gitlab-shell")
	require.NoError(t, os.MkdirAll(gitlabShellDir, 0755))
	testhelper.WriteShellSecretFile(t, gitlabShellDir, "the secret")

	config.Config.Gitlab.URL = serverURL
	configPath, cleanup := testhelper.WriteTemporaryGitalyConfigFile(t, tempDir, serverURL, "wrong", password, path.Join(gitlabShellDir, ".gitlab_shell_secret"))
	defer cleanup()

	cmd := exec.Command(filepath.Join(config.Config.BinDir, "gitaly-hooks"), "check", configPath)

	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	require.Error(t, cmd.Run())
	require.Contains(t, stderr.String(), "Internal API error (401)")
	require.Equal(t, "Checking GitLab API access: FAIL\n", stdout.String())
}

func runHookServiceServer(t *testing.T, token string) (string, func()) {
	return runHookServiceServerWithAPI(t, token, gitalyhook.GitlabAPIStub)
}

func runHookServiceServerWithAPI(t *testing.T, token string, gitlabAPI gitalyhook.GitlabAPI) (string, func()) {
	server := testhelper.NewServerWithAuth(t, nil, nil, token)

	gitalypb.RegisterHookServiceServer(server.GrpcServer(), hook.NewServer(gitalyhook.NewManager(gitlabAPI, config.Config)))
	reflection.Register(server.GrpcServer())
	require.NoError(t, server.Start())

	return server.Socket(), server.Stop
}

func requireContainsOnce(t *testing.T, s string, contains string) {
	r := regexp.MustCompile(contains)
	matches := r.FindAllStringIndex(s, -1)
	require.Equal(t, 1, len(matches))
}
