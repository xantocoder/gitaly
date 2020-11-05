package git

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/git/repository"
	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/internal/helper/text"
)

// InvalidObjectError is returned when trying to get an object id that is invalid or does not exist.
type InvalidObjectError string

func (err InvalidObjectError) Error() string { return fmt.Sprintf("invalid object %q", string(err)) }

func errorWithStderr(err error, stderr []byte) error {
	return fmt.Errorf("%w, stderr: %q", err, stderr)
}

var (
	ErrReferenceNotFound = errors.New("reference not found")
	// ErrAlreadyExists represents an error when the resource is already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound represents an error when the resource can't be found.
	ErrNotFound = errors.New("not found")
)

// FetchOptsTags controls what tags needs to be imported on fetch.
type FetchOptsTags string

func (t FetchOptsTags) String() string {
	return string(t)
}

var (
	// FetchOptsTagsDefault enables importing of tags only on fetched branches.
	FetchOptsTagsDefault = FetchOptsTags("")
	// FetchOptsTagsAll enables importing of every tag from the remote repository.
	FetchOptsTagsAll = FetchOptsTags("--tags")
	// FetchOptsTagsNone disables importing of tags from the remote repository.
	FetchOptsTagsNone = FetchOptsTags("--no-tags")
)

// FetchOpts is used to configure invocation of the 'FetchRemote' command.
type FetchOpts struct {
	// Env is a list of env vars to pass to the cmd.
	Env []string
	// Global is a list of global flags to use with 'git' command.
	Global []Option
	// Prune if set fetch removes any remote-tracking references that no longer exist on the remote.
	// https://git-scm.com/docs/git-fetch#Documentation/git-fetch.txt---prune
	Prune bool
	// Force if set fetch overrides local references with values from remote that's
	// doesn't have the previous commit as an ancestor.
	// https://git-scm.com/docs/git-fetch#Documentation/git-fetch.txt---force
	Force bool
	// Tags controls whether tags will be fetched as part of the remote or not.
	// https://git-scm.com/docs/git-fetch#Documentation/git-fetch.txt---tags
	// https://git-scm.com/docs/git-fetch#Documentation/git-fetch.txt---no-tags
	Tags FetchOptsTags
	// Stderr if set it would be used to redirect stderr stream into it.
	Stderr io.Writer
}

func (opts FetchOpts) buildFlags() []Option {
	flags := []Option{Flag{Name: "--quiet"}}

	if opts.Prune {
		flags = append(flags, Flag{Name: "--prune"})
	}

	if opts.Force {
		flags = append(flags, Flag{Name: "--force"})
	}

	if opts.Tags != FetchOptsTagsDefault {
		flags = append(flags, Flag{Name: opts.Tags.String()})
	}

	return flags
}

// Repository represents a Git repository.
type Repository interface {
	// ResolveRef resolves the given refish to its object ID. This uses the
	// typical DWIM mechanism of Git to resolve the reference. See
	// gitrevisions(1) for accepted syntax. This will not verify whether the
	// object ID exists. To do so, you can peel the reference to a given
	// object type, e.g. by passing `refs/heads/master^{commit}`.
	ResolveRefish(ctx context.Context, ref string) (string, error)

	// ContainsRef checks if a ref in the repository exists. This will not
	// verify whether the target object exists. To do so, you can peel the
	// reference to a given object type, e.g. by passing
	// `refs/heads/master^{commit}`.
	ContainsRef(ctx context.Context, ref string) (bool, error)

	// GetReference looks up and returns the given reference. Returns a
	// ReferenceNotFound error if the reference was not found.
	GetReference(ctx context.Context, ref string) (Reference, error)

	// GetReferences returns references matching the given pattern.
	GetReferences(ctx context.Context, pattern string) ([]Reference, error)

	// GetBranch looks up and returns the given branch. Returns a
	// ErrReferenceNotFound if it wasn't found.
	GetBranch(ctx context.Context, branch string) (Reference, error)

	// GetBranches returns all branches.
	GetBranches(ctx context.Context) ([]Reference, error)

	// UpdateRef updates reference from oldrev to newrev. If oldrev is a
	// non-empty string, the update will fail it the reference is not
	// currently at that revision. If newrev is the zero OID, the reference
	// will be deleted. If oldrev is the zero OID, the reference will
	// created.
	UpdateRef(ctx context.Context, reference, newrev, oldrev string) error

	// WriteBlob writes a blob to the repository's object database and
	// returns its object ID. Path is used by git to decide which filters to
	// run on the content.
	WriteBlob(ctx context.Context, path string, content io.Reader) (string, error)

	// ReadObject reads an object from the repository's object database. InvalidObjectError
	// is returned if the oid does not refer to a valid object.
	ReadObject(ctx context.Context, oid string) ([]byte, error)

	// FetchRemote fetches changes from the specified remote.
	FetchRemote(ctx context.Context, remoteName string, opts FetchOpts) error

	// Config returns executor of the 'config' sub-command.
	Config() Config

	// Remote returns executor of the 'remote' sub-command.
	Remote() Remote
}

