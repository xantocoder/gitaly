// +build static,system_libgit2

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	git "github.com/libgit2/git2go/v30"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
)

type submoduleSubcommand struct {
	request string
}

func (cmd *submoduleSubcommand) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("submodule", flag.ExitOnError)
	flags.StringVar(&cmd.request, "request", "", "git2go.SubmoduleCommand")
	return flags
}

func (cmd *submoduleSubcommand) Run(context.Context, io.Reader, io.Writer) error {
	request, err := git2go.SubmoduleCommandFromSerialized(cmd.request)
	if err != nil {
		return fmt.Errorf("deserializing submodule command request: %w", err)
	}

	if request.AuthorDate.IsZero() {
		request.AuthorDate = time.Now()
	}

	smCommitOID, err := git.NewOid(request.CommitSHA)
	if err != nil {
		return fmt.Errorf("converting %s to OID: %w", request.CommitSHA, err)
	}

	repo, err := git.OpenRepository(request.Repository)
	if err != nil {
		return fmt.Errorf("could not open repository: %w", err)
	}
	defer repo.Free()

	fullBranchRefName := "refs/heads/" + request.Branch
	o, err := repo.RevparseSingle(fullBranchRefName)
	if err != nil {
		return fmt.Errorf("%s: %w", git2go.LegacyErrPrefixInvalidBranch, err) //nolint
	}
	defer o.Free()

	startCommit, err := o.AsCommit()
	if err != nil {
		return fmt.Errorf("peeling %s as a commit: %w", o.Id(), err)
	}
	defer startCommit.Free()

	rootTree, err := startCommit.Tree()
	if err != nil {
		return fmt.Errorf("root tree from starting commit: %w", err)
	}
	defer rootTree.Free()

	index, err := git.NewIndex()
	if err != nil {
		return fmt.Errorf("creating new index: %w", err)
	}
	defer index.Free()

	if err := index.ReadTree(rootTree); err != nil {
		return fmt.Errorf("reading root tree into index: %w", err)
	}

	smEntry, err := index.EntryByPath(request.Submodule, 0)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			git2go.LegacyErrPrefixInvalidSubmodulePath, err,
		) //nolint
	}

	if smEntry.Id.Cmp(smCommitOID) == 0 {
		return fmt.Errorf(
			"The submodule %s is already at %s",
			request.Submodule, request.CommitSHA,
		)
	}

	if smEntry.Mode != git.FilemodeCommit {
		return fmt.Errorf(
			"%s: %w",
			git2go.LegacyErrPrefixInvalidSubmodulePath, err,
		) //nolint
	}

	newEntry := *smEntry      // copy by value
	newEntry.Id = smCommitOID // assign new commit SHA
	if err := index.Add(&newEntry); err != nil {
		return fmt.Errorf("add new submodule entry to index: %w", err)
	}

	newRootTreeOID, err := index.WriteTreeTo(repo)
	if err != nil {
		return fmt.Errorf("write index to repo: %w", err)
	}

	newTree, err := repo.LookupTree(newRootTreeOID)
	if err != nil {
		return fmt.Errorf("looking up new submodule entry root tree: %w", err)
	}
	defer newTree.Free()

	committer := git.Signature(
		git2go.NewSignature(
			request.AuthorName,
			request.AuthorMail,
			request.AuthorDate,
		),
	)
	newCommitOID, err := repo.CreateCommit(
		"", // caller should update branch with hooks
		&committer,
		&committer,
		request.Message,
		newTree,
		startCommit,
	)
	if err != nil {
		return errors.New("Failed to create commit") //nolint
	}

	return git2go.SubmoduleResult{
		CommitID: newCommitOID.String(),
	}.SerializeTo(os.Stdout)
}
