package git

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"gitlab.com/gitlab-org/gitaly/internal/command"
	"gitlab.com/gitlab-org/gitaly/internal/helper"

	pb "gitlab.com/gitlab-org/gitaly-proto/go"
)

// Command creates a git.Command with the given args
func Command(ctx context.Context, repo *pb.Repository, args ...string) (*command.Command, error) {
	repoPath, err := helper.GetRepoPath(repo)
	if err != nil {
		return nil, err
	}
	args = append([]string{"--git-dir", repoPath}, args...)

	var env []string
	if dir := repo.GetGitObjectDirectory(); dir != "" {
		env = append(env, fmt.Sprintf("GIT_OBJECT_DIRECTORY=%s", path.Join(repoPath, dir)))
	}

	if dirs := repo.GetGitAlternateObjectDirectories(); len(dirs) > 0 {
		var dirsList []string

		for _, dir := range dirs {
			dirsList = append(dirsList, path.Join(repoPath, dir))
		}

		env = append(env, fmt.Sprintf("GIT_ALTERNATE_OBJECT_DIRECTORIES=%s", strings.Join(dirsList, ":")))
	}

	return command.New(ctx, exec.Command(command.GitPath(), args...), nil, nil, nil, env...)
}
