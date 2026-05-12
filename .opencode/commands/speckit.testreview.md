---
description: Perform a read-only testability analysis of spec artifacts by delegating to the reviewer-testing agent in Spec Review Mode.
---
<!-- scaffolded by gaze dev -->

## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Goal

Assess the testability of feature specification artifacts (`spec.md`, `plan.md`, `tasks.md`) through a dedicated testing lens. This command identifies vague acceptance criteria, missing coverage strategy, undefined contract surfaces, and infeasible fixture requirements — issues that cause rework if discovered only during implementation. This command MUST run only after `/speckit.tasks` has successfully produced a complete `tasks.md`.

## Operating Constraints

**STRICTLY READ-ONLY**: Do **not** modify any files. Output a structured testability analysis report. Offer an optional remediation plan (user must explicitly approve before any follow-up editing commands would be invoked manually).

**Constitution Authority**: The project constitution (`.specify/memory/constitution.md`) is **non-negotiable** within this analysis scope. Constitution Principle IV (Testability) violations are automatically CRITICAL. Missing coverage strategy is CRITICAL. These require adjustment of the spec, plan, or tasks — not dilution of the principle.

## Execution Steps

### 1. Initialize Analysis Context

Run `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks` once from repo root and parse JSON for FEATURE_DIR and AVAILABLE_DOCS. Derive absolute paths:

- SPEC = FEATURE_DIR/spec.md
- PLAN = FEATURE_DIR/plan.md
- TASKS = FEATURE_DIR/tasks.md

Abort with an error message if any required file is missing. Instruct the user to run `/speckit.tasks` first if tasks.md is absent, or `/speckit.specify` if spec.md is absent.

For single quotes in args like "I'm Groot", use escape syntax: e.g 'I'\''m Groot' (or double-quote if possible: "I'm Groot").

### 2. Load Artifacts (Progressive Disclosure)

Load only the minimal necessary context from each artifact:

**From spec.md:**
- User Stories with acceptance scenarios
- Functional Requirements (FR-xxx)
- Success Criteria (SC-xxx)
- Edge Cases
- Assumptions

**From plan.md:**
- Testing Strategy section
- Phase structure and test-related tasks
- Technical constraints
- Constitution Check results

**From tasks.md:**
- Test-related tasks (if any)
- Task-to-requirement mapping via [US*] labels
- Phase ordering and dependencies

**From constitution:**
- Load `.specify/memory/constitution.md` — focus on Principle IV: Testability

### 3. Delegate to Reviewer-Testing Agent

Use the Task tool to delegate the analysis to the `reviewer-testing` agent in **Spec Review Mode**:

- Instruct the agent to operate in **Spec Review Mode** (not Code Review Mode)
- Pass the feature directory path so the agent can read all spec artifacts
- Instruct the agent to read the constitution and AGENTS.md for context
- Collect the agent's findings and verdict

### 4. Format Testability Report

Produce a Markdown report (no file writes) with the following structure:

#### Testability Analysis Report

| ID | Category | Severity | Location | Summary | Recommendation |
|----|----------|----------|----------|---------|----------------|

Categories map to the reviewer-testing agent's Spec Review Mode audit checklist:
- **Testability**: Vague or unmeasurable acceptance criteria
- **Strategy**: Missing or incomplete test strategy (unit/integration/e2e)
- **Fixtures**: Infeasible or undocumented test fixture requirements
- **Coverage**: Missing coverage targets or ratchet definitions
- **Contracts**: Undefined or ambiguous contract surfaces
- **Constitution**: Principle IV violations

**Testability Summary:**

| Dimension | Status | Notes |
|-----------|--------|-------|
| Acceptance Criteria Testability | PASS/FAIL | |
| Test Strategy Defined | PASS/FAIL | |
| Fixture Feasibility | PASS/FAIL | |
| Coverage Targets Specified | PASS/FAIL | |
| Contract Surfaces Defined | PASS/FAIL | |
| Constitution IV Compliance | PASS/FAIL | |

**Metrics:**
- Total acceptance criteria assessed
- Testable criteria count
- Vague criteria count
- Missing strategy areas
- CRITICAL findings count

### 5. Provide Next Actions

At end of report, output a concise Next Actions block:

- If CRITICAL issues exist: Recommend resolving before `/speckit.implement`
- If only LOW/MEDIUM: User may proceed, with improvement suggestions
- Provide explicit command suggestions: e.g., "Run `/speckit.clarify` to define coverage targets", "Update plan.md to add test strategy section"

### 6. Offer Remediation

Ask the user: "Would you like me to suggest concrete remediation edits for the top N issues?" (Do NOT apply them automatically.)

## Operating Principles

### Context Efficiency

- **Minimal high-signal tokens**: Focus on testability findings, not exhaustive documentation
- **Progressive disclosure**: Load artifacts incrementally
- **Token-efficient output**: Limit findings table to 30 rows; summarize overflow
- **Deterministic results**: Rerunning without changes should produce consistent findings

### Analysis Guidelines

- **NEVER modify files** (this is read-only analysis)
- **NEVER hallucinate missing sections** (if absent, report them accurately)
- **Prioritize Principle IV violations** (these are always CRITICAL)
- **Missing coverage strategy is CRITICAL** — not HIGH, not MEDIUM
- **Report zero issues gracefully** (emit success report with testability statistics)

## Context

$ARGUMENTS
