# Workflow Setup README

This repository contains a workflow for turning an idea into an executable task, then running that task through specialized agents with validation and audit at the end.

## What This Setup Does

The flow is built around four stages:

1. Capture and refine project knowledge into steerings.
2. Create a durable `.plans/PLAN_*.md` planning artifact.
3. Split an approved plan into ACT task artifacts.
4. Execute those subtasks through worker agents, then validate and audit the result.

The main execution skill is `$act`. The planning pipeline is `$plan` followed by `$split-to-tasks`.

## Main Parts

### Skills

- `.agents/skills/act/`
  Runs a prepared task from `.tasks/{TASK-ID}/` through execution, validation, audit, and approval.

- `.agents/skills/plan/`
  Creates or updates `.plans/PLAN_{id-or-feature-name}.md` as the source-of-truth planning artifact. It can lazily consult technical module skills, but it does not create ACT subtasks.

- `.agents/skills/split-to-tasks/`
  Converts a ready `.plans/PLAN_*.md` into ACT-ready `.tasks/{TASK-ID}/plan.md`, `subtasks/index.md`, and `stt-*.md` files.

- `.agents/skills/interview-me/`
  Optional pre-plan skill for clarifying intent when the ask is underspecified.

- `.agents/skills/idea-refine/`
  Optional pre-plan skill for refining a vague idea into a concrete one-pager.

- `.agents/skills/api-and-interface-design/`
  Lazy planning module for APIs, module boundaries, schemas, CLI contracts, integration contracts, and data/model contracts.

- `.agents/skills/source-driven-development/`
  Lazy planning/development module for framework, library, runtime, and platform decisions that need official documentation.

- `.agents/skills/ci-cd-and-automation/`
  Standalone skill and lazy planning module for CI, build automation, release gates, deployment automation, and command contracts.

- `.agents/skills/deprecation-and-migration/`
  Standalone skill and lazy planning module for replacing, removing, deprecating, migrating, or consolidating existing behavior.

- `.agents/skills/documentation-and-adrs/`
  Standalone skill and lazy planning module for ADRs, docs, README, changelog, and memory-bank documentation requirements.

- `.agents/skills/code-review-and-quality/`
  Standalone manual code review skill. ACT automated review uses the `code-reviewer` agent.

- `.agents/skills/security-and-hardening/`
  Standalone manual security guidance skill. ACT automated security review uses the `security-auditor` agent.

- `.agents/skills/onboard/`
  Helps an agent understand the project structure and current state before starting work.

- `.agents/skills/steering-specs-generator/`
  Extracts tacit knowlege and conventions and turns them into steering files and action items.

### Agents

Configured in `.codex/agents/`.

- `code-implementer`
  Main implementation worker for feature/refactor/fix subtasks.

- `code-writer`
  Narrow code-edit worker for direct code-writing subtasks.

- `docs-writer`
  Documentation worker.

- `test-writer`
  Testing worker that writes must-have tests for assigned subtasks.

- `validator`
  Runs quality checks and build.

- `code-reviewer`
  Phase 3 read-only reviewer that checks general engineering quality, correctness, maintainability, architecture fit, reliability, and regression risk.

- `test-auditor`
  Phase 3 read-only reviewer that checks whether tests are meaningful, sufficient, and aligned with the plan.

- `security-auditor`
  Conditional Phase 3 read-only reviewer that checks practical security risks when a task touches security-sensitive surfaces.

- `auditor`
  Thin final Phase 3 gate that aggregates validator context plus code-reviewer, test-auditor, and security-auditor reports into one authoritative audit decision.

## Recommended End-to-End Flow

### 1. Define project conventions

If the project does not already have clear conventions, use `$steering-specs-generator`.

Its goal is to produce steering files such as:

- `.memory-bank/steerings/development-conventions.md`
- `.memory-bank/steerings/testing-conventions.md`
- `.memory-bank/steerings/project-commands.md`

These are important because the worker agents depend on them.

### 2. Create or update a plan

Use optional pre-plan skills only when needed:

- `$interview-me` if intent, user, success criteria, constraints, or non-goals are unclear.
- `$idea-refine` if the idea or MVP scope is still fuzzy.

Then use `$plan` to create or update:

```text
.plans/PLAN_{id-or-feature-name}.md
```

