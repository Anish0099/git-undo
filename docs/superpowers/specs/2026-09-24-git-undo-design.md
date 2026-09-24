# git-undo — Design Spec (v1)

- **Date:** 2026-09-24
- **Status:** Approved design, pre-implementation
- **Language:** Go (single static binary)
- **Primary goal:** A robust, honest tool people actually install and trust — robustness and honest edge-case handling over flash.

## 1. Summary

`git undo` is a command-line tool that detects a user's most recent git
action and offers to safely reverse it, with a plain-English preview and a
guaranteed backup. It is **read-only by inference** (Approach A): it figures
out what happened by reading git's own logs after the fact, requiring no
setup, no hooks, and no changes to the user's git.

```
git undo
```

The tool guesses the last action, previews the reversal in plain English,
and — only after the user confirms and a backup ref is created — executes the
exact git commands needed to reverse it.

## 2. Goals & non-goals

### Goals (v1)
- Reliably detect and reverse the three most common, most recoverable git
  panics: **reset**, **commit**, **merge**.
- Never surprise the user: preview before acting, backup before mutating.
- Be brutally honest about what git genuinely cannot recover.
- Zero setup: works retroactively, read-only, no hooks or shims.
- Robust, well-tested, single-binary distribution.

### Non-goals (v1 — deferred to later versions)
- Undo **rebase** (v2 — multi-entry reflog detection).
- Restore a **deleted branch** (v2 — dangling-commit scanning).
- Undo a **force-push** (v2 — involves the remote, not just local state).
- Rescuing **uncommitted** work that git never saved (impossible; we report
  it honestly rather than attempt it).
- Command-recording via git shim/hooks (possible future enhancement; not built).

When the user hits a non-goal action, v1 **detects and names it honestly**
rather than doing the wrong thing or failing silently.

## 3. Interaction model — "guess first, list on demand"

1. `git undo` detects the single most recent undoable action.
2. Shows a plain-English preview + the literal git commands + any warnings.
3. Prompts:
   - `y` → create backup, execute.
   - `N` / Enter → show a ranked list of recent undoable actions; user picks
     one; return to preview for that action.
   - `q` → exit, nothing touched.

Internally the detector always produces a **ranked list** (newest first); the
CLI simply shows the top item by default and the full list on demand
(`--list` flag or on "no"). Low-confidence detection skips the single-guess
and goes straight to the list.

## 4. Architecture

Design principle: **everything that decides is a pure function; only two tiny
units actually touch the repository.** This is what makes a safety tool
testable and trustworthy.

```
cmd/git-undo/main.go     Thin entrypoint; wires cobra.
internal/gitcmd/         The ONLY code that shells out to git.
                         (RunGit, RunGitCapture; centralizes errors + dry-run.)
internal/reflog/         Reads .git/logs/HEAD + `git reflog`; parses to []Entry.
internal/detect/         Pure: maps []reflog.Entry -> []Action (ranked, newest first).
internal/plan/           Pure: Action -> UndoPlan{Commands, Description, Warnings}.
internal/safety/         Pure: Guard(plan, repoState) -> Verdict{Safe, Blockers, DataAtRisk}.
                         Also owns backup-ref creation logic.
internal/tui/            Preview rendering, y/N prompt, ranked-list picker.
internal/undo/           Orchestrator: detect -> plan -> confirm -> backup -> execute.
                         undo.Execute is the ONLY unit that mutates the repo.
```

### Key interfaces (contracts)
- `reflog.Parse() -> []reflog.Entry` — pure parsing, no decisions.
- `detect.Detect([]reflog.Entry) -> []detect.Action` — pure inference, no git calls.
- `plan.Build(Action) -> plan.UndoPlan` — pure; produces what *would* run.
- `safety.Guard(plan, repoState) -> Verdict` — pure decision.
- `undo.Execute(plan)` — the only mutating unit; runs only after backup.

## 5. Data flow (one invocation)

