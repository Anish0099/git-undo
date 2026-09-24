// Package gitcmd is the only code that executes git and reads repo state.
package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoState is a snapshot of the working repository.
type RepoState struct {
	HeadHash      string
	CurrentBranch string // "" if detached
	Dirty         bool   // uncommitted staged or unstaged changes
}

// Capture runs `git args...` and returns trimmed stdout.
func Capture(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// Run runs `git args...`, forwarding output to the current process.
func Run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InRepo reports whether the cwd is inside a git work tree.
func InRepo() (bool, error) {
	out, err := Capture("rev-parse", "--is-inside-work-tree")
	if err != nil {
		return false, nil // git exits non-zero outside a repo; treat as "not in repo"
	}
	return out == "true", nil
}

// ReadState snapshots HEAD, branch, and dirtiness.
func ReadState() (RepoState, error) {
	var st RepoState
	head, err := Capture("rev-parse", "HEAD")
	if err != nil {
		return st, err
	}
	st.HeadHash = head
	branch, err := Capture("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return st, err
	}
	if branch != "HEAD" {
		st.CurrentBranch = branch
	}
	status, err := Capture("status", "--porcelain")
	if err != nil {
		return st, err
	}
	st.Dirty = strings.TrimSpace(status) != ""
	return st, nil
}

// ReadHeadLog returns the contents of .git/logs/HEAD ("" if absent).
func ReadHeadLog() (string, error) {
	dir, err := Capture("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "logs", "HEAD"))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}
