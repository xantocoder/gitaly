// +build static,system_libgit2

package conflicts

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	git "github.com/libgit2/git2go/v30"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
	"google.golang.org/grpc/codes"
)

// Subcommand contains params to performs conflicts calculation from main
type Subcommand struct {
	request string
}

func (cmd *Subcommand) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("conflicts", flag.ExitOnError)
	flags.StringVar(&cmd.request, "request", "", "git2go.ConflictsCommand")
	return flags
}

// Run performs a merge and prints resulting conflicts to stdout.
func (cmd *Subcommand) Run(context.Context, io.Reader, io.Writer) error {
	request, err := git2go.ConflictsCommandFromSerialized(cmd.request)
	if err != nil {
		return err
	}

	repo, err := git.OpenRepository(request.Repository)
	if err != nil {
		return fmt.Errorf("could not open repository: %w", err)
	}

	oursOid, err := git.NewOid(request.Ours)
	if err != nil {
		return err
	}

	ours, err := repo.LookupCommit(oursOid)
	if err != nil {
		return err
	}

	theirsOid, err := git.NewOid(request.Theirs)
	if err != nil {
		return err
	}

	theirs, err := repo.LookupCommit(theirsOid)
	if err != nil {
		return err
	}

	index, err := repo.MergeCommits(ours, theirs, nil)
	if err != nil {
		return conflictError(codes.FailedPrecondition, fmt.Sprintf("could not merge commits: %v", err))
	}

	result, err := GetIndexConflicts(repo, index)
	if err != nil {
		return err
	}

	if err := result.SerializeTo(os.Stdout); err != nil {
		return err
	}

	return nil
}
