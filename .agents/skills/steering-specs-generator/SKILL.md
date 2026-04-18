---
name: steering-specs-generator
description: Extract tacit engineering knowledge through guided interviews and generate structured steerings. Use when user mentions "steerings", "tacit knowledge", "conventions", "engineering practices", "interview", or wants to document team/project knowledge. Also activates when user asks for "steerings for X", "document X conventions", "continue steerings", "resume interview", or wants to extract knowledge about a specific topic. Supports reviewing and transforming existing steerings to standard format. Auto-detects existing sessions and offers to continue incomplete ones.
---

# Steering Specs Generator

Conducts context-aware interviews to extract tacit engineering knowledge and generate agent-readable steerings. Format: **Intent (Why) → Rules (What) → Practices (How) → Meta**.

Supports **predefined packs** (8 areas) and **custom topics** (user-specified).

**Flow overview:** See [flow-diagram.md](flow-diagram.md) for visual representation.

## Prerequisites

- [pack-reference.md](pack-reference.md) — Topic areas and questions
- [steering-template.md](steering-template.md) — Output format
- Access to an "ask user" tool and a "run subagent task" tool

## Tooling Compatibility (Claude Code ↔ Codex CLI)

This skill was originally authored for **Claude Code** tool names (e.g. `AskUserQuestion`, `Task`). To keep it portable, treat these as *capabilities* and map them to your runtime:

- **Ask user (blocking input)**
  - Claude Code: `AskUserQuestion`
  - Codex CLI: `request_user_input`
  - If choice options aren't supported by your tool, present choices in text and ask for an index/label.
- **Run subagents / parallel work**
  - Claude Code: `Task` (+ its output/wait mechanism)
  - Codex CLI: `spawn_agent` + `wait` (+ optionally `send_input` to clarify) + `close_agent` when done
- **Repo scanning / file IO**
  - Claude Code: `Read` / `Write` / `Glob`
  - Codex CLI: `exec_command` for search/listing, `apply_patch` for edits (or your runtime’s native file tools)

In the rest of this doc:
- `ASK_USER(...)` means "use your environment’s ask-user tool".
- "Task agent" means "spawn a subagent and wait for its output".

## Mode Selection

| Keywords | Mode |
|----------|------|
| "review steerings", "transform steerings", "fix format" | → Review Mode (Step R) |
| "continue steerings", "continue session", "resume interview" | → Interview Mode (Step 0) with session check |
| "steerings", "tacit knowledge", "interview", "conventions" | → Interview Mode (Step 0) |

---

## Interview Flow

### Pack File State Model

Pack progress is tracked via **YAML frontmatter** in each `{packId}.md` file. This frontmatter is the source of truth for session continuation and pack readiness.

**Supported statuses:**

- `pending` — Pack was selected and the file exists, but work has not started
- `in_progress` — Parent-led interview and extraction is underway for this pack
- `complete` — Interview is finished, extraction is complete, and the pack is ready for synthesis
- `blocked` — Progress cannot continue because required input or context is missing
- `skipped` — Pack was intentionally excluded from this session
- `error` — Agent failed or produced invalid output and needs retry/manual repair

**Required frontmatter shape:**

```yaml
---
pack_id: architecture-design-invariants
pack_name: Architecture & Design Invariants
pack_type: predefined
status: in_progress
interview_mode: interactive
question_count: 5
answered_count: 2
created_at: 2026-04-18T20:30:00Z
updated_at: 2026-04-18T20:42:00Z
completed_at:
source_reports:
  - .sessions/my-session/explore-docs-conventions.md
  - .sessions/my-session/explore-repo-context.md
---
```

**Status lifecycle rules:**

1. When a session starts and packs are selected, create `{sessionsPath}{sessionId}/{packId}.md` with frontmatter and `status: pending`
2. When the parent workflow begins the live interview for a pack, update `status: in_progress` and `updated_at`
3. During parent-led interview execution, update `answered_count` and `updated_at` as responses are collected
4. When extraction is finished and the file is valid, set `status: complete` and `completed_at`
5. If required input/context is missing, set `status: blocked`
6. If the agent fails or the output is invalid, set `status: error`
7. If the user explicitly excludes a pack, set `status: skipped`

