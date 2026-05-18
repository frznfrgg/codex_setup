---
name: split-to-tasks
description: "Convert an approved .plans/PLAN_*.md file into ACT-ready .tasks/{TASK-ID}/plan.md, subtasks/index.md, and stt-*.md artifacts. Use after a plan has READY_FOR_SPLIT: yes. This skill is deterministic decomposition, not exploratory planning."
---

# Split To Tasks

## Overview

Convert a ready planning artifact into the exact task structure consumed by `$act`.

Input:

```text
.plans/PLAN_{id-or-feature-name}.md
```

Output:

```text
.tasks/{TASK-ID}/
├── plan.md
└── subtasks/
    ├── index.md
    ├── stt-001.md
    ├── stt-002.md
    └── ...
```

This skill is an execution compiler. It does not make new product decisions, redesign architecture, or answer unresolved planning questions.

## When To Use

- A `.plans/PLAN_*.md` file exists.
- The plan contains `READY_FOR_SPLIT: yes`.
- The user wants ACT-compatible task artifacts.

Do not use this skill when the plan is still exploratory or has blocking questions.

## Required Inputs

The user should provide:

- `plan_path`: path to `.plans/PLAN_*.md`
- `task_id`: target task ID, such as `TASK-001`, `GDS-019`, or another project convention

If `task_id` is missing, ask for it.

## Required Reference

Read `.agents/skills/act/SKILL.md` before writing artifacts. The generated files must match ACT expectations exactly.

## Readiness Gate

Before creating files:

1. Read `plan_path`.
2. Confirm it contains `READY_FOR_SPLIT: yes`.
3. Confirm no blocking open questions remain.
4. Confirm required plan sections are specific enough to produce subtasks without inventing decisions.

Stop and return the missing requirements if the plan is not ready.

## Validation-Only Module Checks

Use these only to validate the plan before decomposition. Do not invent missing content during split.

| If the plan says... | Validate using |
|---|---|
| APIs, schemas, public interfaces, module boundaries, CLI contracts, ML/data contracts, or integration contracts are involved | `.agents/skills/api-and-interface-design/SKILL.md` |
| Framework/library/runtime behavior is version-sensitive or official docs are required | `.agents/skills/source-driven-development/SKILL.md` |
| CI/build/release automation or command contracts are involved | `.agents/skills/ci-cd-and-automation/SKILL.md` |
| Migration, replacement, deprecation, removal, or compatibility is involved | `.agents/skills/deprecation-and-migration/SKILL.md` |
| Docs or ADRs are required | `.agents/skills/documentation-and-adrs/SKILL.md` |

If a validation module reveals that the plan lacks required decisions, stop and send the user back to `$plan`.

## Decomposition Process

### Step 1: Create Execution Plan

Write `.tasks/{TASK-ID}/plan.md` as the execution-focused version of the source plan.

It should include:

- objective
- scope and non-goals
- implementation constraints
- acceptance criteria
- required review gates
- verification strategy
- references

Do not include unresolved brainstorming.

### Step 2: Build Dependency Graph

Map what depends on what. Implementation order should follow the dependency graph.

Prefer vertical slices when possible: each task should produce working, testable progress. Use foundational horizontal subtasks only when they unblock later vertical slices.

### Step 3: Create Subtasks

Each subtask file must contain:

```markdown
# stt-00N [Title]

## Goal

[Single focused outcome.]

## Tech Requirements

- [Concrete constraints and implementation requirements]

## Acceptance Criteria

- [ ] [Specific, testable condition]

## Refs

- [.tasks/{TASK-ID}/plan.md]
- [.memory-bank/...]
- [.plans/PLAN_*.md]
- [External docs if relevant]
```

Assign each subtask to exactly one Phase 2 ACT worker:

- `code-implementer` for normal feature/refactor/fix implementation
- `code-writer` for narrow direct code edits
- `docs-writer` for documentation, memory-bank, ADR, README, or changelog tasks
- `test-writer` for writing tests only

Never assign Phase 3 agents to subtasks:

- `validator`
- `code-reviewer`
- `test-auditor`
- `security-auditor`
- `auditor`

### Step 4: Create `subtasks/index.md`

Use exact ACT index format:

```markdown
# {TASK-ID} Subtasks

- [ ] stt-001 | code-implementer | feature | seq=1 / [Title]
      [One-line description.]

- [ ] stt-002 | test-writer | test | seq=2 / [Title]
      [One-line description.]
```

Rules:

- `seq` is the only scheduling mechanism.
- Same `seq` may run in parallel.
- Higher `seq` waits for lower `seq`.
- Every subtask must have a positive integer `seq`.
- Use gaps only when there is a real reason.
- Do not infer hidden dependencies from prose; encode order with `seq`.

### Step 5: Size Check

Prefer S/M subtasks:

| Size | Files | Action |
|---|---:|---|
| XS | 1 | acceptable |
| S | 1-2 | ideal |
| M | 3-5 | acceptable |
| L | 5-8 | split if possible |
| XL | 8+ | must split |

Split a task further when:

- it has more than three acceptance criteria
- the title contains "and" for unrelated work
- it crosses independent subsystems
- it cannot be verified in one focused pass

### Step 6: Final Validation

Before finishing:

- [ ] `.tasks/{TASK-ID}/plan.md` exists
- [ ] `subtasks/index.md` exists
- [ ] every `stt-*.md` listed in the index exists
- [ ] every subtask has Goal, Tech Requirements, Acceptance Criteria, and Refs
- [ ] every index line uses an allowed Phase 2 worker
- [ ] no Phase 3 agents appear in `subtasks/index.md`
- [ ] every subtask declares `seq=N`
- [ ] every required docs/ADR/test/security/CI/migration concern from the plan is represented as a subtask or review gate

## Output

Finish with:

- task folder path
- number of subtasks created
- sequence wave summary
- any plan concerns that were preserved as review gates
- next command: `$act {TASK-ID}`
