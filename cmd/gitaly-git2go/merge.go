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
	"gitlab.com/gitlab-org/gitaly/cmd/gitaly-git2go/conflicts"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
)

type mergeSubcommand struct {
	request string
}

func (cmd *mergeSubcommand) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("merge", flag.ExitOnError)
	flags.StringVar(&cmd.request, "request", "", "git2go.MergeCommand")
	return flags
}

func (cmd *mergeSubcommand) Run(context.Context, io.Reader, io.Writer) error {
	request, err := git2go.MergeCommandFromSerialized(cmd.request)
	if err != nil {
		return err
	}

	if request.AuthorDate.IsZero() {
		request.AuthorDate = time.Now()
	}

	repo, err := git.OpenRepository(request.Repository)
	if err != nil {
		return fmt.Errorf("could not open repository: %w", err)
	}
	defer repo.Free()

	ours, err := lookupCommit(repo, request.Ours)
	if err != nil {
		return fmt.Errorf("could not lookup commit %q: %w", request.Ours, err)
	}

	theirs, err := lookupCommit(repo, request.Theirs)
	if err != nil {
		return fmt.Errorf("could not lookup commit %q: %w", request.Theirs, err)
	}

	index, err := repo.MergeCommits(ours, theirs, nil)
	if err != nil {
		return fmt.Errorf("could not merge commits: %w", err)
	}
	defer index.Free()

	if err := processConflicts(repo, index, request.AllowConflicts); err != nil {
		return err
	}

	tree, err := index.WriteTreeTo(repo)
	if err != nil {
		return fmt.Errorf("could not write tree: %w", err)
	}

	committer := git.Signature(git2go.NewSignature(request.AuthorName, request.AuthorMail, request.AuthorDate))
	commit, err := repo.CreateCommitFromIds("", &committer, &committer, request.Message, tree, ours.Id(), theirs.Id())
	if err != nil {
		return fmt.Errorf("could not create merge commit: %w", err)
	}

	response := git2go.MergeResult{
		CommitID: commit.String(),
	}

	if err := response.SerializeTo(os.Stdout); err != nil {
		return err
	}

	return nil
}

func processConflicts(repo *git.Repository, index *git.Index, allowConflicts bool) error {
	if !index.HasConflicts() {
		return nil
	}

	if !allowConflicts {
		return errors.New("could not auto-merge due to conflicts")
	}

	result, err := conflicts.GetIndexConflicts(repo, index)
	if err != nil {
		return fmt.Errorf("could not get conflicts content: %w", err)
	}

	odb, err := repo.Odb()
	if err != nil {
		return fmt.Errorf("could not get odb for adding conflicts content: %w", err)
	}

	for _, conflict := range result.Conflicts {
		oid, err := odb.Write(conflict.Content, git.ObjectBlob)
		if err != nil {
			return fmt.Errorf("could not write conflicts content to repo: %w", err)
		}

		entry := &git.IndexEntry{Path: conflict.Ancestor.Path, Id: oid, Mode: git.FilemodeBlob}
		if err := index.Add(entry); err != nil {
			return fmt.Errorf("could not add conflicts content to index: %w", err)
		}

		if err := index.RemoveConflict(conflict.Ancestor.Path); err != nil {
			return fmt.Errorf("could not resolve conflicts: %w", err)
		}
	}

	return nil
}
