# Codex Hooks

This scaffold installs a project-local Bash safety hook through `.codex/hooks.json`.

The hook uses one Go binary for both configured events:

- `PreToolUse` blocks risky Bash commands.
- `UserPromptSubmit` records one-shot approvals for prompts shaped as `approve-risk <hash>`.

Build the Go engine before relying on the hook:

```bash
bash .codex/hooks/bash-guard/build.sh
```

Then open `/hooks` in Codex and review/trust the project hooks.

Risk approvals are scoped to the current Codex `session_id` and the exact command SHA-256. They expire after 30 minutes and are consumed after one matching retry.
