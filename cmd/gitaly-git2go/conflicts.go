// +build static,system_libgit2

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	git "github.com/libgit2/git2go/v30"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
)

type conflictsSubcommand struct {
	request string
}

func (cmd *conflictsSubcommand) Flags() *flag.FlagSet {
	flags := flag.NewFlagSet("conflicts", flag.ExitOnError)
	flags.StringVar(&cmd.request, "request", "", "git2go.ConflictsCommand")
	return flags
}

func indexEntryPath(entry *git.IndexEntry) string {
	if entry == nil {
		return ""
	}
	return entry.Path
}

func conflictContent(repo *git.Repository, conflict git.IndexConflict) (string, error) {
	var ancestor, our, their git.MergeFileInput

	for entry, input := range map[*git.IndexEntry]*git.MergeFileInput{
		conflict.Ancestor: &ancestor,
		conflict.Our:      &our,
		conflict.Their:    &their,
	} {
		if entry == nil {
			continue
		}

		blob, err := repo.LookupBlob(entry.Id)
		if err != nil {
			return "", fmt.Errorf("could not get conflicting blob: %w", err)
		}

		input.Path = entry.Path
		input.Mode = uint(entry.Mode)
		input.Contents = blob.Contents()
	}

	merge, err := git.MergeFile(ancestor, our, their, nil)
	if err != nil {
		return "", fmt.Errorf("could not compute conflicts: %w", err)
	}

	return string(merge.Contents), nil
}

// Run performs a merge and prints resulting conflicts to stdout.
func (cmd *conflictsSubcommand) Run() error {
	request, err := git2go.ConflictsCommandFromSerialized(cmd.request)
	if err != nil {
		return err
	}

	repo, err := git.OpenRepository(request.Repository)
	if err != nil {
		return fmt.Errorf("could not open repository: %w", err)
	}

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

	conflicts, err := index.ConflictIterator()
	if err != nil {
		return fmt.Errorf("could not get conflicts: %w", err)
	}

	var result git2go.ConflictsResult
	for {
		conflict, err := conflicts.Next()
		if err != nil {
			var gitError git.GitError
			if errors.As(err, &gitError) && gitError.Code != git.ErrIterOver {
				return err
			}
			break
		}

		content, err := conflictContent(repo, conflict)
		if err != nil {
			return err
		}

		result.Conflicts = append(result.Conflicts, git2go.Conflict{
			AncestorPath: indexEntryPath(conflict.Ancestor),
			OurPath:      indexEntryPath(conflict.Our),
			TheirPath:    indexEntryPath(conflict.Their),
			Content:      content,
		})
	}

	if err := result.SerializeTo(os.Stdout); err != nil {
		return err
	}

	return nil
}