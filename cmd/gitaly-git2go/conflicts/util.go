// +build static,system_libgit2

package conflicts

import (
	"errors"
	"fmt"
	"os"

	git "github.com/libgit2/git2go/v30"
	"gitlab.com/gitlab-org/gitaly/internal/git2go"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetIndexConflicts(repo *git.Repository, index *git.Index) (git2go.ConflictsResult, error) {
	var result git2go.ConflictsResult

	conflicts, err := index.ConflictIterator()
	if err != nil {
		return result, fmt.Errorf("could not get conflicts: %w", err)
	}

	for {
		conflict, err := conflicts.Next()
		if err != nil {
			var gitError git.GitError
			if errors.As(err, &gitError) && gitError.Code != git.ErrIterOver {
				return result, err
			}
			break
		}

		content, err := conflictContent(repo, conflict)
		if err != nil {
			if status, ok := status.FromError(err); ok {
				return result, conflictError(status.Code(), status.Message())
			}
			return result, err
		}

		result.Conflicts = append(result.Conflicts, git2go.Conflict{
			Ancestor: conflictEntryFromIndex(conflict.Ancestor),
			Our:      conflictEntryFromIndex(conflict.Our),
			Their:    conflictEntryFromIndex(conflict.Their),
			Content:  content,
		})
	}

	return result, nil
}

func conflictError(code codes.Code, message string) error {
	result := git2go.ConflictsResult{
		Error: git2go.ConflictError{
			Code:    code,
			Message: message,
		},
	}

	if err := result.SerializeTo(os.Stdout); err != nil {
		return err
	}

	return nil
}

func conflictContent(repo *git.Repository, conflict git.IndexConflict) ([]byte, error) {
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
			return nil, helper.ErrPreconditionFailedf("could not get conflicting blob: %w", err)
		}

		input.Path = entry.Path
		input.Mode = uint(entry.Mode)
		input.Contents = blob.Contents()
	}

	merge, err := git.MergeFile(ancestor, our, their, nil)
	if err != nil {
		return nil, fmt.Errorf("could not compute conflicts: %w", err)
	}

	return merge.Contents, nil
}

func conflictEntryFromIndex(entry *git.IndexEntry) git2go.ConflictEntry {
	if entry == nil {
		return git2go.ConflictEntry{}
	}
	return git2go.ConflictEntry{
		Path: entry.Path,
		Mode: int32(entry.Mode),
	}
}
