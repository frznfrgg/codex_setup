---
description: Template implementation conventions consumed by ACT code workers.
status: template
version: 0.1.0
---

# Development Conventions

Implementation conventions consumed by ACT worker agents.

## Intent (Why)

We keep implementation work scoped, maintainable, and consistent with the project's existing architecture.

## Rules (What)

1. Replace this template with project-specific rules before running `$act`.
2. Keep changes scoped to the assigned task or subtask.
3. Follow established project patterns before introducing new abstractions.
4. Add abstractions only when they protect a real boundary or remove concrete duplication.
5. Prefer small, readable modules over broad rewrites.
6. Preserve public contracts unless the task explicitly changes them.
7. Keep project-specific secrets, credentials, and local machine paths out of committed files.
8. Document non-obvious architectural decisions in `.memory-bank/architecture/`.

## Practices (How)

### Scoped Changes

Read the plan, subtask, and referenced files before editing. Limit implementation to the requested outcome.

```text
Good: update the service and tests named by the subtask.
Avoid: refactor unrelated callers because they look nearby.
```

### Existing Patterns First

Use the codebase's current style, framework choices, and helper APIs.

```text
Before adding a helper, search for an existing local helper that already solves the same problem.
```

### Architecture Notes

When a decision needs durable context, record it in `.memory-bank/architecture/` and cite it from future tasks.

```markdown
[.memory-bank/architecture/example-boundary.md]: Explains the boundary used by this implementation.
```

## Meta

**Dependencies**:
- `project-commands.md` - Defines quality checks workers should run.
