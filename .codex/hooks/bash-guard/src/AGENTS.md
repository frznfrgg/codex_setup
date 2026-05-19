# bash-guard — maintenance guide

This is the Go-only Codex Bash safety hook. It keeps an internal `allow|ask` rule model, then maps risky `ask` decisions to Codex's supported `PreToolUse` deny output.

## Locations

- **Source** (this dir): `<repo>/.codex/hooks/bash-guard/src/`
- **Binary**: `<repo>/.codex/hooks/bash-guard/src/bash_guard.bin`
- **Audit log**: `~/.codex/logs/bash-guard.jsonl`
- **Hook config**: `<repo>/.codex/hooks.json` → `hooks.PreToolUse[matcher=Bash]`

## Hard rules (do NOT relax)

1. **Keep the internal rule model to `allow` or `ask`.** Codex-facing `deny` output belongs only at the final hook response boundary.
2. **Asymmetric fail-open**: pre-trigger failures → `allow`; post-trigger failures (parse error after a trigger keyword matched) → `ask`. Don't change one without the other.
3. **Untrusted project configs**: `<repo>/.codex/bash-guard.toml` is NOT auto-loaded. Trust requires `trusted-projects.toml` opt-in. Don't add an "auto-load if signed" shortcut.

## Edit + rebuild loop

```bash
cd <repo>/.codex/hooks/bash-guard/src
# edit any .go file
make build     # ~50 ms warm; produces bash_guard.bin
make test      # full suite, ~700 ms
```

Codex loads `bash_guard.bin` directly for both `PreToolUse` and `UserPromptSubmit`. The next Bash hook invocation picks up the rebuilt binary automatically.

For tight iteration use `make watch` (requires `entr`; `brew install entr`).

## Testing protocol

Every behavioural change must come with a fixture. Tests are golden-table style under `testdata/fixtures/*.json` — one file per case, ~155 covered today. Pattern:

```json
{
  "name": "rm_rf_some_path",
  "description": "what it covers",
  "input": {
    "tool_name": "Bash",
    "tool_input": { "command": "..." },
    "cwd": "/home/example-user/myproject"
  },
  "expect": {
    "decision": "allow|ask",
    "rule": "default|rm|aggregate",
    "reason_code": "no_rule_matched|rm.catastrophic|...",
    "reason_substring": "optional"
  }
}
```

Reasons-strings are NOT in golden tests on purpose — they evolve. Assert on `(decision, rule, reason_code)` tuple plus optional substring.

## Adding a new rule

1. New file `rule_<name>.go`, type implementing `Rule` interface from `rule.go`:
   - `Name()`, `Triggers()`, `Check(cmd ExecutedCommand, env *RuleEnv) *Decision`
   - `Triggers()` returns the keyword set the rule applies to (matched against `ExecutedCommand.Name` after parser unwrap).
2. Register it in `main.go` → `newRegistry(RmRule{}, MyRule{})`.
3. If your rule introduces new executor keywords (e.g., a new shell-evaluator) that should force AST descent, add them to `parserDescentKeywords` in `main.go`.
4. Add ≥3 fixtures: one allow case, one ask case, one corner case.
5. `make test` must pass before commit.

## Adding an executor wrapper to the unwrap table

If a new wrapper command needs unwrapping (e.g., the team starts using `firejail` or `bwrap`):

1. Add a case in `unwrap.go` → `stripExecutorPrefix`.
2. Add the keyword to `parserDescentKeywords` in `main.go` so quick-reject doesn't skip it.
3. Add a fixture `<wrapper>_rm_rf_etc.json` covering the unwrapped path.

## Switching Modes

Set hook mode with `BASH_GUARD_CODEX_MODE`:

```bash
BASH_GUARD_CODEX_MODE=shadow codex   # logs and reports would-block context
BASH_GUARD_CODEX_MODE=dry-run codex  # same behavior, distinct label
BASH_GUARD_CODEX_MODE=live codex     # default; blocks risky commands
```

After changing hook mode with `BASH_GUARD_CODEX_MODE`, no Codex restart is needed.

## Inspecting behaviour

- **Audit log**: `tail -f ~/.codex/logs/bash-guard.jsonl | jq '.'`
- **Selftest** (4 fixed cases including FP-1, FP-2 regressions): `./bash_guard.bin --selftest`
- **Single-shot** (raw): `echo '{"tool_name":"Bash","tool_input":{"command":"rm -rf /tmp/x"},"cwd":"/tmp"}' | ./bash_guard.bin`
- **Force re-trigger build** (after `make clean` or upstream Go upgrade): `make build`

## What NOT to do

- Don't pre-commit `bash_guard.bin` — gitignored on purpose. Cross-arch and supply-chain concerns.
- Don't add a runtime `go build` wrapper "for convenience" — measured at ~80 ms/invocation, killed in v3.1. The Makefile is the rebuild path.
- Don't paper over a parser bug with a regex pre-strip in `quickReject` — fix the AST walker in `parser.go` instead. Regex hacks accumulate and silently mis-classify.
- Don't add per-rule reason-string golden tests. Reasons rotate as English improves; tests would break for cosmetic edits. Golden table is `(decision, rule, reason_code)` only.
- Don't disable the post-trigger fail-open asymmetry. If the parser breaks on a real corpus item, the right move is to add a fixture and fix the parser, not to widen fail-open.

## Performance budget

| Path | Budget | Current |
|---|---|---|
| Quick-reject hit (no trigger) | < 1 ms | ~0.16 ms |
| Full parse + rule eval | < 30 ms | < 5 ms |
| End-to-end (incl. exec startup) | < 50 ms warm | 0-10 ms warm |

If you change anything in `parser.go` or `unwrap.go`, run a manual sanity check on a heredoc-heavy or pipeline-heavy command — those exercise the AST walker most.

## Pointers

- `main.go`: pipeline orchestration, hook I/O JSON shapes
- `parser.go`: mvdan/sh AST walk + span classification
- `unwrap.go`: executor wrapper handling (sudo, env, ssh, find, xargs, ...)
- `safe_paths.go`: allowlist + catastrophic-path classification + symlink semantics
- `rule_rm.go`: rm/unlink/rmdir/shred logic
- `rule_supabase.go`: Supabase CLI + ORM migrations (Alembic, Django, Prisma, Drizzle, Knex, Sequelize, Flyway, Liquibase, Rails, TypeORM, Goose)
- `rule_infra.go`: kubectl, gcloud, helm, docker, mongo*, terraform/tofu, gsutil, curl-vs-OpenSearch + cloud control-plane API
- `rule_git.go`: git operations that lose work or rewrite history (push -f / --delete, reset --hard, clean -f, checkout/restore pathspec, branch -D, stash drop/clear, filter-branch, filter-repo, bfg)
- `rule.go`: Rule interface + registry
- `decision.go`: Decision type + aggregation (no deny tier)
- `audit.go`: JSONL log + rotation + 0o600 perms
- `config.go`: TOML loader + trust system
