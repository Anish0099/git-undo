// Package reflog parses git's HEAD reflog into structured entries.
package reflog

import "strings"

// Entry is one line of .git/logs/HEAD.
type Entry struct {
	Old     string // object name before the action ("0"*40 for none)
	New     string // object name after the action
	Message string // reflog message, e.g. "reset: moving to HEAD~1"
}

// ParseLog parses .git/logs/HEAD contents and returns entries newest-first.
// Blank and malformed lines are skipped.
func ParseLog(content string) []Entry {
	var entries []Entry
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, "\r")
		tab := strings.IndexByte(line, '\t')
		if tab < 0 {
			continue
		}
		fields := strings.Fields(line[:tab])
		if len(fields) < 2 {
			continue
		}
		entries = append(entries, Entry{Old: fields[0], New: fields[1], Message: line[tab+1:]})
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}
