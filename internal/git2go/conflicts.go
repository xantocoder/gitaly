package git2go

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
)

// ConflictsCommand contains parameters to perform a merge and return its conflicts.
type ConflictsCommand struct {
	// Repository is the path to execute merge in.
	Repository string `json:"repository"`
	// Ours is the commit that is to be merged into theirs.
	Ours string `json:"ours"`
	// Theirs is the commit into which ours is to be merged.
	Theirs string `json:"theirs"`
}

// Conflict represents a merge conflict for a single file.
type Conflict struct {
	// AncestorPath is the path of the ancestor.
	AncestorPath string `json:"ancestor_path"`
	// OurPath is the path of ours.
	OurPath string `json:"our_path"`
	// TheirPath is the path of theirs.
	TheirPath string `json:"their_path"`
	// Content contains the conflicting merge results.
	Content string `json:"content"`
}

// ConflictsResult contains all conflicts resulting from a merge.
type ConflictsResult struct {
	// Conflicts
	Conflicts []Conflict `json:"conflicts"`
}

// ConflictsCommandFromSerialized constructs a ConflictsCommand from its serialized representation.
func ConflictsCommandFromSerialized(serialized string) (ConflictsCommand, error) {
	var request ConflictsCommand
	if err := deserialize(serialized, &request); err != nil {
		return ConflictsCommand{}, err
	}

	if err := request.verify(); err != nil {
		return ConflictsCommand{}, fmt.Errorf("conflicts: %w: %s", ErrInvalidArgument, err.Error())
	}

	return request, nil
}

// Serialize serializes the conflicts result.
func (m ConflictsResult) SerializeTo(writer io.Writer) error {
	return serializeTo(writer, m)
}

// Run performs a merge via gitaly-git2go and returns all resulting conflicts.
func (c ConflictsCommand) Run(ctx context.Context, cfg config.Cfg) (ConflictsResult, error) {
	if err := c.verify(); err != nil {
		return ConflictsResult{}, fmt.Errorf("conflicts: %w: %s", ErrInvalidArgument, err.Error())
	}

	serialized, err := serialize(c)
	if err != nil {
		return ConflictsResult{}, err
	}

	stdout, err := run(ctx, cfg, "conflicts", serialized)
	if err != nil {
		return ConflictsResult{}, err
	}

	var response ConflictsResult
	if err := deserialize(stdout, &response); err != nil {
		return ConflictsResult{}, err
	}

	return response, nil
}

func (c ConflictsCommand) verify() error {
	if c.Repository == "" {
		return errors.New("missing repository")
	}
	if c.Ours == "" {
		return errors.New("missing ours")
	}
	if c.Theirs == "" {
		return errors.New("missing theirs")
	}
	return nil
}