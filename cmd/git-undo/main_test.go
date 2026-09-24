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
	if !strings.Contains(out.String(), resolveVersion()) {
		t.Fatalf("stdout = %q, want it to contain version %q", out.String(), resolveVersion())
	}
}

func TestResolveVersionPrefersInjected(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })
	version = "v1.2.3"
	if got := resolveVersion(); got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want the injected %q", got, "v1.2.3")
	}
}

func TestResolveVersionNeverEmpty(t *testing.T) {
	// With no ldflags injection and no usable build info (the test binary),
	// resolveVersion must still fall back to a non-empty value.
	if got := resolveVersion(); got == "" {
		t.Fatal("resolveVersion() returned empty string")
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