The plan is the exploratory source of truth. It should end with `READY_FOR_SPLIT: yes` only when decisions are stable enough for ACT decomposition.

### 3. Split the plan into a task

Once the plan is ready, use `$split-to-tasks` with the plan path and task ID:

```text
$split-to-tasks .plans/PLAN_feature-name.md TASK-001
```

It creates:

```text
.tasks/{TASK-ID}/
├── plan.md
└── subtasks/
    ├── index.md
    ├── stt-001.md
    ├── stt-002.md
    └── ...
```

### 4. Ensure valid subtasks

Each subtask should be assigned to exactly one Phase 2 worker agent:

- `code-implementer`
- `code-writer`
- `docs-writer`
- `test-writer`

Do not place `validator`, `code-reviewer`, `test-auditor`, `security-auditor`, or `auditor` in `subtasks/index.md`.

### 5. Execute the task with `$act`

`$act` is the orchestrator.

It will:

1. Validate the task structure.
2. Read `subtasks/index.md`.
3. Group subtasks by `seq`.
4. Run each `seq` wave in parallel.
5. Wait for the whole wave to finish before starting the next one.
6. Run `validator`.
7. Run `code-reviewer` and `test-auditor`, plus `security-auditor` when the task touches security-sensitive surfaces.
8. Run the final `auditor` to aggregate review reports into one audit decision.
9. Create fix subtasks if validation, code review, test audit, security audit, or final audit fails.
10. Repeat until the task passes or the flow must stop and escalate.

### 6. Review the final summary

If validation and audit pass, the flow ends with a short approval summary for the user.

## Required Files Before Running `$act`

At minimum, make sure these exist:

```text
.memory-bank/steerings/development-conventions.md
.memory-bank/steerings/testing-conventions.md
.memory-bank/steerings/project-commands.md
.plans/PLAN_{id-or-feature-name}.md
.tasks/{TASK-ID}/plan.md
.tasks/{TASK-ID}/subtasks/index.md
```

Also make sure you are on a task-specific git branch before execution.

## How `subtasks/index.md` Works

Each line should follow this shape:

```markdown
- [ ] stt-001 | code-implementer | feature | seq=1 / Build feature shell
      Scaffold the core structure for the feature.
```

Meaning:

- `[ ]` = pending
- `stt-001` = subtask file ID
- `code-implementer` = exact runtime agent name
- `feature` = category
- `seq=1` = execution wave
- `Build feature shell` = title

Rules:

- same `seq` means those subtasks may run in parallel
- higher `seq` waits for lower `seq`
- `seq` must be a positive integer
- agent names must be exact

## What Each Worker Receives

All Phase 2 workers use the same contract:

- `task_id`
- `task_root`
- `plan_path`
- `subtask_path`

That means each worker reads:

1. the task plan
2. the specific subtask
3. any references listed inside that subtask

## Example Usage Pattern

For a normal piece of work:

1. Create project steerings if needed.
2. Optionally use `$interview-me` or `$idea-refine`.
3. Use `$plan` to create or update `.plans/PLAN_*.md`.
4. Review the plan and set `READY_FOR_SPLIT: yes`.
5. Use `$split-to-tasks` to create `.tasks/{TASK-ID}/`.
6. Switch to the task branch.
7. Run `$act`.
8. Wait for validation and audit.
9. Review the final summary and approve.

## Common Mistakes

- Missing steering files in `.memory-bank/steerings/`
- Using display names instead of exact agent IDs
- Forgetting `seq={N}` in `subtasks/index.md`
- Putting `validator`, `code-reviewer`, `test-auditor`, `security-auditor`, or `auditor` into Phase 2 subtasks
- Running `$split-to-tasks` before `.plans/PLAN_*.md` has `READY_FOR_SPLIT: yes`
- Running execution without a proper task folder

## If You Want to Dive Deeper

Start in this order:

1. Read `.agents/skills/act/SKILL.md`
2. Read `.agents/skills/plan/SKILL.md`
3. Read `.agents/skills/split-to-tasks/SKILL.md`
4. Read `.agents/skills/steering-specs-generator/SKILL.md`
5. Read `.codex/agents/*.toml`
6. Create one small test task and run the flow on it

That is the fastest way to understand how the pieces fit together.