### Step 0: Check for Existing Sessions

Before configuring paths, check if sessions already exist in the repo:

1. **Scan common session directories:**
   - `.sessions/`
   - `sessions/`
   - `docs/sessions/`

2. **If sessions found**, present `ASK_USER(...)`:
```yaml
questions:
  - question: "Found existing session(s). Continue or start fresh?"
    header: "Session"
    options:
      - label: "Continue {sessionId}"
        description: "Resume incomplete session ({N} of {M} packs done)"
      - label: "Start new session"
        description: "Create fresh session with new ID"
```

3. **If continuing existing session:**
   - Set `sessionsPath` to parent directory of found session
   - Set `sessionId` to selected session name
   - Scan `{sessionsPath}{sessionId}/` for completed pack files
   - **Pack state detection:** Read YAML frontmatter from each `{packId}.md` file and use `status` as the source of truth
   - **Resume rules by status:**
     - `complete` → skip this pack, but only if the file also contains `## Conventions` and `## Action Items`
     - `pending` → queue this pack
     - `in_progress` → queue this pack first and resume from partial state using existing answers when possible
     - `blocked` → surface to the user and ask whether to retry or skip
     - `error` → surface to the user and ask whether to retry or regenerate
     - `skipped` → exclude unless the user explicitly re-enables it
   - Store `completedPacks[]` list (pack IDs to skip)
   - Read `explore-docs-conventions.md` and `explore-repo-context.md` paths if they exist
   - Skip to Step 3 (Pack Question Preparation), filtering out completed packs

4. **If starting new or no sessions found** → Continue to Step 0a

### Step 0a: Configure Output Paths

Ask user to confirm paths using `ASK_USER(...)`:

```yaml
questions:
  - question: "Where should steering files be saved?"
    header: "Steerings"
    options: ["./steerings/", "docs/steerings/", ".memory-bank/steerings/", "Custom"]
  - question: "Where should session files be saved?"
    header: "Sessions"
    options: ["./sessions/", ".sessions/", "docs/sessions/", "Custom"]
  - question: "Where should action items be saved?"
    header: "Backlog"
    options: ["./backlog/", "Same as steerings parent", ".backlog/", "Custom"]
```


**Store:** `steeringsPath`, `sessionsPath`, `backlogPath`. Create directories if needed.

**Defaults:** `steerings/`, `sessions/`, `backlog/`

### Step 0b: Generate `sessionId` - short words id

### Step 1: Define Topics

**Custom topic detected** (patterns: "steerings for X", "document X conventions"):
- If clear → Generate `packId`, `packName`, `packType: "custom"`, `customTopicDescription`
- If broad → Clarify with `ASK_USER(...)` (aspects, level, scope)

**No custom topic** → Present 8 predefined packs as multi-select:

### Step 1a: Choose Interview Mode

Ask user to select interview mode using `ASK_USER(...)`:

```yaml
questions:
  - question: "Interview mode preference?"
    header: "Mode"
    options:
      - label: "Interactive (default)"
        description: "Answer questions one-by-one with discussion"
      - label: "Fast mode"
        description: "Answer all questions at once, faster execution"
```

**Store:** `interviewMode` ("interactive" or "fast")

| Pack | ID |
|------|----|
| Codebase Topology & Ownership | `codebase-topology-ownership` |
| Architecture & Design Invariants | `architecture-design-invariants` |
| Business Domain Contracts | `business-domain-contracts` |
| Quality & Style Assurance | `quality-style-assurance` |
| Testing & Verification Strategy | `testing-verification-strategy` |
| Risk & Historical Landmines | `risk-historical-landmines` |
| Security, Data & Compliance | `security-data-compliance` |
| Delivery Lifecycle & Change Flow | `delivery-lifecycle-change-flow` |

### Step 2: Discovery (Parallel Explore)

Run TWO Explore agents in parallel. Each writes report to `{sessionsPath}{sessionId}` and returns path.

**Explore #1 - Docs & Conventions:**
```
Analyze repository for: steering files, CONVENTIONS.md, ARCHITECTURE.md,
CLAUDE.md, README conventions, eslint/prettier/tsconfig.

OUTPUT: Write to `{sessionsPath}{sessionId}/explore-docs-conventions.md`, return path.
```

