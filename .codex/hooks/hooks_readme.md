# Codex Hooks

This scaffold installs project-local lifecycle and Bash safety hooks through `.codex/hooks.json`.

The hooks use one Go binary for all configured events:

- `SessionStart` injects compact runtime facts such as the start source, working directory, permission mode, and model.
- `SubagentStart` injects the subagent identity/type and current runtime facts.
- `PreToolUse` blocks risky Bash commands.
- `UserPromptSubmit` records one-shot approvals for prompts shaped as `approve-risk <hash>`.
- `SubagentStop` and `Stop` emit valid no-op JSON. Codex does not currently support `additionalContext` for these events, and a block decision would force a continuation.

Lifecycle context is deliberately programmatic and compact. It never includes transcript contents, the last assistant message, or copied skill instructions.

Build the Go engine before relying on the hook:

```bash
bash .codex/hooks/bash-guard/build.sh
```

Then open `/hooks` in Codex and review/trust the project hooks.

Codex uses `SubagentStart` and `SubagentStop` for before/after-agent lifecycle integration. `BeforeAgent` and `AfterAgent` are not Codex hook event names.

Risk approvals are scoped to the current Codex `session_id` and the exact command SHA-256. They expire after 30 minutes and are consumed after one matching retry.
