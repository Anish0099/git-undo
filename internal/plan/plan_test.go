package plan

import (
	"testing"

	"github.com/anish/git-undo/internal/detect"
	"github.com/anish/git-undo/internal/reflog"
)

func act(k detect.Kind, old, msg string) detect.Action {
	return detect.Action{Kind: k, Supported: k != detect.KindRebase && k != detect.KindUnknown,
		Entry: reflog.Entry{Old: old, New: "new", Message: msg}, Summary: msg}
}

func TestBuildReset(t *testing.T) {
	p := Build(act(detect.KindReset, "abc1234def", "reset: moving to HEAD~3"))
	if !p.Supported || !p.ClobbersWorkingTree {
		t.Fatalf("reset should be supported and clobbering: %+v", p)
	}
	if len(p.Commands) != 1 || p.Commands[0][0] != "reset" || p.Commands[0][1] != "--hard" || p.Commands[0][2] != "abc1234def" {
		t.Fatalf("commands = %v", p.Commands)
	}
}

func TestBuildCommitSoftReset(t *testing.T) {
	p := Build(act(detect.KindCommit, "abc1234def", "commit: work"))
	if !p.Supported || p.ClobbersWorkingTree {
		t.Fatalf("commit undo should be supported and non-clobbering: %+v", p)
	}
	if len(p.Commands) != 1 || p.Commands[0][2] != "HEAD~1" || p.Commands[0][1] != "--soft" {
		t.Fatalf("commands = %v", p.Commands)
	}
}

func TestBuildInitialCommitUnsupported(t *testing.T) {
	p := Build(act(detect.KindCommit, "0000000000000000000000000000000000000000", "commit (initial): first"))
	if p.Supported || p.ManualHint == "" {
		t.Fatalf("initial commit should be unsupported with a manual hint: %+v", p)
	}
}

func TestBuildMerge(t *testing.T) {
	p := Build(act(detect.KindMerge, "abc1234def", "merge feature: Merge branch 'feature'"))
	if !p.Supported || !p.ClobbersWorkingTree || p.Commands[0][2] != "abc1234def" {
		t.Fatalf("merge undo = %+v", p)
	}
}

func TestBuildRebaseUnsupported(t *testing.T) {
	p := Build(act(detect.KindRebase, "abc1234def", "rebase (finish): returning"))
	if p.Supported || len(p.Commands) != 0 || p.ManualHint == "" {
		t.Fatalf("rebase should be unsupported with a manual hint, no commands: %+v", p)
	}
}