**Explore #2 - Repo Context:**
```
Analyze: project purpose, tech stack, directory structure, main modules, patterns.

OUTPUT: Write to `{sessionsPath}{sessionId}/explore-repo-context.md`, return path.
```

**Capture paths:** `docsConventionsReportPath`, `repoContextReportPath`

### Step 3: Pack Question Preparation (Parallel)

**Filter packs:** If `completedPacks[]` exists (from session continuation), exclude those pack IDs from processing.

**Before launching question-prep agents:**
- Ensure `{sessionsPath}{sessionId}/` exists
- For each selected pack, ensure `{sessionsPath}{sessionId}/{packId}.md` exists with the required frontmatter and `status: pending`
- Populate `pack_id`, `pack_name`, `pack_type`, `interview_mode`, `created_at`, `updated_at`, and `source_reports`
- Reserve a prep artifact path at `{sessionsPath}{sessionId}/prep-{packId}.md`

**Show progress when continuing:**
```
Continuing session: {sessionId}
✅ Completed: {completedPacks.join(', ')}
⏳ Remaining: {remainingPacks.join(', ')}
```

**When running question preparation in parallel:**
- Launch all question-prep agents as background tasks
- Capture all task IDs
- Wait for all to complete (Claude: Task output/wait; Codex: `wait`)
- Display progress as prep artifacts finish
- Do not ask the user from subagents; all user interaction stays in the parent workflow

Spawn a **Task agent per pack** for question preparation only - run all in parallel for faster execution.
Claude Code: use `run_in_background: true` for each `Task` call, then wait for all to complete.

Codex CLI: `spawn_agent` for each pack, capture agent IDs, then `wait` for completion (optionally streaming progress as each finishes).

```yaml
subagent_type: "general-purpose"
description: "Prepare interview questions for {packName}"
prompt: |
  Prepare a grounded interview question set for a single pack.

  ## Pack Info
  - Pack ID: {packId}
  - Pack Name: {packName}
  - Pack Type: {packType}  # "predefined" or "custom"
  - Custom Description: {customTopicDescription}  # only if custom

  ## Context Files
  - Pack Reference: pack-reference.md
  - Repo Context: {repoContextReportPath}
  - Docs & Conventions: {docsConventionsReportPath}

  ## Output
  - Path: {sessionsPath}{sessionId}/prep-{packId}.md

  ## Instructions

  ### 1. Read Context
  Read pack-reference.md to get question themes for this pack.

  Read repoContextReportPath and docsConventionsReportPath for grounding.
  These reports contain ALL necessary findings - do NOT run additional explores.

  ### 2. Generate Questions
  Question count:
  - Predefined: 5
  - Custom (narrow): 3-4
  - Custom (medium): 5
  - Custom (broad): 6-7

  Guidelines:
  - Ground ONLY in the provided explore reports + existing docs
  - Reference actual code: "I see X in Y file..." (from reports)
  - Ask about conventions, not roadmap
  - Offer 4 options (A/B/C/D)
  - Mark one as "⭐ Recommended" at end of description

  Pattern:
  Q: I found {finding}. How should {convention question}?
  A) {Option} — {rationale}
  B) {Option} — {rationale} ⭐ Recommended
  C) {Option} — {rationale}
  D) {Option} — {rationale}

  ### 3. Output Requirements
  Write ONLY preparation material to {sessionsPath}{sessionId}/prep-{packId}.md:
  - Pack summary grounded in the reports
  - Final question list
  - 4 answer options per question
  - One recommended option per question
  - Brief rationale for each question
  - Suggested classification hints for each answer:
    - CONVENTION: Timeless, future-focused ("When implementing X, do Y")
    - ACTION_ITEM: Temporal, fixes current state ("Replace X", "Fix X")

  ### 4. Hard Constraints
  - Do NOT ask the user any questions
  - Do NOT use ASK_USER(...)
  - Do NOT classify final answers, because no user answers exist yet
  - Do NOT write the final pack result file
```

**Session structure:**
```
{sessionsPath}
└── {sessionId}/
    ├── prep-codebase-topology-ownership.md
    ├── prep-architecture-design-invariants.md
    ├── codebase-topology-ownership.md
    ├── architecture-design-invariants.md
    ├── {custom-pack-id}.md
    └── ...
```

