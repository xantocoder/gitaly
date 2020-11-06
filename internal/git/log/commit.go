package log

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"gitlab.com/gitlab-org/gitaly/internal/git"
	"gitlab.com/gitlab-org/gitaly/internal/git/catfile"
	"gitlab.com/gitlab-org/gitaly/internal/helper"
	"gitlab.com/gitlab-org/gitaly/proto/go/gitalypb"
)

// GetCommit tries to resolve revision to a Git commit. Returns nil if
// no object is found at revision.
func GetCommit(ctx context.Context, repo *gitalypb.Repository, revision string) (*gitalypb.GitCommit, error) {
	c, err := catfile.New(ctx, repo)
	if err != nil {
		return nil, err
	}

	return GetCommitCatfile(c, revision)
}

// GetCommitCatfile looks up a commit by revision using an existing *catfile.Batch instance.
func GetCommitCatfile(c *catfile.Batch, revision string) (*gitalypb.GitCommit, error) {
	obj, err := c.Commit(revision + "^{commit}")
	if err != nil {
		return nil, err
	}

	return parseRawCommit(obj.Reader, &obj.ObjectInfo)
}

// GetCommitMessage looks up a commit message and returns it in its entirety.
func GetCommitMessage(c *catfile.Batch, repo *gitalypb.Repository, revision string) ([]byte, error) {
	obj, err := c.Commit(revision + "^{commit}")
	if err != nil {
		return nil, err
	}

	_, body, err := splitRawCommit(obj.Reader)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func parseRawCommit(r io.Reader, info *catfile.ObjectInfo) (*gitalypb.GitCommit, error) {
	header, body, err := splitRawCommit(r)
	if err != nil {
		return nil, err
	}
	return buildCommit(header, body, info)
}

func splitRawCommit(r io.Reader) ([]byte, []byte, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}

	split := bytes.SplitN(raw, []byte("\n\n"), 2)

	header := split[0]
	var body []byte
	if len(split) == 2 {
		body = split[1]
	}

	return header, body, nil
}

func buildCommit(header, body []byte, info *catfile.ObjectInfo) (*gitalypb.GitCommit, error) {
	commit := &gitalypb.GitCommit{
		Id:       info.Oid,
		BodySize: int64(len(body)),
		Body:     body,
		Subject:  subjectFromBody(body),
	}

	if max := helper.MaxCommitOrTagMessageSize; len(body) > max {
		commit.Body = commit.Body[:max]
	}

	scanner := bufio.NewScanner(bytes.NewReader(header))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == ' ' {
			continue
		}

		headerSplit := strings.SplitN(line, " ", 2)
		if len(headerSplit) != 2 {
			continue
		}

		switch headerSplit[0] {
		case "parent":
			commit.ParentIds = append(commit.ParentIds, headerSplit[1])
		case "author":
			commit.Author = parseCommitAuthor(headerSplit[1])
		case "committer":
			commit.Committer = parseCommitAuthor(headerSplit[1])
		case "gpgsig":
			commit.SignatureType = detectSignatureType(headerSplit[1])
		case "tree":
			commit.TreeId = headerSplit[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commit, nil
}

const maxUnixCommitDate = 1 << 53

func parseCommitAuthor(line string) *gitalypb.CommitAuthor {
	author := &gitalypb.CommitAuthor{}

	splitName := strings.SplitN(line, "<", 2)
	author.Name = []byte(strings.TrimSuffix(splitName[0], " "))

	if len(splitName) < 2 {
		return author
	}

	line = splitName[1]
	splitEmail := strings.SplitN(line, ">", 2)
	if len(splitEmail) < 2 {
		return author
	}

	author.Email = []byte(splitEmail[0])

	secSplit := strings.Fields(splitEmail[1])
	if len(secSplit) < 1 {
		return author
	}

	sec, err := strconv.ParseInt(secSplit[0], 10, 64)
	if err != nil || sec > maxUnixCommitDate || sec < 0 {
		sec = git.FallbackTimeValue.Unix()
	}

	author.Date = &timestamp.Timestamp{Seconds: sec}

	if len(secSplit) == 2 {
		author.Timezone = []byte(secSplit[1])
	}

	return author
}

func subjectFromBody(body []byte) []byte {
	return bytes.TrimRight(bytes.SplitN(body, []byte("\n"), 2)[0], "\r\n")
}

func detectSignatureType(line string) gitalypb.SignatureType {
	switch strings.TrimSuffix(line, "\n") {
	case "-----BEGIN SIGNED MESSAGE-----":
		return gitalypb.SignatureType_X509
	case "-----BEGIN PGP MESSAGE-----":
		return gitalypb.SignatureType_PGP
	case "-----BEGIN PGP SIGNATURE-----":
		return gitalypb.SignatureType_PGP
	default:
		return gitalypb.SignatureType_NONE
	}
}
