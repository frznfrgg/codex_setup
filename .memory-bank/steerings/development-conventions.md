---
description: Template implementation conventions consumed by ACT code workers.
status: template
version: 0.2.0
---

# Development Conventions

Implementation and version-control conventions consumed by project agents.

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
9. Keep commits atomic when ACT workers commit: one logical subtask completion per commit.
10. Separate formatting-only cleanup, refactors, tests, docs, and behavior changes unless the subtask explicitly couples them.
11. Do not commit generated files, dependency folders, build output, or environment files unless the project explicitly expects them.

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

### Version Control Discipline

Treat commits as reviewable save points.

```text
Format: <type>(<scope>): <description>

Task work: feat(TASK-001): add session renewal for stt-002
Non-task work: feat(hooks): add lifecycle runtime context
Avoid: task(TASK-001): complete stt-002
Avoid: misc updates
```

Use Conventional Commits for every commit:

- Choose a lowercase type that matches the change: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`, or `chore`.
- For ACT work, use the exact `TASK-ID` as the scope, describe the delivered change, and include the completed `stt-xxxx` identifier in the description.
- For work outside an ACT task, use the affected feature, component, or subsystem as the scope.
- Write a concise imperative description without a trailing period.
- Mark breaking changes with `!` before the colon and explain them in a `BREAKING CHANGE:` footer.

Before committing, inspect the diff and ensure the commit contains only files required by the current task or subtask. If unrelated modified files exist, leave them alone.

```text
Check:
- no secrets or local paths
- no dependency folders or build output
- no formatting-only churn mixed into behavior changes
- generated files included only when project conventions require them
```

## Meta

**Dependencies**:
- `project-commands.md` - Defines quality checks workers should run.
