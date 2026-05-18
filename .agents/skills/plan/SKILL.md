---
name: plan
description: Create or update a flat .plans/PLAN_{id-or-feature-name}.md planning artifact before ACT decomposition. Use when turning a feature/project idea into a structured source-of-truth plan, before creating .tasks artifacts. Does not create ACT subtasks; use split-to-tasks after the plan is READY_FOR_SPLIT.
---

# Plan

## Overview

Create and maintain the source-of-truth planning artifact for a feature, project, migration, or substantial change.

This skill owns exploratory planning. It does **not** create ACT task folders. ACT task generation belongs to `$split-to-tasks`.

Output path convention:

```text
.plans/PLAN_{id-or-feature-name}.md
```

## When To Use

- You are discussing or shaping a feature before ACT execution.
- You need a durable source-of-truth plan.
- You need to capture scope, non-goals, architecture, contracts, risks, verification, and open questions.
- You need to decide whether the work is ready to split into ACT tasks.

When intent or product scope is too unclear, stop and tell the user to run `$interview-me` or `$idea-refine` first. Do not invoke those skills automatically.

## Inputs

Use the user's prompt and any provided files. If no plan path is given, create a slug from the feature name and write:

```text
.plans/PLAN_{slug}.md
```

If a plan path is given, update that file in place.

Read project context as needed:

- `.memory-bank/index.md`
- relevant `.memory-bank/steerings/*.md`
- relevant `.memory-bank/architecture/*.md`
- relevant source files
- existing `.plans/PLAN_*.md` if continuing prior planning

Do not load optional module skills by default. First classify whether their trigger applies, then read only the relevant skill file.

## Optional Module Skills

Use these as lazy technical modules. Read their `SKILL.md` only when the trigger is true.

| Trigger | Read and apply |
|---|---|
| Public API, module boundary, schema, CLI contract, component props, ML/data contract, integration contract | `.agents/skills/api-and-interface-design/SKILL.md` |
| Framework/library/runtime behavior depends on exact current docs or versions | `.agents/skills/source-driven-development/SKILL.md` |
| CI, build automation, release gates, deployment automation, test runners in CI, or project bootstrap command contracts | `.agents/skills/ci-cd-and-automation/SKILL.md` |
| Replacing, removing, migrating, deprecating, or consolidating behavior, APIs, data, dependencies, or systems | `.agents/skills/deprecation-and-migration/SKILL.md` |
| Significant architectural decision, public API, user-facing behavior change, major dependency choice, or decision expensive to reverse | `.agents/skills/documentation-and-adrs/SKILL.md` |

Record module conclusions in the plan. Do not paste whole module checklists into the plan unless directly needed.

## Planning Process

### Step 1: Establish Current Understanding

Restate what is being built and why. If you cannot state the objective, target user/system, success criteria, and non-goals, ask concise clarifying questions or stop and recommend `$interview-me` / `$idea-refine`.

### Step 2: Inspect Relevant Context

Read only context that materially affects the plan. Prefer existing project steerings and architecture notes over generic assumptions.

Capture references explicitly so `$split-to-tasks` can preserve them in ACT artifacts.

### Step 3: Classify Technical Modules

Determine which optional modules apply. For each triggered module:

1. Read its `SKILL.md`.
2. Apply only relevant guidance.
3. Add a concise conclusion to the plan.
4. Add follow-up requirements or open questions if the module exposes a gap.

### Step 4: Draft Or Update The Plan

Create or update the `.plans/PLAN_*.md` file. Keep it structured enough that `$split-to-tasks` can consume it without guessing.

### Step 5: Set Split Readiness

At the end, set:

```text
READY_FOR_SPLIT: yes
```

only when all blocking decisions are made and the plan is stable enough to decompose into ACT subtasks.

Use:

```text
READY_FOR_SPLIT: no
```

when open questions, unresolved trade-offs, missing refs, or missing module conclusions would force `$split-to-tasks` to invent requirements.

## Plan Template

```markdown
# PLAN: [Feature / Project / Change Name]

## Status

READY_FOR_SPLIT: no

## Objective

[What we are trying to accomplish.]

## Source Of Truth

- [User prompt / discussion summary]
- [.memory-bank/...]
- [Relevant source files]
- [External docs if source-driven requirements apply]

## Current Understanding

[Concise restatement of the desired behavior and context.]

## Scope

- [In scope]

## Non-Goals

- [Explicitly out of scope]

## Constraints

- [Technical, product, schedule, compatibility, security, deployment, or data constraints]

## Proposed Approach

[Architecture and implementation strategy.]

## Interfaces And Contracts

[APIs, schemas, module boundaries, CLI args, component props, ML/data contracts, integration contracts. Write "None" if not applicable.]

## Source-Driven Requirements

[Official docs and version-sensitive implementation constraints. Write "None" if not applicable.]

## CI / Automation Impact

[Build, test, release, CI, deployment, command-contract impact. Write "None" if not applicable.]

## Migration / Deprecation Impact

[Replacement/removal/compatibility/rollback/migration concerns. Write "None" if not applicable.]

## Documentation / ADR Requirements

[Docs, ADRs, changelog, memory-bank updates. Write "None" if not applicable.]

## Security / Test / Release Risks

- [Risk] - [Mitigation]

## Verification Strategy

- [Quality checks]
- [Tests]
- [Manual or release checks if needed]

## Decisions

- [Decision] - [Rationale]

## Open Questions

- [Question or "None"]

## Split Readiness

READY_FOR_SPLIT: no

Blocking questions:
- [Question or "None"]

Required before split:
- [Requirement or "None"]
```

## Rules

- Keep `.plans/PLAN_*.md` as the planning source of truth.
- Do not create `.tasks/` artifacts from this skill.
- Do not silently invent product requirements or architecture decisions.
- Do not auto-run `$interview-me` or `$idea-refine`; recommend them only when the plan is too vague.
- Use optional module skills only after their trigger is detected.
- Prefer concise module conclusions over copied checklists.
- If updating an existing plan, preserve useful prior decisions and explicitly revise stale sections.

## Output

Finish with:

- the plan path
- `READY_FOR_SPLIT: yes/no`
- remaining blockers, if any
- recommended next command, usually `$split-to-tasks <plan-path> <task-id>` when ready
