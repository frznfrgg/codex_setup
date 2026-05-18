---
description: Template command contract for ACT workers, validator, and auditor.
status: template
version: 0.1.0
---

# Project Commands

Command contract for ACT workers, validator, and auditor.

## Intent (Why)

We keep validation commands explicit so all agents verify work through the same project entry points.

## Rules (What)

1. Replace this template with concrete project commands before running `$act`.
2. Keep quality check, build, and test commands runnable from the repository root.
3. ACT code workers should run the quality check command after implementation when available.
4. The validator must run quality check before build.
5. The auditor must run the test command when assessing `TEST_STATUS`.
6. Update this file whenever package scripts, build commands, or release gates change.
7. Define any project-specific pre-commit or pre-push checks here instead of relying on agent memory.
8. Document generated-file expectations so workers know what should and should not be committed.

## Practices (How)

### Command Contract

Replace the placeholders with commands for this project.

```text
Quality check: <replace with lint/typecheck/format command>
Build: <replace with build command>
Tests: <replace with test command>
Release check: <optional release command or "not configured">
```

If a command is intentionally not configured, explain why and how validators should handle that case.

### Examples

```text
Quality check: pnpm check
Build: pnpm build
Tests: pnpm test
Release check: pnpm release:check
```

### Git Hygiene Contract

Replace placeholders with project-specific commands when available.

```text
Inspect working tree: git status --short
Inspect staged diff: git diff --staged
Secret scan: <replace with command or "manual diff inspection">
Generated files policy: <replace with tracked generated files, or "none">
```

Workers should inspect what they are about to commit and avoid staging unrelated changes.

## Meta

**Dependencies**:
- `development-conventions.md` - Defines worker implementation expectations.
- `testing-conventions.md` - Defines test-writing expectations.
