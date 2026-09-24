// Package safety decides whether an undo may run and names backup refs.
package safety

import (
	"time"

	"github.com/anish0099/git-undo/internal/gitcmd"
	"github.com/anish0099/git-undo/internal/plan"
)

// Verdict is the guard's decision.
type Verdict struct {
	Safe     bool
	Blockers []string
}

// Guard blocks undos that would overwrite uncommitted work, unless forced.
func Guard(p plan.UndoPlan, s gitcmd.RepoState, force bool) Verdict {
	if !p.Supported {
		return Verdict{Safe: false, Blockers: []string{"This action is not supported by git-undo v1."}}
	}
	if p.ClobbersWorkingTree && s.Dirty && !force {
		return Verdict{Safe: false, Blockers: []string{
			"You have uncommitted changes that this undo would overwrite. " +
				"Commit or stash them first, or re-run with --force to discard them.",
		}}
	}
	return Verdict{Safe: true}
}

// BackupRefName returns the ref used to snapshot HEAD before an undo.
func BackupRefName(now time.Time) string {
	return "refs/git-undo/backup-" + now.UTC().Format("20060102-150405")
}