**Question-prep file format:**
```markdown
# Prep: {Pack Name}
**Pack ID:** {id}

## Context Summary
- {Grounded finding from reports}

## Questions
### Q1
Question: I found {finding}. How should {convention question}?
A) {Option} — {rationale}
B) {Option} — {rationale} ⭐ Recommended
C) {Option} — {rationale}
D) {Option} — {rationale}

Classification hints:
- A → {CONVENTION|ACTION_ITEM}: {why}
- B → {CONVENTION|ACTION_ITEM}: {why}
- C → {CONVENTION|ACTION_ITEM}: {why}
- D → {CONVENTION|ACTION_ITEM}: {why}
```

**Pack file format:**
```markdown
---
pack_id: {id}
pack_name: {Pack Name}
pack_type: {predefined|custom}
status: {pending|in_progress|complete|blocked|skipped|error}
interview_mode: {interactive|fast}
question_count: {N}
answered_count: {N}
created_at: {ISO-8601 timestamp}
updated_at: {ISO-8601 timestamp}
completed_at: {ISO-8601 timestamp or empty}
source_reports:
  - {docsConventionsReportPath}
  - {repoContextReportPath}
---

# {Pack Name}
**Pack ID:** {id}

## Conventions
- Q1: {extracted rule}
- Q2: {extracted rule}

## Action Items
- Q3: {action item with severity}

## Raw Interview (optional, for reference)
Preserve Q&A in collapsed detail if needed for debugging.
```

### Step 4: Await Prepared Question Sets

Wait for all question-prep agents to complete. Each writes its results to `{sessionsPath}{sessionId}/prep-{packId}.md`.

**Note:** Question-prep agents use existing explore reports - no additional explores are run.

Treat a prep artifact as ready only if all of the following are true:
- File is parseable
- `## Questions` section exists
- Each question includes four options
- At least one recommended option is present per question

If any of these checks fail, treat the prep artifact as `error` and surface it for retry/manual repair.

### Step 5: Sequential Interview Loop

The parent workflow conducts all user interviews. Do not delegate live interviewing to subagents.

For each remaining pack, in order:

1. Read `{sessionsPath}{sessionId}/prep-{packId}.md`
2. If `{sessionsPath}{sessionId}/{packId}.md` already exists with `status: in_progress`:
   - read existing `answered_count`
   - read `## Raw Interview` if present
   - continue from the next unanswered prepared question instead of restarting from question 1
   - if partial answers are missing or inconsistent with `answered_count`, reset `answered_count` to the recoverable number of answers and continue from there
3. Update `{sessionsPath}{sessionId}/{packId}.md` frontmatter:
   - set `status: in_progress`
   - set `updated_at`
   - set `question_count`
   - preserve `answered_count` if resuming
4. Conduct the interview in the parent workflow:
   - If `interviewMode` is `fast`, present the full prepared question set in one message
   - If `interviewMode` is `interactive`, ask questions sequentially or in small batches
5. As answers arrive:
   - update `answered_count`
   - update `updated_at`
6. Classify each answer using the prep artifact's classification hints:
   - CONVENTION: Timeless, future-focused
   - ACTION_ITEM: Temporal, fixes current state
7. Write or update `{sessionsPath}{sessionId}/{packId}.md`:
   - CONVENTION items → `## Conventions`
   - ACTION_ITEM items → `## Action Items`
   - Preserve raw Q&A in `## Raw Interview` when available so interrupted interviews can resume cleanly
8. On successful completion:
   - set `status: complete`
   - set `completed_at`
   - update `updated_at`
9. If required user input is missing or the interview is interrupted:
   - set `status: blocked`
   - update `updated_at`

Treat a pack as complete only if all of the following are true:
- Frontmatter contains `status: complete`
- File is parseable
- `## Conventions` section exists
- `## Action Items` section exists

If any of these checks fail, treat the pack as `error` and surface it for retry/manual repair.

### Step 6: Generate Outputs

Delegate to general-purpose subagent (use the strongest available model in your runtime):

