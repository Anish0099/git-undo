package gitcmd

import (
	"os/exec"
	"testing"
)

func initRepo(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "t@t"},
		{"config", "user.name", "t"},
		{"commit", "--allow-empty", "-q", "-m", "first"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestInRepoAndState(t *testing.T) {
	initRepo(t)
	in, err := InRepo()
	if err != nil || !in {
		t.Fatalf("InRepo = %v, %v; want true, nil", in, err)
	}
	st, err := ReadState()
	if err != nil {
		t.Fatal(err)
	}
	if st.CurrentBranch == "" || len(st.HeadHash) < 7 {
		t.Fatalf("state = %+v", st)
	}
	if st.Dirty {
		t.Fatalf("fresh repo should be clean: %+v", st)
	}
}

func TestReadHeadLogHasEntries(t *testing.T) {
	initRepo(t)
	log, err := ReadHeadLog()
	if err != nil {
		t.Fatal(err)
	}
	if log == "" {
		t.Fatal("expected reflog content after a commit")
	}
}

func TestInRepoNotARepo(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	in, err := InRepo()
	if err != nil || in {
		t.Fatalf("InRepo = %v, %v; want false, nil", in, err)
	}
}
