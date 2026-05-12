---
name: speckit-workflow
description: "Teaches the Swarm coordinator to follow Speckit's pre-decomposed tasks.md as the authoritative task breakdown instead of generating its own decomposition via CASS."
tags:
  - speckit
  - workflow
  - decomposition
---
<!-- scaffolded by uf vdev -->

# Speckit Workflow — Swarm Skill

This skill teaches you (the Swarm coordinator) how to work
with Speckit's pre-decomposed task lists. When a `tasks.md`
file exists for the active feature, it is the authoritative
decomposition — do NOT re-decompose via CASS.

## When This Skill Applies

Check for a `tasks.md` file in the active spec directory:

1. Look for `specs/NNN-*/tasks.md` relative to the project
   root (Speckit strategic workflow)
2. Look for `openspec/changes/*/tasks.md` (OpenSpec
   tactical workflow)

If `tasks.md` exists, follow the instructions below. If no
`tasks.md` exists, proceed with standard CASS decomposition.

## Reading tasks.md

### Phase Structure

Tasks are organized into numbered phases:

```
## Phase 1: Setup
## Phase 2: Foundational
## Phase 3: User Story 1 (P1)
## Phase 4: User Story 2 (P1)
...
## Phase N: Polish
```

**Map each phase to a Swarm epic.** Phase 1 becomes
epic "Setup", Phase 2 becomes epic "Foundational", etc.

### Task Format

Each task follows this format:

```
- [ ] T001 [P] [US1] Description with file path
```

- **Checkbox** `- [ ]`: Uncompleted task. `- [x]`: Done.
- **Task ID** `T001`: Sequential identifier
- **`[P]` marker**: This task can run in parallel with
  other `[P]` tasks in the same phase (they touch different
  files). Tasks WITHOUT `[P]` are sequential.
- **`[US1]` label**: Maps to User Story 1 from the spec.
  Use as cell metadata when creating Swarm cells.
- **Description**: What to do and which file to modify.

### Phase Dependencies

- **Setup** and **Foundational** phases MUST complete
  before any User Story phases begin.
- User Story phases may run in parallel if they are marked
  as independent in the Dependencies section of tasks.md.
- **Polish** phase runs after all User Story phases.

## Creating Swarm Work Items

1. **Read the full `tasks.md`** to understand all phases.

2. **Create one epic per phase** using `hive_create_epic()`:
   - Epic title = phase name (e.g., "Phase 3: US1 - Doctor Core")
   - Subtasks = the `- [ ]` tasks within that phase

3. **For tasks marked `[P]`**: These can be spawned as
   parallel workers. Each `[P]` task touches different
   files, so file reservation conflicts are unlikely.

4. **For tasks without `[P]`**: These MUST be executed
   sequentially within the phase. Spawn one worker at a
   time, waiting for completion before the next.

5. **For tasks already checked `[x]`**: Skip these — they
   are already completed.

## Worker Instructions

When spawning workers via `swarm_spawn_subtask()`:

1. Include the task description and file path in the
   worker's prompt
2. Workers MUST call `swarmmail_reserve()` on the target
   file(s) before editing
3. Workers MUST mark the task `[x]` in `tasks.md` after
   completing it
4. Workers MUST call `swarm_complete()` when done

## Phase Checkpoints

After all tasks in a phase are complete:

1. Run the project's test suite (check `AGENTS.md` for
   the build/test commands)
2. If tests fail, create a new cell to fix the failure
   before advancing to the next phase
3. Report phase completion via `swarm_status()`

## Branch Safety

**CRITICAL**: All work MUST be committed and pushed on the
current feature branch before any branch switch occurs.

- After completing all phases, commit and push all changes
  before suggesting PR creation, merging, or switching to
  `main`.
- Before creating a new feature branch (via `/speckit.specify`),
  check `git status --short` for uncommitted changes. If
  uncommitted changes exist, **STOP** and ask the user for
  confirmation before switching branches.
- Never silently switch branches with a dirty working tree.
  Uncommitted changes may follow to the wrong branch or be
  lost entirely.

## Prerequisite Skill

Load `unbound-force-heroes` alongside this skill to get
hero routing and workflow stage context:

```
skills_use({ name: "unbound-force-heroes" })
skills_use({ name: "speckit-workflow" })
```

## Entry Point

The `/unleash` command is the primary way to trigger
autonomous pipeline execution using this skill. It
orchestrates the full Speckit pipeline (clarify, plan,
tasks, spec review, implement, code review,
retrospective, demo) and uses the task format described
above for the implementation phase. `/unleash` also
supports OpenSpec (`opsx/*`) branches.