```
1. gitcmd verifies we're in a git repo (else: friendly error, non-zero exit).
2. reflog.Parse reads recent history.
3. detect.Detect ranks undoable actions (newest first).
4. Top action -> plan.Build -> UndoPlan.
5. safety.Guard checks: dirty tree? unrecoverable data? -> Verdict.
6. tui renders plain-English preview + warnings + exact git commands.
7. User responds:
      y        -> step 8
      N/Enter  -> ranked list; pick one -> back to step 4
      q        -> exit, nothing touched
8. safety creates backup ref (refs/git-undo/backup-<UTC-timestamp>) at current HEAD;
   records current branch name.
9. undo.Execute runs the plan's git commands.
10. Confirm success + print the backup ref + the exact restore command.
```

## 6. Safety model (the heart of it)

1. **Always back up before mutating.** Create `refs/git-undo/backup-<UTC-timestamp>`
   pointing at current HEAD before any undo executes. Use our own ref namespace
   (not a real branch) to avoid cluttering `git branch`, but print the exact
   restore command. The undo is always itself undoable.

2. **Never silently destroy uncommitted work.** Before an undo that would
   overwrite the working tree (e.g. undoing `reset --hard`), `safety.Guard`
   checks for a dirty working tree / staged changes. If undoing would clobber
   uncommitted work, **block by default** and explain; offer `--force` (or
   stash-first) as an explicit, informed override. This is the single most
   important guardrail.

3. **Be brutally honest about the unrecoverable.** When we detect a shape where
   git never saved the data (e.g. `reset --hard` that discarded uncommitted
   changes), we say so plainly instead of implying a rescue. Overpromising here
   destroys trust and violates the primary goal.

4. **Preview shows the real commands.** The plain-English description is backed
   by the literal git commands we will run, shown to the user. No hidden magic.

## 7. Error handling & honesty

- **Not a git repo / no reflog / empty history** → friendly, specific message,
  non-zero exit. Never a stack trace.
- **Ambiguous detection** → skip the single guess; go straight to the ranked list.
- **Unsupported action** (e.g. rebase in v1) → detect and name it honestly, show
  the reflog entry and the manual command the user could run.
- **git command fails mid-undo** → report the failure, point at the backup ref,
  stop. Never leave the user guessing about repo state.
- All git calls go through `internal/gitcmd`, so error handling and dry-run
  support live in exactly one place.

## 8. CLI surface (v1)

```
git undo               Detect last action, preview, confirm, undo.
git undo --list        Show the ranked list of recent undoable actions directly.
git undo --dry-run     Show the preview + exact commands, but never execute.
git undo --force       Proceed even when uncommitted work would be clobbered.
git undo --help        Usage.
git undo --version     Version.
```

Installable as `git-undo` on PATH so that `git undo` works as a git subcommand.

## 9. Testing strategy

Testing is the product for a safety tool. TDD throughout — failing test first,
especially for safety guardrails.

- **`detect` and `plan` are pure functions** → unit-tested against **fixture
  reflogs** (canned `.git/logs/HEAD` contents per scenario: soft/mixed/hard
  reset, commit, merge, plus ambiguous and unsupported cases). Fast,
  deterministic, no real repos.
- **`safety.Guard`** → table-driven tests over (plan × repo-state), especially
  the dirty-tree-blocks cases.
- **Integration tests** → a helper spins up a real temp git repo, performs an
  action (e.g. an actual `reset --hard`), runs the full undo pipeline, and
  asserts the repo returns to the expected state AND that a backup ref exists.

## 10. Distribution (post-MVP, noted not built)

- Single static Go binary; trivial cross-compilation.
- Homebrew formula (`brew install git-undo`), `go install` one-liner.
- GitHub README with animated GIFs per scenario.

## 11. Open questions / future doors

- v2: rebase, deleted-branch restore, force-push undo.
- Optional future: command-recording (git shim/hooks) for higher-fidelity
  detection, layered on top of reflog inference (Approach C).
- Possible `git undo redo` to reverse the last undo via the backup ref.
