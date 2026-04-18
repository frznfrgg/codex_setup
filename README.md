# Workflow Setup README

This repository contains a workflow for turning an idea into an executable task, then running that task through specialized agents with validation and audit at the end.

## What This Setup Does

The flow is built around three stages:

1. Capture and refine project knowledge into steerings.
2. Turn a concrete idea into a task plan and subtasks.
3. Execute those subtasks through worker agents, then validate and audit the result.

The main execution skill is `$act`.

## Main Parts

### Skills

- `.agents/skills/act/`
  Runs a prepared task from `.tasks/{TASK-ID}/` through execution, validation, audit, and approval.

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
  Testing worker.

- `validator`
  Runs quality checks and build.

- `auditor`
  Reviews the completed implementation against the plan and test results.

## Recommended End-to-End Flow

### 1. Define project conventions

If the project does not already have clear conventions, use `$steering-specs-generator`.

Its goal is to produce steering files such as:

- `.memory-bank/steerings/development-conventions.md`
- `.memory-bank/steerings/testing-conventions.md`
- `.memory-bank/steerings/project-commands.md`

These are important because the worker agents depend on them.

### 2. Create a task

Once you have a concrete idea, convert it into a task folder:

```text
.tasks/{TASK-ID}/
├── plan.md
└── subtasks/
    ├── index.md
    ├── stt-001.md
    ├── stt-002.md
    └── ...
```

### 3. Split the task into valid subtasks

Each subtask should be assigned to exactly one Phase 2 worker agent:

- `code-implementer`
- `code-writer`
- `docs-writer`
- `test-writer`

Do not place `validator` or `auditor` in `subtasks/index.md`.

### 4. Execute the task with `$act`

`$act` is the orchestrator.

It will:

1. Validate the task structure.
2. Read `subtasks/index.md`.
3. Group subtasks by `seq`.
4. Run each `seq` wave in parallel.
5. Wait for the whole wave to finish before starting the next one.
6. Run `validator`.
7. Run `auditor`.
8. Create fix subtasks if validation or audit fails.
9. Repeat until the task passes or the flow must stop and escalate.

### 5. Review the final summary

If validation and audit pass, the flow ends with a short approval summary for the user.

## Required Files Before Running `$act`

At minimum, make sure these exist:

```text
.memory-bank/steerings/development-conventions.md
.memory-bank/steerings/testing-conventions.md
.memory-bank/steerings/project-commands.md
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
2. Write the task plan.
3. Create valid subtasks and `index.md`.
4. Switch to the task branch.
5. Run `$act`.
6. Wait for validation and audit.
7. Review the final summary and approve.

## Common Mistakes

- Missing steering files in `.memory-bank/steerings/`
- Using display names instead of exact agent IDs
- Forgetting `seq={N}` in `subtasks/index.md`
- Putting `validator` or `auditor` into Phase 2 subtasks
- Running execution without a proper task folder

## If You Want to Dive Deeper

Start in this order:

1. Read `.agents/skills/act/SKILL.md`
2. Read `.agents/skills/steering-specs-generator/SKILL.md`
3. Read `.codex/agents/*.toml`
4. Create one small test task and run the flow on it

That is the fastest way to understand how the pieces fit together.
