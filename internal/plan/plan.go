// Package plan turns a detected action into a concrete, previewable undo plan.
package plan

import (
	"fmt"
	"strings"

	"github.com/anish0099/git-undo/internal/detect"
)

const zeroHash = "0000000000000000000000000000000000000000"

// UndoPlan describes exactly what an undo will do.
type UndoPlan struct {
	Description         string     // plain-English summary
	Commands            [][]string // git arg slices, run in order (no leading "git")
	Warnings            []string
	ClobbersWorkingTree bool
	Supported           bool   // false => nothing to execute; explain instead
	ManualHint          string // guidance shown when unsupported
}

// Build produces the undo plan for a detected action.
func Build(a detect.Action) UndoPlan {
	switch a.Kind {
	case detect.KindReset:
		return UndoPlan{
			Supported:           true,
			Description:         fmt.Sprintf("Restore your branch to %s (its state before the reset).", short(a.Entry.Old)),
			Commands:            [][]string{{"reset", "--hard", a.Entry.Old}},
			Warnings:            []string{"This overwrites your working tree to match the restored commit."},
			ClobbersWorkingTree: true,
		}
	case detect.KindCommit:
		if a.Entry.Old == zeroHash || strings.HasPrefix(a.Entry.Message, "commit (initial):") {
			return UndoPlan{
				Supported:   false,
				Description: "Your last commit was the initial commit.",
				ManualHint:  "Undoing the very first commit is unusual. If you're sure, run: git update-ref -d HEAD (this keeps your files but removes the commit).",
			}
		}
		return UndoPlan{
			Supported:           true,
			Description:         "Undo the last commit, keeping all your changes staged.",
			Commands:            [][]string{{"reset", "--soft", a.Entry.Old}},
			ClobbersWorkingTree: false,
		}
	case detect.KindMerge:
		return UndoPlan{
			Supported:           true,
			Description:         fmt.Sprintf("Undo the merge, restoring your branch to %s.", short(a.Entry.Old)),
			Commands:            [][]string{{"reset", "--hard", a.Entry.Old}},
			Warnings:            []string{"This overwrites your working tree to match the pre-merge commit."},
			ClobbersWorkingTree: true,
		}
	default:
		return UndoPlan{
			Supported:   false,
			Description: fmt.Sprintf("Your last action looks like %q, which git-undo v1 doesn't handle yet.", a.Summary),
			ManualHint:  fmt.Sprintf("Its reflog entry restores to %s. To undo manually (review first): git reset --hard %s", short(a.Entry.Old), a.Entry.Old),
		}
	}
}

func short(h string) string {
	if len(h) >= 7 {
		return h[:7]
	}
	return h
}
