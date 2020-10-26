// +build static,system_libgit2

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		return errors.New("Invalid branch") //nolint
	}

	startCommit, err := o.AsCommit()
	if err != nil {
		return fmt.Errorf("peeling %s as a commit: %w", o.Id(), err)
	}

	rootTree, err := startCommit.Tree()
	if err != nil {
		return fmt.Errorf("root tree from starting commit: %w", err)
	}

	smParentDir := filepath.Dir(request.Submodule)
	parentTree, err := parentTree(smParentDir, rootTree, repo)
	if err != nil {
		return fmt.Errorf("parent tree of submodule entry: %w", err)
	}

	smBasePath := filepath.Base(request.Submodule)
	smEntry := parentTree.EntryByName(smBasePath)
	if smEntry == nil {
		return errors.New("Invalid submodule path") //nolint
	}
	if smEntry.Type != git.ObjectCommit {
		return errors.New("Invalid submodule path") //nolint
	}

	if smEntry.Id.Equal(smCommitOID) {
		return fmt.Errorf(
			"The submodule %s is already at %s",
			request.Submodule, request.CommitSHA,
		) //nolint
	}

	parentTreeBuilder, err := repo.TreeBuilderFromTree(parentTree)
	if err != nil {
		return fmt.Errorf("tree builder for submodule entry: %w", err)
	}

	if err := parentTreeBuilder.Insert(
		smBasePath,
		smCommitOID,
		git.FilemodeCommit,
	); err != nil {
		return fmt.Errorf("inserting submodule entry: %w", err)
	}

	newTreeOID, err := parentTreeBuilder.Write()
	if err != nil {
		return fmt.Errorf("writing tree entry for submodule: %w", err)
	}

	newRootTreeOID, err := updateTreeAncestors(repo, rootTree, smParentDir, smBasePath, newTreeOID)
	if err != nil {
		return fmt.Errorf("updating submodule tree ancestors: %w", err)
	}

	newTree, err := repo.LookupTree(newRootTreeOID)
	if err != nil {
		return fmt.Errorf("looking up new submodule entry root tree: %w", err)
	}
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

func parentTree(smParentDir string, rootTree *git.Tree, repo *git.Repository) (*git.Tree, error) {
	if smParentDir == "" || smParentDir == "." {
		return rootTree, nil
	}
	parentEntry, err := rootTree.EntryByPath(smParentDir)
	if err != nil {
		return nil, err
	}

	parentTree, err := repo.LookupTree(parentEntry.Id)
	if err != nil {
		return nil, err
	}

	return parentTree, nil
}

// updateTreeAncestors will traverse up the tree ancestors and update the entry
// for the new tree OID.
func updateTreeAncestors(repo *git.Repository, rootTree *git.Tree, smParentDir, smBasePath string, newTreeOID *git.Oid) (_ *git.Oid, err error) {
	if smParentDir == "." {
		return newTreeOID, nil // root tree doesn't require traversal
	}
	var (
		tPath = smParentDir
		tName = smBasePath
		tOID  = newTreeOID
	)
	defer func() {
		// decorate any returned errors with details about which tree
		// entry failed
		if err != nil {
			err = fmt.Errorf(
				"tree entry for path:%q, name: %q, oid: %q: %w",
				tPath, tName, tOID, err,
			)
		}
	}()
	for {
		tName = filepath.Base(tPath)
		tPath = filepath.Dir(tPath)

		var (
			tb  *git.TreeBuilder
			err error
		)
		if tPath == "." {
			tb, err = repo.TreeBuilderFromTree(rootTree)
			if err != nil {
				return nil, err
			}
		} else {
			e, err := rootTree.EntryByPath(tPath)
			if err != nil {
				return nil, err
			}

			te, err := repo.LookupTree(e.Id)
			if err != nil {
				return nil, err
			}

			tb, err = repo.TreeBuilderFromTree(te)
			if err != nil {
				return nil, err
			}
		}

		if err := tb.Insert(tName, tOID, git.FilemodeTree); err != nil {
			return nil, err
		}

		tOID, err = tb.Write()
		if err != nil {
			return nil, err
		}

		if tPath == "." {
			break // processed the root tree, no more to traverse
		}
	}

	return tOID, nil
}
