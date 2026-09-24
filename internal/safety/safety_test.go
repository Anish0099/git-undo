package safety

import (
	"testing"
	"time"

	"github.com/anish/git-undo/internal/gitcmd"
	"github.com/anish/git-undo/internal/plan"
)

func TestGuardBlocksDirtyClobber(t *testing.T) {
	p := plan.UndoPlan{Supported: true, ClobbersWorkingTree: true}
	v := Guard(p, gitcmd.RepoState{Dirty: true}, false)
	if v.Safe || len(v.Blockers) == 0 {
		t.Fatalf("dirty + clobber should block: %+v", v)
	}
}

func TestGuardForceOverrides(t *testing.T) {
	p := plan.UndoPlan{Supported: true, ClobbersWorkingTree: true}
	if v := Guard(p, gitcmd.RepoState{Dirty: true}, true); !v.Safe {
		t.Fatalf("force should allow: %+v", v)
	}
}

func TestGuardCleanTreeSafe(t *testing.T) {
	p := plan.UndoPlan{Supported: true, ClobbersWorkingTree: true}
	if v := Guard(p, gitcmd.RepoState{Dirty: false}, false); !v.Safe {
		t.Fatalf("clean tree should be safe: %+v", v)
	}
}

func TestGuardNonClobberDirtySafe(t *testing.T) {
	p := plan.UndoPlan{Supported: true, ClobbersWorkingTree: false}
	if v := Guard(p, gitcmd.RepoState{Dirty: true}, false); !v.Safe {
		t.Fatalf("non-clobbering undo is safe even when dirty: %+v", v)
	}
}

func TestBackupRefName(t *testing.T) {
	got := BackupRefName(time.Date(2026, 9, 24, 14, 32, 5, 0, time.UTC))
	want := "refs/git-undo/backup-20260924-143205"
	if got != want {
		t.Fatalf("BackupRefName = %q, want %q", got, want)
	}
}
