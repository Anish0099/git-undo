# git-undo

**Made a mess in git? Type two words.**

```bash
git undo
```

`git-undo` figures out what you just did — a `reset`, a `commit`, a `merge` — and
offers to safely reverse it, in plain English, with a preview and an automatic
backup. No more panic-Googling "how to undo git reset without losing work."

[![ci](https://github.com/anish0099/git-undo/actions/workflows/ci.yml/badge.svg)](https://github.com/anish0099/git-undo/actions/workflows/ci.yml)
[![license: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## The problem

Git keeps almost everything recoverable (via the reflog), but the commands to
recover are cryptic and different for every situation. `git-undo` reads git's own
logs, works out your last action, and runs the right recovery command for you —
after showing you exactly what it will do.

## What it looks like

```
$ git undo

Last action: reset (moving to HEAD~3)

Restore your branch to a3f21b0 (its state before the reset).

This will run:
  git reset --hard a3f21b0e9c4d1f77b2c8e5a6f0d3b1c9e7a4d2f8

! This overwrites your working tree to match the restored commit.

Proceed? [y/N]
```

Say `y` and it creates a backup ref first, then restores your commits. Say
anything else and it shows a numbered list of your recent actions so you can pick
a different one.

## Install

### With Go (easiest if you have Go)

```bash
go install github.com/anish0099/git-undo/cmd/git-undo@latest
```

This puts a `git-undo` binary in your Go bin directory (`go env GOPATH`/bin, usually
`~/go/bin`). Make sure that directory is on your `PATH`.

### Prebuilt binary (no Go required)

1. Download the archive for your OS/architecture from the
   [latest release](https://github.com/anish0099/git-undo/releases/latest).
2. Extract it and move the `git-undo` binary onto your `PATH`, e.g.:

   ```bash
   tar -xzf git-undo_*_linux_amd64.tar.gz
   sudo mv git-undo /usr/local/bin/    # or: mv git-undo ~/.local/bin/
   ```

Builds are published for Linux, macOS (Intel + Apple Silicon), and Windows.

### From source

```bash
git clone https://github.com/anish0099/git-undo
cd git-undo
go install ./cmd/git-undo
```

### Making `git undo` work

Git treats any executable named `git-<name>` on your `PATH` as the subcommand
`git <name>`. So once `git-undo` is on your `PATH`, both of these work:

```bash
git-undo      # direct
git undo      # as a git subcommand
```

If `git undo` reports *"'undo' is not a git command"*, the binary isn't on your
`PATH` yet — check `command -v git-undo`.

## Usage

```bash
git undo              # detect the last action, preview it, confirm
git undo --list       # pick from a list of recent actions instead
git undo --dry-run    # show the preview and exact commands, change nothing
git undo --force      # proceed even if uncommitted changes would be overwritten
git undo --version    # print the version
```

## What it can undo (v1)

| You did… | `git undo`… |
|----------|-------------|
| `git reset` (soft/mixed/hard) | restores your branch to where it was before the reset |
| `git commit` (including `--amend`) | removes the commit but **keeps all your changes staged** |
| `git merge` | restores your branch to its pre-merge state |

For anything it doesn't handle yet — a **rebase**, a **deleted branch**, a
**force-push** — it won't guess. It recognizes the action, tells you plainly that
v1 doesn't cover it, and shows the manual command so you're never left stuck.

## Safety

`git-undo` is built to be trusted:

- **Backup before it touches anything.** Before any undo, it saves your current
  state to a ref like `refs/git-undo/backup-20260924-143205`. The undo is always
  itself undoable:

  ```bash
  git reset --hard refs/git-undo/backup-20260924-143205
  ```

- **It won't silently destroy uncommitted work.** If an undo would overwrite
  uncommitted changes, it stops and tells you. Use `--force` (or stash first) to
  override, deliberately.

- **It's honest about the unrecoverable.** Changes that were never committed
  can't be brought back by anything, and `git-undo` says so rather than pretending.

- **No magic.** The preview shows the exact git commands it will run. It reads your
  reflog and runs standard git commands — no hooks, no setup, nothing installed
  into your repos.

## Roadmap

- Undo `rebase`
- Restore a deleted branch
- Undo a `force-push`
- `git undo redo` (reverse the last undo via its backup ref)

## How it works

`git-undo` is read-only inference over git's reflog: it parses `.git/logs/HEAD`,
classifies the most recent entry (`reset:`, `commit:`, `merge:` …), builds the
matching recovery plan, checks it's safe, and — only after you confirm and it has
made a backup — runs it. All of the deciding is done by pure functions; only a
single thin layer actually invokes git.

## Development

```bash
go test ./...       # run the full suite
go vet ./...        # static checks
go build ./cmd/git-undo
```

## License

[MIT](LICENSE)
