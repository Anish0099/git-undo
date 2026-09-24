package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/anish0099/git-undo/internal/undo"
)

// version is overridden at release build time via
// -ldflags "-X main.version=<tag>" (see .goreleaser.yaml). Local/dev builds
// leave it as "dev", in which case resolveVersion falls back to the module
// version Go embeds (so `go install ...@v0.1.0` still reports v0.1.0).
var version = "dev"

// resolveVersion returns the release-injected version when present, otherwise
// the module version embedded by `go install` (empty/"(devel)" for a plain
// local build, which falls back to "dev").
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return version
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("git-undo", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	list := fs.Bool("list", false, "choose from a list of recent actions")
	dryRun := fs.Bool("dry-run", false, "show what would happen, but do nothing")
	force := fs.Bool("force", false, "proceed even if uncommitted changes would be lost")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(stdout, "git-undo", resolveVersion())
		return 0
	}
	return undo.Run(undo.Options{Force: *force, DryRun: *dryRun, List: *list}, os.Stdin, stdout)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
