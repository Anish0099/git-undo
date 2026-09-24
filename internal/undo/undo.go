// Package undo orchestrates detect -> plan -> guard -> backup -> execute.
package undo

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/anish/git-undo/internal/detect"
	"github.com/anish/git-undo/internal/gitcmd"
	"github.com/anish/git-undo/internal/plan"
	"github.com/anish/git-undo/internal/reflog"
	"github.com/anish/git-undo/internal/safety"
	"github.com/anish/git-undo/internal/tui"
)

// Options configures a run.
type Options struct {
	Force  bool
	DryRun bool
	List   bool
}

// Run executes one git-undo invocation. Returns a process exit code.
func Run(opts Options, in io.Reader, out io.Writer) int {
	inRepo, err := gitcmd.InRepo()
	if err != nil {
		fmt.Fprintln(out, "git is not available:", err)
		return 1
	}
	if !inRepo {
		fmt.Fprintln(out, "Not a git repository. Run git-undo from inside one.")
		return 1
	}

	logText, err := gitcmd.ReadHeadLog()
	if err != nil {
		fmt.Fprintln(out, "Could not read the reflog:", err)
		return 1
	}
	actions := detect.Detect(reflog.ParseLog(logText))
	if len(actions) == 0 {
		fmt.Fprintln(out, "No recent actions found in the reflog.")
		return 0
	}

	br := bufio.NewReader(in)

	chosen := actions[0]
	if opts.List {
		idx, _ := tui.SelectFromList(br, out, actions)
		if idx < 0 {
			fmt.Fprintln(out, "Cancelled.")
			return 0
		}
		chosen = actions[idx]
	}

	for {
		p := plan.Build(chosen)
		if !p.Supported {
			fmt.Fprint(out, tui.RenderPreview(chosen, p, safety.Verdict{}))
			if opts.DryRun {
				return 0
			}
			idx, _ := tui.SelectFromList(br, out, actions)
			if idx < 0 {
				fmt.Fprintln(out, "Cancelled. Nothing was changed.")
				return 0
			}
			chosen = actions[idx]
			continue
		}
		st, err := gitcmd.ReadState()
		if err != nil {
			fmt.Fprintln(out, "Could not read repo state:", err)
			return 1
		}
		v := safety.Guard(p, st, opts.Force)
		fmt.Fprint(out, tui.RenderPreview(chosen, p, v))
		if !v.Safe {
			return 1
		}
		if opts.DryRun {
			return 0
		}
		ok, _ := tui.Confirm(br, out, "\nProceed? [y/N] ")
		if ok {
			return execute(p, out)
		}
		idx, _ := tui.SelectFromList(br, out, actions)
		if idx < 0 {
			fmt.Fprintln(out, "Cancelled. Nothing was changed.")
			return 0
		}
		chosen = actions[idx]
	}
}

func execute(p plan.UndoPlan, out io.Writer) int {
	backup := safety.BackupRefName(time.Now())
	if err := gitcmd.Run("update-ref", backup, "HEAD"); err != nil {
		fmt.Fprintln(out, "Could not create a backup ref; aborting:", err)
		return 1
	}
	for _, args := range p.Commands {
		if err := gitcmd.Run(args...); err != nil {
			fmt.Fprintf(out, "A git command failed: git %v\nYour previous state is saved at %s\n", args, backup)
			return 1
		}
	}
	fmt.Fprintf(out, "\nDone. Your previous state is saved at %s\n", backup)
	fmt.Fprintf(out, "To undo this undo: git reset --hard %s\n", backup)
	return 0
}
