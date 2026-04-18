---
name: onboard
description: Onboard the AI coding agent to the project structure and current state. Use when starting work in a repository, after context loss, before making code changes, or whenever Codex needs a fast understanding of the architecture, recent development status, `.memory-bank` notes, and Serena memories.
---

# Onboard

## Overview

Build project context quickly before implementation. Read the project's memory bank, inspect recent changes, supplement with Serena memories when available, and persist key facts for future sessions when they do not exist yet.

## Workflow

Use these tools when available: `Bash`, `Task`, `Read`, `mcp__serena__check_onboarding_performed`, `mcp__serena__list_memories`, `mcp__serena__read_memory`, `mcp__serena__write_memory`.

Perform the following steps in order:

1. Read `.memory-bank/index.md` to understand the overall project structure and architecture.
2. Review recent code changes to understand the current development state.
3. Check whether Serena memories exist with `mcp__serena__check_onboarding_performed` and `mcp__serena__list_memories`.
4. Read available Serena memories with `mcp__serena__read_memory` to supplement project understanding.
5. If no Serena memories exist, use `mcp__serena__write_memory` to save key project information for future sessions.

When writing Serena memory, capture the minimum durable context:

- project purpose
- important commands
- coding conventions
- high-level structure

Prefer recent, concrete signals when reviewing project state:

- `git status`
- recent commits
- files touched recently
- open work tracked in `.memory-bank`

If Serena tools are unavailable in the current environment, continue with local project artifacts and explicitly note that Serena memory sync could not be checked or updated.

## Output

After completing the workflow, say `Onboarding complete` and briefly describe:

- the project overview
- the current status
- any missing context or unavailable memory tooling
