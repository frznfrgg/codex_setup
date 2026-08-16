# Repository Purpose

<!-- Replace this placeholder with the purpose of the repository. -->

[Describe what this repository is for, who or what it serves, and its primary responsibilities.]

# Hard Constraints

<!-- Replace this placeholder with the repository's non-negotiable constraints. -->

[Describe rules that agents must never violate, or write "None" if no additional hard constraints apply.]

# Context Map

- `.memory-bank/` contains durable project knowledge.
  - `index.md` is the entry point and directory of available project context.
  - `steerings/development-conventions.md` defines implementation rules and coding conventions.
  - `steerings/testing-conventions.md` defines testing rules and expectations.
  - `steerings/project-commands.md` defines the authoritative quality, build, test, and release commands.
  - `steerings/index.md` lists additional project-specific steering documents.
  - `architecture/` contains durable architectural decisions, boundaries, constraints, and relevant technology choices.
  - `references/` contains curated technical sources that can be reused across tasks.
  - `backlog/` contains follow-up work discovered while documenting or working on the project.
- `.plans/` contains exploratory source-of-truth plans that are not directly executable.
- `.tasks/{TASK-ID}/` contains ACT execution plans, subtasks, reviews, and audits.
- `.agents/skills/` contains workflow instructions loaded when invoked or when their trigger applies.
- `.codex/agents/` contains specialized ACT worker and reviewer profiles.
- `.codex/hooks/` contains project-local lifecycle and safety hooks.

# Context Loading Rules

1. At the beginning of substantial work, read `.memory-bank/index.md`.
2. Before changing code, read:
   - `.memory-bank/steerings/development-conventions.md`
   - `.memory-bank/steerings/project-commands.md`
   - relevant files under `.memory-bank/architecture/`
3. Before writing or reviewing tests, also read `.memory-bank/steerings/testing-conventions.md`.
4. Read additional steering documents through `.memory-bank/steerings/index.md` when they are relevant to the task.
5. Load relevant skills, architecture notes, references, and other project instructions when the task requires them.
6. During ACT work, read the task plan, the assigned subtask, and every reference declared by that subtask.
