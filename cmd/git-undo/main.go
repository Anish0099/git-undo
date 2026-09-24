package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/anish0099/git-undo/internal/undo"
)

// version is overridden at release build time via
// -ldflags "-X main.version=<tag>" (see .goreleaser.yaml). Local/dev builds
// report "dev".
var version = "dev"

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
		fmt.Fprintln(stdout, "git-undo", version)
		return 0
	}
	return undo.Run(undo.Options{Force: *force, DryRun: *dryRun, List: *list}, os.Stdin, stdout)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
