package git2go

import (
	"context"
	"errors"
	"fmt"
	"io"

	"gitlab.com/gitlab-org/gitaly/internal/gitaly/config"
)

// CherryPickCommand contains parameters to perform a cherry pick.
type CherryPickCommand struct {
	// Repository is the path where to execute the cherry pick.
	Repository string `json:"repository"`
	// AuthorName is the author name for the resulting commit.
	AuthorName string `json:"author_name"`
	// AuthorMail is the author mail for the resulting commit.
	AuthorMail string `json:"author_mail"`
	// AuthorDate is the author date for the resulting commit.
	//AuthorDate time.Time `json:"author_date"`
	// Message is the message to be used for the resulting commit.
	Message string `json:"message"`
	// Commit is the commit that is to be picked
	Commit string `json:"commit"`
	// BranchName is the target branch where the picked commit should end up.
	BranchName string `json:"branch_name"`
	// StartBranchName is the name of the branch that should be picked.
	StartBranchName string `json:"start_branch_name"`
	// StartRepository is the path where to get the picked commit from. TODO??
	Repository string `json:"start_repository"`
	// DryRun is true when cherry pick should de attempted only.
	DryRun bool `json:"dry_run"`
}

// CherryPickResult contains results from a cherry pick.
type CherryPickResult struct {
	// CommitID is the object ID of the generated commit.
	CommitID string `json:"commit_id"`
}

// CherryPickCommandFromSerialized deserializes the request from its JSON representation encoded with base64.
func CherryPickCommandFromSerialized(serialized string) (CherryPickCommand, error) {
	var request CherryPickCommand
	if err := deserialize(serialized, &request); err != nil {
		return CherryPickCommand{}, err
	}

	if err := request.verify(); err != nil {
		return CherryPickCommand{}, fmt.Errorf("cherry-pick: %w: %s", ErrInvalidArgument, err.Error())
	}
	return request, nil
}

// SerializeTo serializes the result and writes it into the writer.
func (m CherryPickResult) SerializeTo(w io.Writer) error {
	return serializeTo(w, m)
}

// CherryPick performs a cherry pick via gitaly-git2go.
func (m CherryPickCommand) Run(ctx context.Context, cfg config.Cfg) (CherryPickResult, error) {
	if err := m.verify(); err != nil {
		return CherryPickResult{}, fmt.Errorf("cherry-pick: %w: %s", ErrInvalidArgument, err.Error())
	}

	serialized, err := serialize(m)
	if err != nil {
		return CherryPickResult{}, err
	}

	stdout, err := run(ctx, cfg, "cherry-pick", serialized)
	if err != nil {
		return CherryPickResult{}, err
	}

	var response CherryPickResult
	if err := deserialize(stdout, &response); err != nil {
		return CherryPickResult{}, err
	}

	return response, nil
}

func (m CherryPickCommand) verify() error {
	if m.Repository == "" {
		return errors.New("missing repository")
	}
	if m.AuthorName == "" {
		return errors.New("missing author name")
	}
	if m.AuthorMail == "" {
		return errors.New("missing author mail")
	}
	if m.Message == "" {
		return errors.New("missing message")
	}
	if m.Commit == "" {
		return errors.New("missing commit")
	}
	if m.BranchName == "" {
		return errors.New("missing branch name")
	}
	if m.StartBranchName == "" {
		return errors.New("missing start branch name")
	}
	if m.StartRepository == "" {
		return errors.New("missing start repository")
	}
	// TODO DryRun?
	return nil
}
