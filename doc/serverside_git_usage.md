## Server side Git usage

Gitaly uses three implementations to read and write to Git repositories:
1. `git(1)` - The same Git used by clients all over the world
1. [LibGit2](https://github.com/libgit2/libgit2) - a linkable library used through Rugged and Git2Go
1. On ad-hoc basis, part of Git is implemented in this repository if the
   implementation is easy and stable. For example the [pktline](../internal/git/pktline) package.

### Using Git

`git(1)` is the default choice to access repositories for Gitaly. Not all
commands that are available should be used in the Gitaly code base. Git makes
a distinction between porcelain and plumbing commands. Porcelain commands are
intended for the end-user and are the user-interface of the default `git`
client, where plumbing commands are intended for scripted use or to build
another porcelain. Gitaly should only use plumbing commands. `man 1 git`
contains a section on the low level plumbing.

When executing Git, developers should always use the `git.SafeCmd()` and sibling
interfaces. These make sure Gitaly is protected against command injection, the
correct `git` is used, and correct setup for observable command invocations are
used. When working with `git(1)` in Ruby, please be sure to read the
[Ruby shell scripting guide](https://docs.gitlab.com/ee/development/shell_commands.html).

### Using LibGit2

Gitaly uses [Git2Go](https://github.com/libgit2/git2go) for Golang, and
[Rugged](https://github.com/libgit2/rugged) which both are thin adapters to call
the C functions of LibGit2. Git2Go is always invoked through `cmd/gitaly-git2go`
to mitigate issues with context cancellation and the potential for memory leaks.
