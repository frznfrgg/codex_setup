---
description: Template testing conventions consumed by ACT test workers.
status: template
version: 0.1.0
---

# Testing Conventions

Testing conventions consumed by ACT test workers.

## Intent (Why)

We write tests that catch meaningful regressions without creating brittle or low-value coverage.

## Rules (What)

1. Replace this template with project-specific testing rules before running `$act`.
2. Prioritize tests for business logic, boundary behavior, and high-risk failure paths.
3. Keep tests deterministic and independent of local machine state.
4. Use the project's existing test framework and fixtures.
5. Prefer focused tests over broad snapshots unless snapshots are already the local convention.
6. Avoid testing implementation details that are not part of the behavior contract.
7. Add regression tests for bugs fixed by a task when the behavior can be isolated.
8. Keep slow, networked, or credentialed checks out of the default test command unless the project explicitly requires them.

## Practices (How)

### Behavior Tests

Name tests around observable behavior.

```text
Good: rejects invalid input before persistence
Avoid: calls private helper with argument x
```

### Failure Paths

Cover errors that would matter in production.

```text
- invalid user input
- unavailable dependency
- persistence failure
- permission or authorization failure
```

### Test Scope

Match test scope to risk.

```text
Unit: pure logic and validation
Integration: module boundaries and adapters
End-to-end: critical user workflows only
```

## Meta

**Dependencies**:
- `project-commands.md` - Defines the default test command.
- `development-conventions.md` - Defines implementation boundaries that tests should respect.
