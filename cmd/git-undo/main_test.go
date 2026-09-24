package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run([]string{"--version"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), version) {
		t.Fatalf("stdout = %q, want it to contain version %q", out.String(), version)
	}
}

func TestRunNotARepo(t *testing.T) {
	t.Chdir(t.TempDir())
	var out, errOut bytes.Buffer
	code := run(nil, &out, &errOut)
	if code != 1 || !strings.Contains(out.String(), "Not a git repository") {
		t.Fatalf("code=%d out=%q", code, out.String())
	}
}
