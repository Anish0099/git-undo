package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var version = "0.1.0-dev"

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("git-undo", flag.ContinueOnError)
	fs.SetOutput(stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *showVersion {
		fmt.Fprintln(stdout, "git-undo", version)
		return 0
	}
	fmt.Fprintln(stdout, "git-undo", version, "- run with a supported action in a git repo")
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