// ErrUnimplemented indicates the repository abstraction does not yet implement
// a specific behavior
var ErrUnimplemented = errors.New("behavior not implemented yet")

// UnimplementedRepo satisfies the Repository interface to reduce friction in
// writing new Repository implementations
type UnimplementedRepo struct{}

func (UnimplementedRepo) ResolveRefish(ctx context.Context, ref string) (string, error) {
	return "", ErrUnimplemented
}

func (UnimplementedRepo) ContainsRef(ctx context.Context, ref string) (bool, error) {
	return false, ErrUnimplemented
}

func (UnimplementedRepo) GetReference(ctx context.Context, ref string) (Reference, error) {
	return Reference{}, ErrUnimplemented
}

func (UnimplementedRepo) GetReferences(ctx context.Context, pattern string) ([]Reference, error) {
	return nil, ErrUnimplemented
}

func (UnimplementedRepo) GetBranch(ctx context.Context, branch string) (Reference, error) {
	return Reference{}, ErrUnimplemented
}

func (UnimplementedRepo) GetBranches(ctx context.Context) ([]Reference, error) {
	return nil, ErrUnimplemented
}

func (UnimplementedRepo) UpdateRef(ctx context.Context, reference, newrev, oldrev string) error {
	return ErrUnimplemented
}

func (UnimplementedRepo) WriteBlob(context.Context, string, io.Reader) (string, error) {
	return "", ErrUnimplemented
}

func (UnimplementedRepo) ReadObject(context.Context, string) ([]byte, error) {
	return nil, ErrUnimplemented
}

func (UnimplementedRepo) FetchRemote(context.Context, string, FetchOpts) error {
	return ErrUnimplemented
}

func (UnimplementedRepo) Remote() Remote {
	return UnimplementedRemote{}
}

func (UnimplementedRepo) Config() Config {
	return UnimplementedConfig{}
}

var _ Repository = UnimplementedRepo{} // compile time assertion

// localRepository represents a local Git repository.
type localRepository struct {
	repo repository.GitRepo
}

// NewRepository creates a new Repository from its protobuf representation.
func NewRepository(repo repository.GitRepo) Repository {
	return &localRepository{
		repo: repo,
	}
}

// command creates a Git Command with the given args and Repository, executed
// in the Repository. It validates the arguments in the command before
// executing.
func (repo *localRepository) command(ctx context.Context, globals []Option, cmd SubCmd, opts ...CmdOpt) (*command.Command, error) {
	return SafeCmd(ctx, repo.repo, globals, cmd, opts...)
}

func (repo *localRepository) WriteBlob(ctx context.Context, path string, content io.Reader) (string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd, err := repo.command(ctx, nil,
		SubCmd{
			Name: "hash-object",
			Flags: []Option{
				ValueFlag{Name: "--path", Value: path},
				Flag{Name: "--stdin"}, Flag{Name: "-w"},
			},
		},
		WithStdin(content),
		WithStdout(stdout),
		WithStderr(stderr),
	)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return "", errorWithStderr(err, stderr.Bytes())
	}

	return text.ChompBytes(stdout.Bytes()), nil
}

func (repo *localRepository) ReadObject(ctx context.Context, oid string) ([]byte, error) {
	const msgInvalidObject = "fatal: Not a valid object name "

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd, err := repo.command(ctx, nil,
		SubCmd{
			Name:  "cat-file",
			Flags: []Option{Flag{"-p"}},
			Args:  []string{oid},
		},
		WithStdout(stdout),
		WithStderr(stderr),
	)
	if err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		msg := text.ChompBytes(stderr.Bytes())
		if strings.HasPrefix(msg, msgInvalidObject) {
			return nil, InvalidObjectError(strings.TrimPrefix(msg, msgInvalidObject))
		}

		return nil, errorWithStderr(err, stderr.Bytes())
	}

	return stdout.Bytes(), nil
}

