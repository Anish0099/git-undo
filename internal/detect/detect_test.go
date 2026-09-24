package detect

import (
	"testing"

	"github.com/anish0099/git-undo/internal/reflog"
)

func TestDetectClassifies(t *testing.T) {
	in := []reflog.Entry{
		{Old: "a", New: "b", Message: "reset: moving to HEAD~1"},
		{Old: "b", New: "c", Message: "commit: work"},
		{Old: "c", New: "d", Message: "merge feature: Merge branch 'feature'"},
		{Old: "d", New: "e", Message: "rebase (finish): returning to refs/heads/main"},
		{Old: "e", New: "f", Message: "checkout: moving from x to y"},
	}
	got := Detect(in)
	wantKind := []Kind{KindReset, KindCommit, KindMerge, KindRebase, KindUnknown}
	wantSupported := []bool{true, true, true, false, false}
	for i := range got {
		if got[i].Kind != wantKind[i] {
			t.Errorf("entry %d kind = %q, want %q", i, got[i].Kind, wantKind[i])
		}
		if got[i].Supported != wantSupported[i] {
			t.Errorf("entry %d supported = %v, want %v", i, got[i].Supported, wantSupported[i])
		}
	}
}
