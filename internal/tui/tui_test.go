package tui

import (
	"strings"
	"testing"

	"github.com/anish/git-undo/internal/detect"
	"github.com/anish/git-undo/internal/plan"
	"github.com/anish/git-undo/internal/reflog"
	"github.com/anish/git-undo/internal/safety"
)

func TestRenderPreviewShowsCommandsAndWarnings(t *testing.T) {
	a := detect.Action{Kind: detect.KindReset, Summary: "reset (moving to HEAD~3)"}
	p := plan.UndoPlan{Supported: true, Description: "Restore your branch.",
		Commands: [][]string{{"reset", "--hard", "abc1234"}}, Warnings: []string{"overwrites tree"}}
	out := RenderPreview(a, p, safety.Verdict{Safe: true})
	for _, want := range []string{"Restore your branch.", "git reset --hard abc1234", "overwrites tree"} {
		if !strings.Contains(out, want) {
			t.Errorf("preview missing %q\n%s", want, out)
		}
	}
}

func TestRenderPreviewShowsBlockers(t *testing.T) {
	out := RenderPreview(detect.Action{}, plan.UndoPlan{Supported: true},
		safety.Verdict{Safe: false, Blockers: []string{"dirty tree blocker"}})
	if !strings.Contains(out, "dirty tree blocker") {
		t.Errorf("preview should show blockers\n%s", out)
	}
}

func TestRenderList(t *testing.T) {
	acts := []detect.Action{
		{Kind: detect.KindReset, Entry: reflog.Entry{Old: "aaaaaaa"}, Summary: "reset (HEAD~1)"},
		{Kind: detect.KindCommit, Entry: reflog.Entry{Old: "bbbbbbb"}, Summary: "commit: x"},
	}
	out := RenderList(acts)
	if !strings.Contains(out, "1.") || !strings.Contains(out, "2.") || !strings.Contains(out, "reset") {
		t.Errorf("list output wrong\n%s", out)
	}
}

func TestConfirm(t *testing.T) {
	yes, _ := Confirm(strings.NewReader("y\n"), &strings.Builder{}, "?")
	no, _ := Confirm(strings.NewReader("\n"), &strings.Builder{}, "?")
	if !yes || no {
		t.Fatalf("Confirm y=%v empty=%v", yes, no)
	}
}

func TestSelectFromList(t *testing.T) {
	acts := []detect.Action{{Summary: "a"}, {Summary: "b"}}
	idx, _ := SelectFromList(strings.NewReader("2\n"), &strings.Builder{}, acts)
	if idx != 1 {
		t.Fatalf("idx = %d, want 1", idx)
	}
	cancel, _ := SelectFromList(strings.NewReader("q\n"), &strings.Builder{}, acts)
	if cancel != -1 {
		t.Fatalf("cancel = %d, want -1", cancel)
	}
}