func (repo *localRepository) ResolveRefish(ctx context.Context, refish string) (string, error) {
	if refish == "" {
		return "", errors.New("repository cannot contain empty reference name")
	}

	cmd, err := repo.command(ctx, nil, SubCmd{
		Name:  "rev-parse",
		Flags: []Option{Flag{Name: "--verify"}},
		Args:  []string{refish},
	})
	if err != nil {
		return "", err
	}

	var stdout bytes.Buffer
	io.Copy(&stdout, cmd)

	if err := cmd.Wait(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return "", ErrReferenceNotFound
		}
		return "", err
	}

	oid := strings.TrimSpace(stdout.String())
	if len(oid) != 40 {
		return "", fmt.Errorf("unsupported object hash %q", oid)
	}

	return oid, nil
}

func (repo *localRepository) ContainsRef(ctx context.Context, ref string) (bool, error) {
	if _, err := repo.ResolveRefish(ctx, ref); err != nil {
		if errors.Is(err, ErrReferenceNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (repo *localRepository) GetReference(ctx context.Context, ref string) (Reference, error) {
	refs, err := repo.GetReferences(ctx, ref)
	if err != nil {
		return Reference{}, err
	}

	if len(refs) == 0 {
		return Reference{}, ErrReferenceNotFound
	}

	return refs[0], nil
}

func (repo *localRepository) GetReferences(ctx context.Context, pattern string) ([]Reference, error) {
	var args []string
	if pattern != "" {
		args = []string{pattern}
	}

	cmd, err := repo.command(ctx, nil, SubCmd{
		Name:  "for-each-ref",
		Flags: []Option{Flag{Name: "--format=%(refname)%00%(objectname)%00%(symref)"}},
		Args:  args,
	})
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(cmd)

	var refs []Reference
	for scanner.Scan() {
		line := bytes.SplitN(scanner.Bytes(), []byte{0}, 3)
		if len(line) != 3 {
			return nil, errors.New("unexpected reference format")
		}

		if len(line[2]) == 0 {
			refs = append(refs, NewReference(string(line[0]), string(line[1])))
		} else {
			refs = append(refs, NewSymbolicReference(string(line[0]), string(line[1])))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading standard input: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	return refs, nil
}

func (repo *localRepository) GetBranch(ctx context.Context, branch string) (Reference, error) {
	if strings.HasPrefix(branch, "refs/heads/") {
		return repo.GetReference(ctx, branch)
	}

	if strings.HasPrefix(branch, "heads/") {
		branch = strings.TrimPrefix(branch, "heads/")
	}
	return repo.GetReference(ctx, "refs/heads/"+branch)
}

func (repo *localRepository) GetBranches(ctx context.Context) ([]Reference, error) {
	return repo.GetReferences(ctx, "refs/heads/")
}

func (repo *localRepository) UpdateRef(ctx context.Context, reference, newrev, oldrev string) error {
	cmd, err := repo.command(ctx, nil,
		SubCmd{
			Name:  "update-ref",
			Flags: []Option{Flag{Name: "-z"}, Flag{Name: "--stdin"}},
		},
		WithStdin(strings.NewReader(fmt.Sprintf("update %s\x00%s\x00%s\x00", reference, newrev, oldrev))),
		WithRefTxHook(ctx, helper.ProtoRepoFromRepo(repo.repo), config.Config),
	)
	if err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("UpdateRef: failed updating reference %q from %q to %q: %v", reference, newrev, oldrev, err)
	}

	return nil
}

func (repo *localRepository) FetchRemote(ctx context.Context, remoteName string, opts FetchOpts) error {
	if err := validateNotBlank(remoteName, "remoteName"); err != nil {
		return err
	}

	cmd, err := SafeCmdWithEnv(ctx, opts.Env, repo.repo, opts.Global,
		SubCmd{
			Name:  "fetch",
			Flags: opts.buildFlags(),
			Args:  []string{remoteName},
		},
		WithStderr(opts.Stderr),
	)
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (repo *localRepository) Config() Config {
	return RepositoryConfig{repo: repo.repo}
}

func (repo *localRepository) Remote() Remote {
	return RepositoryRemote{repo: repo.repo}
}

func isExitWithCode(err error, code int) bool {
	actual, ok := command.ExitStatus(err)
	if !ok {
		return false
	}

	return code == actual
}

func validateNotBlank(val, name string) error {
	if strings.TrimSpace(val) == "" {
		return fmt.Errorf("%w: %q is blank or empty", ErrInvalidArg, name)
	}
	return nil
}
