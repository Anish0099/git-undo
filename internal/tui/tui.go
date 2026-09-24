// Package tui renders previews/lists and reads simple prompts.
package tui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/anish/git-undo/internal/detect"
	"github.com/anish/git-undo/internal/plan"
	"github.com/anish/git-undo/internal/safety"
)

// RenderPreview builds the plain-English preview for an action and its plan.
func RenderPreview(a detect.Action, p plan.UndoPlan, v safety.Verdict) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Last action: %s\n\n", a.Summary)
	fmt.Fprintf(&b, "%s\n", p.Description)
	if p.ManualHint != "" {
		fmt.Fprintf(&b, "\n%s\n", p.ManualHint)
	}
	if len(p.Commands) > 0 {
		b.WriteString("\nThis will run:\n")
		for _, c := range p.Commands {
			fmt.Fprintf(&b, "  git %s\n", strings.Join(c, " "))
		}
	}
	for _, w := range p.Warnings {
		fmt.Fprintf(&b, "\n! %s\n", w)
	}
	for _, blk := range v.Blockers {
		fmt.Fprintf(&b, "\nBLOCKED: %s\n", blk)
	}
	return b.String()
}

// RenderList numbers actions 1-based in the order supplied (caller provides newest-first).
func RenderList(actions []detect.Action) string {
	var b strings.Builder
	b.WriteString("Recent actions:\n")
	for i, a := range actions {
		mark := ""
		if !a.Supported {
			mark = " (not supported yet)"
		}
		fmt.Fprintf(&b, "  %d. %s%s\n", i+1, a.Summary, mark)
	}
	return b.String()
}

// Confirm returns true only for y/yes (case-insensitive).
func Confirm(r *bufio.Reader, w io.Writer, prompt string) (bool, error) {
	fmt.Fprint(w, prompt)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return false, nil
	}
	s := strings.ToLower(strings.TrimSpace(line))
	return s == "y" || s == "yes", nil
}

// SelectFromList reads a 1-based choice, returning a 0-based index or -1.
func SelectFromList(r *bufio.Reader, w io.Writer, actions []detect.Action) (int, error) {
	fmt.Fprint(w, RenderList(actions))
	fmt.Fprint(w, "Pick a number to undo (or press Enter/q to cancel): ")
	line, _ := r.ReadString('\n')
	s := strings.TrimSpace(line)
	if s == "" || strings.EqualFold(s, "q") {
		return -1, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > len(actions) {
		return -1, nil
	}
	return n - 1, nil
}
