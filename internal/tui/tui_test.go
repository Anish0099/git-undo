package tui

import (
	"bufio"
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
	yes, _ := Confirm(bufio.NewReader(strings.NewReader("y\n")), &strings.Builder{}, "?")
	no, _ := Confirm(bufio.NewReader(strings.NewReader("\n")), &strings.Builder{}, "?")
	if !yes || no {
		t.Fatalf("Confirm y=%v empty=%v", yes, no)
	}
}

func TestSelectFromList(t *testing.T) {
	acts := []detect.Action{{Summary: "a"}, {Summary: "b"}}
	idx, _ := SelectFromList(bufio.NewReader(strings.NewReader("2\n")), &strings.Builder{}, acts)
	if idx != 1 {
		t.Fatalf("idx = %d, want 1", idx)
	}
	cancel, _ := SelectFromList(bufio.NewReader(strings.NewReader("q\n")), &strings.Builder{}, acts)
	if cancel != -1 {
		t.Fatalf("cancel = %d, want -1", cancel)
	}
}

func TestRenderPreviewShowsManualHint(t *testing.T) {
	a := detect.Action{Kind: detect.KindRebase, Summary: "rebase (finish)"}
	p := plan.UndoPlan{Supported: false, Description: "This action is not yet supported.",
		ManualHint: "run git reset --hard abc1234 manually"}
	out := RenderPreview(a, p, safety.Verdict{Safe: true})
	if !strings.Contains(out, "run git reset --hard abc1234 manually") {
		t.Errorf("preview should show manual hint\n%s", out)
	}
}

func TestRenderListShowsUnsupportedMarker(t *testing.T) {
	acts := []detect.Action{
		{Kind: detect.KindRebase, Entry: reflog.Entry{Old: "aaaaaaa"}, Summary: "rebase (finish)", Supported: false},
	}
	out := RenderList(acts)
	if !strings.Contains(out, "(not supported yet)") {
		t.Errorf("list should show unsupported marker\n%s", out)
	}
}