```yaml
subagent_type: "general-purpose"
description: "Generate steerings and action items"
prompt: |
  Generate steerings AND action items from interview sessions.

  Session directory: {sessionsPath}{sessionId}/
  (contains one .md file per pack with classifications)

  Docs report: {docsConventionsReportPath}
  Repo report: {repoContextReportPath}
  Template: steering-template.md

  Output paths: {steeringsPath}, {backlogPath}

  Instructions:
  1. Read all pack files from session directory
  2. Extract CONVENTION items → generate steering files
  3. Extract ACTION_ITEM items → generate backlog file
  4. Generate index.md for steerings
```

**Steering filenames:**

| Pack ID | Filename |
|---------|----------|
| codebase-topology-ownership | code-ownership.md |
| architecture-design-invariants | architecture-invariants.md |
| business-domain-contracts | domain-invariants.md |
| quality-style-assurance | quality-and-style.md |
| testing-verification-strategy | testing-strategy.md |
| risk-historical-landmines | risk-registry.md |
| security-data-compliance | security-and-compliance.md |
| delivery-lifecycle-change-flow | delivery-lifecycle.md |
| {custom-pack-id} | {custom-pack-id}.md |

**Rule style:**
- ✅ "When implementing X, do Y" / "Use X for Y" / "New features require X"
- ❌ "Proactively refactor X" / "Continue using X" / "Add X to existing code"

**Action items file:** `{backlogPath}steering-specs-action-items.md`

Severity: 🔴 CRITICAL (data loss, security) → 🟡 HIGH → 🟢 MEDIUM → 🔵 LOW → ⏸️ DEFERRED

### Step 7: Present Results

```
Generated steerings:
- {steeringsPath}*.md (N rules, M practices each)
- {steeringsPath}index.md

Action Items: {backlogPath}steering-specs-action-items.md
- 🔴 Critical: N | 🟡 High: N | 🟢 Medium: N

Session: {sessionsPath}{sessionId}/
- prep-{packId-1}.md
- prep-{packId-2}.md
- {packId-1}.md
- {packId-2}.md
- ...
```

---

## Review Mode (Step R)

Transform existing steerings to standard format.

### R1: Locate Steerings

Auto-detect or ask: `steerings/`, `docs/steerings/`, `.steerings/`

### R2: Analyze with Explore

```yaml
subagent_type: "Explore"
prompt: |
  Analyze steering files in {steeringsPath}:
  - Structure: Intent → Rules → Practices → Meta?
  - Rules: numbered, prescriptive, no metadata?
  - Code examples: 5-15 lines?
  - File size: <200 lines?

  Report compliant vs non-compliant files with specific issues.
```

### R3: Present Summary

```
✅ Compliant: file1.md, file2.md
⚠️ Need transformation:
- file3.md: Missing Intent, verbose rules
- file4.md: Code examples too long
```

### R4: Ask Action

Options: Transform all | Transform specific | Backup and transform | Plan only | Skip

### R5: Transform

Delegate to general-purpose subagent with transformation guidelines from steering-template.md.

### R6: Present Results

Show changes per file, regenerate index.md.

---

## Quick Reference

**Activation keywords:** steerings, tacit knowledge, interview, conventions, "steerings for X", "continue steerings"

**Session continuation:** Auto-detects existing sessions in `.sessions/`, `sessions/`, `docs/sessions/`. Offers to resume incomplete sessions, skipping completed packs.

**8 Predefined Packs:** Topology, Architecture, Domain, Quality, Testing, Risk, Security, Delivery

**Output format:** Intent (1 sentence) → Rules (numbered) → Practices (heading + explanation) → Meta

**Files generated:**
- `{steeringsPath}*.md` — Steering files
- `{steeringsPath}index.md` — Table of contents
- `{backlogPath}steering-specs-action-items.md` — Action items
- `{sessionsPath}{sessionId}/prep-{packId}.md` — Prepared question sets per pack
- `{sessionsPath}{sessionId}/{packId}.md` — Interview responses per pack
- `{sessionsPath}{sessionId}/explore-*.md` — Discovery reports

**Reference files:**
- [flow-diagram.md](flow-diagram.md) — Visual flow
- [pack-reference.md](pack-reference.md) — Pack definitions and Explore prompts
- [steering-template.md](steering-template.md) — Output format validation
