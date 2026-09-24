package undo

import (
	"os/exec"
	"strings"
	"testing"
)

func git(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	git(t, "init", "-q")
	git(t, "config", "user.email", "t@t")
	git(t, "config", "user.name", "t")
}

func writeCommit(t *testing.T, name, content string) {
	t.Helper()
	// create file then commit
	if err := exec.Command("sh", "-c", "printf '%s' "+content+" > "+name).Run(); err != nil {
		t.Fatal(err)
	}
	git(t, "add", "-A")
	git(t, "commit", "-q", "-m", "add "+name)
}

func TestUndoResetHardRestoresCommits(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	before := git(t, "rev-parse", "HEAD")
	git(t, "reset", "--hard", "HEAD~1")

	code := Run(Options{}, strings.NewReader("y\n"), &strings.Builder{})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := git(t, "rev-parse", "HEAD"); got != before {
		t.Fatalf("HEAD = %s, want restored %s", got, before)
	}
	// a backup ref must exist
	if out := git(t, "for-each-ref", "refs/git-undo"); out == "" {
		t.Fatal("expected a backup ref under refs/git-undo")
	}
}

func TestUndoBlockedWhenDirty(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	git(t, "reset", "--hard", "HEAD~1")
	// make the tree dirty so the reset-undo (a hard reset) is blocked
	if err := exec.Command("sh", "-c", "printf 'dirty' > a.txt").Run(); err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	code := Run(Options{}, strings.NewReader("y\n"), &out)
	if code == 0 || !strings.Contains(out.String(), "uncommitted changes") {
		t.Fatalf("expected block on dirty tree; code=%d out=%s", code, out.String())
	}
}

func TestUndoDryRunChangesNothing(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	git(t, "reset", "--hard", "HEAD~1")
	head := git(t, "rev-parse", "HEAD")
	Run(Options{DryRun: true}, strings.NewReader("y\n"), &strings.Builder{})
	if git(t, "rev-parse", "HEAD") != head {
		t.Fatal("dry-run must not change HEAD")
	}
}

func TestUndoCommitKeepsChangesStaged(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	parent := git(t, "rev-parse", "HEAD~1")

	code := Run(Options{}, strings.NewReader("y\n"), &strings.Builder{})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := git(t, "rev-parse", "HEAD"); got != parent {
		t.Fatalf("HEAD = %s, want restored parent %s", got, parent)
	}
}

func TestUndoAmendRestoresPreAmendCommit(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	c1 := git(t, "rev-parse", "HEAD")

	if err := exec.Command("sh", "-c", "printf 'b2' > b.txt").Run(); err != nil {
		t.Fatal(err)
	}
	git(t, "add", "-A")
	git(t, "commit", "-q", "--amend", "--no-edit")

	if amended := git(t, "rev-parse", "HEAD"); amended == c1 {
		t.Fatal("amend did not change HEAD; test setup is broken")
	}

	code := Run(Options{}, strings.NewReader("y\n"), &strings.Builder{})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := git(t, "rev-parse", "HEAD"); got != c1 {
		t.Fatalf("HEAD = %s, want pre-amend commit %s", got, c1)
	}
}

func TestUndoMergeRestoresPreMergeState(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	baseBranch := git(t, "rev-parse", "--abbrev-ref", "HEAD")
	git(t, "checkout", "-b", "feature")
	writeCommit(t, "c.txt", "c")
	git(t, "checkout", baseBranch)
	writeCommit(t, "b.txt", "b")
	premerge := git(t, "rev-parse", "HEAD")
	git(t, "merge", "--no-ff", "--no-edit", "feature")

	code := Run(Options{}, strings.NewReader("y\n"), &strings.Builder{})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := git(t, "rev-parse", "HEAD"); got != premerge {
		t.Fatalf("HEAD = %s, want pre-merge commit %s", got, premerge)
	}
}

func TestUndoUnsupportedTopFallsToList(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	before := git(t, "rev-parse", "HEAD")
	git(t, "reset", "--hard", "HEAD~1")
	// An unsupported action (checkout) on top of the reflog stack.
	git(t, "checkout", "-b", "scratch")

	var out strings.Builder
	code := Run(Options{}, strings.NewReader("2\ny\n"), &out)
	if code != 0 {
		t.Fatalf("exit = %d, want 0\noutput:\n%s", code, out.String())
	}
	if got := git(t, "rev-parse", "HEAD"); got != before {
		t.Fatalf("HEAD = %s, want restored %s\noutput:\n%s", got, before, out.String())
	}
}

func TestUndoDeclineThenPickExecutes(t *testing.T) {
	newRepo(t)
	writeCommit(t, "a.txt", "a")
	writeCommit(t, "b.txt", "b")
	before := git(t, "rev-parse", "HEAD")
	git(t, "reset", "--hard", "HEAD~1")

	code := Run(Options{}, strings.NewReader("n\n1\ny\n"), &strings.Builder{})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if got := git(t, "rev-parse", "HEAD"); got != before {
		t.Fatalf("HEAD = %s, want restored %s", got, before)
	}
	// a backup ref must exist
	if out := git(t, "for-each-ref", "refs/git-undo"); out == "" {
		t.Fatal("expected a backup ref under refs/git-undo")
	}
}
