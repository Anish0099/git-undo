// Package detect classifies reflog entries into undoable actions.
package detect

import (
	"strings"

	"github.com/anish/git-undo/internal/reflog"
)

type Kind string

const (
	KindReset   Kind = "reset"
	KindCommit  Kind = "commit"
	KindMerge   Kind = "merge"
	KindRebase  Kind = "rebase"
	KindUnknown Kind = "unknown"
)

// Action is a classified reflog entry.
type Action struct {
	Kind      Kind
	Entry     reflog.Entry
	Summary   string
	Supported bool
}

// Detect classifies each entry. Input and output are both newest-first.
func Detect(entries []reflog.Entry) []Action {
	actions := make([]Action, 0, len(entries))
	for _, e := range entries {
		actions = append(actions, classify(e))
	}
	return actions
}

func classify(e reflog.Entry) Action {
	m := e.Message
	switch {
	case strings.HasPrefix(m, "reset:"):
		return Action{Kind: KindReset, Entry: e, Supported: true,
			Summary: "reset (" + strings.TrimSpace(strings.TrimPrefix(m, "reset:")) + ")"}
	case strings.HasPrefix(m, "commit:"),
		strings.HasPrefix(m, "commit (amend):"),
		strings.HasPrefix(m, "commit (initial):"):
		return Action{Kind: KindCommit, Entry: e, Supported: true, Summary: m}
	case strings.HasPrefix(m, "merge "):
		return Action{Kind: KindMerge, Entry: e, Supported: true, Summary: m}
	case strings.HasPrefix(m, "rebase"):
		return Action{Kind: KindRebase, Entry: e, Supported: false, Summary: m}
	default:
		return Action{Kind: KindUnknown, Entry: e, Supported: false, Summary: m}
	}
}
