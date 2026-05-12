# ComplyBeacon

## Project Overview

ComplyBeacon is an open-source observability toolkit that collects,
normalizes, and enriches compliance evidence by extending the
OpenTelemetry (OTEL) standard. It bridges raw policy scanner output
and modern logging pipelines into a unified, auditable data stream.

- **Type**: Go monorepo (multi-module workspace)
- **License**: Apache-2.0
- **Go version**: 1.25.8 (toolchain 1.25.9)
- **Modules**: `proofwatch` (instrumentation library),
  `truthbeam` (OTel Collector processor)
- **Mission**: Provide enriched, standards-based compliance
  evidence as an OpenTelemetry log stream

## Build & Test Commands

All commands run from the repository root via `make`.

### Build & Deploy

```sh
make workspace          # Setup Go workspace with all modules
make deploy             # Deploy full stack via podman-compose
make undeploy           # Tear down container stack
make sync-otel-versions # Sync beacon-distro manifest from truthbeam
```

### Test

```sh
make test               # Unit tests with coverage for all modules
                        # (includes go-version, go-mod-consistency,
                        #  and otel-version drift checks)
make test-race          # Tests with -race detection
make coverage-report    # Generate HTML coverage reports
make crapload           # CRAP + GazeCRAP analysis (human-readable)
make crapload-baseline  # Generate .gaze/baseline.json per module
make crapload-check     # Check for CRAP regressions vs baseline
```

### Lint

```sh
make golangci-lint      # golangci-lint for all modules
                        # (config: .golangci.yml)
```

### Code Generation & Semantic Conventions

```sh
make api-codegen              # go generate for all modules
make weaver-codegen           # Generate Go code from model/
make weaver-docsgen           # Generate docs from model/
make weaver-check             # Validate model schema
make weaver-semantic-check    # Validate logs against semconv
```

### Version Drift Checks

```sh
make check-go-version           # Containerfile vs module Go version
make check-otel-versions        # manifest.yaml vs truthbeam OTel
make check-go-mod-consistency   # OTel dep consistency within go.mod
make help                       # List all available targets
```

### CI Workflow Structure

| Workflow | File | Purpose |
|----------|------|---------|
| CI | `ci_checks.yml` | Reusable standardized CI (org-infra) |
| Local CI | `ci_local.yml` | golangci-lint, test, weaver-check, verify-codegen |
| CRAP Load | `ci_crapload.yml` | CRAP analysis on PRs |
| Security | `ci_security.yml` | OSV-Scanner, Trivy, OpenSSF Scorecards |
| SonarCloud | `ci_sonarcloud.yml` | SonarQube coverage analysis |
| Version Sync | `ci_version_sync.yml` | Beacon-distro/truthbeam version alignment |
| Go Compat | `ci_go_compat.yml` | Weekly forward-compat build (pinned + latest) |
| Dependencies | `ci_dependencies.yml` | Dependency review + Dependabot auto-approval |
| Publish GHCR | `ci_publish_ghcr.yml` | Build, scan, sign, publish container images |
| Promote Quay | `ci_publish_quay.yml` | Tag-triggered promotion to Quay.io |
| Scheduled | `ci_scheduled.yml` | Daily OSV-Scanner + Scorecards |

## Project Structure

```text
proofwatch/              Go module -- evidence collection & emission library
  cmd/validate-logs/     CLI tool for validating log output
  internal/metrics/      OTel metrics observer (evidence counters)
truthbeam/               Go module -- OTel Collector enrichment processor
  internal/applier/      Attribute application logic
  internal/client/       Generated OpenAPI client + otter cache
beacon-distro/           Custom OTel Collector distribution
  Containerfile.collector
  manifest.yaml          OCB manifest (version-synced with truthbeam)
  config.yaml            Collector configuration
compass/                 Stub -- real service at gemara-content-service repo
model/                   Weaver semantic convention definitions (source of truth)
  attributes.yaml        Attribute registry
  entities.yaml          Entity definitions
templates/               Weaver code generation templates
  registry/              Go code templates
docs/                    Project documentation
  attributes/            Generated attribute docs
  integration/           Integration guides
hack/                    Development tooling
  demo/                  Demo infrastructure (compose, Terraform)
  sampledata/            Sample compliance evidence payloads
  self-signed-cert/      TLS certificate generation
scripts/                 Version drift and sync scripts
openspec/                OpenSpec specification workflow
  changes/               Active and archived change specs
  schemas/               Spec templates (unbound-force)
  specs/                 Standalone specs
.specify/memory/         ComplyTime constitution (org-wide standards)
```

## Coding Conventions

### Go

- **Formatting**: `goimports` and `go fmt` (enforced by golangci-lint)
- **Linter**: golangci-lint v2 with `standard` defaults + `gosec`
  (config: `.golangci.yml`)
- **Multi-language CI lint**: MegaLinter (config: `.mega-linter.yml`)
- **File headers**: `// SPDX-License-Identifier: Apache-2.0`
- **File naming**: lowercase with underscores (`my_file.go`)
- **Package names**: short, lowercase, no underscores
- **Error handling**: always check and handle; return to caller
  when unresolvable locally
- **Import grouping**: stdlib, external, internal (enforced by
  `goimports`)
- **Line length**: 99 characters unless exceeding improves
  readability
- **Magic values**: centralize in dedicated constants; no inline
  magic strings or numbers

### Spec Writing

- Use RFC 2119 language: MUST, SHOULD, MAY
- Scenarios: Given/When/Then format
- Requirement numbering: FR-NNN
- Line length: < 72 characters

## Testing Conventions

- **Framework**: Go stdlib `testing` package
- **Assertions**: `stretchr/testify` (`assert` and `require`)
- **Naming**: `TestFunctionName` or
  `TestFunctionName_Description` (e.g., `TestProcessLogs`,
  `TestNewTruthBeamProcessorWithInvalidConfig`)
- **Test fixtures**: struct-based fixtures with setup helpers
  (e.g., `setupProofWatchTest`)
- **Isolation**: `httptest.NewServer` for HTTP mocking,
  in-memory exporters for OTel, `t.TempDir()` for filesystem
- **Coverage**: `-coverprofile=coverage.out -covermode=atomic`
  per module via `make test`
- **Quality gates**: CRAP load analysis via `gaze` with
  baseline regression checks
- **Negative tests**: each scenario MUST include positive and
  negative cases (error paths, invalid configs)
- **Module isolation**: tests run with `GOWORK=off` to verify
  each module compiles independently

## Constraints

- **Go workspace, no root go.mod**: This repo uses `go.work`
  to link modules. All module-level commands iterate over
  `MODULES := ./proofwatch ./truthbeam`. Running
  `go test ./...` from root will not work -- use `make test`.
- **Generated files -- DO NOT EDIT**:
  - `proofwatch/attributes.go` -- regenerate with
    `make weaver-codegen`
  - `truthbeam/internal/applier/attributes.go` -- regenerate
    with `make weaver-codegen`
  - `truthbeam/internal/client/client.gen.go` -- regenerate
    with `make api-codegen`
  - `docs/attributes/*.md` -- regenerate with
    `make weaver-docsgen`
- **compass/ is external**: Do not build or modify. The actual
  service is maintained at
  `github.com/complytime/gemara-content-service`. This
  directory exists only for compose integration.
- **Podman, not Docker**: Container operations use `podman`
  and `podman-compose`. Do not reference `docker` commands.
- **No pre-commit hooks**: Run `make golangci-lint` locally
  before submitting PRs.

## Commits

All commits MUST use Conventional Commits, the `-s` flag
(Signed-off-by), and include an `Assisted-by` trailer when
AI-assisted. See `.specify/memory/constitution.md` for full
details.

## Active Technologies

- Go 1.25.8 (toolchain 1.25.9), multi-module workspace
  (`go.work`)
- OpenTelemetry Collector SDK v1.57.0 / v0.151.0 (component
  framework, pipeline data, processor interfaces)
- `github.com/gemaraproj/go-gemara` v0.4.0 (compliance
  evidence model -- Gemara schema)
- `github.com/Santiago-Labs/go-ocsf` (OCSF cybersecurity
  schema types)
- `github.com/maypok86/otter/v2` (in-memory cache, truthbeam)
- `github.com/oapi-codegen` (OpenAPI client generation,
  truthbeam)
- `go.uber.org/zap` (structured logging, truthbeam)
- `github.com/stretchr/testify` (test assertions, both modules)
- OTel Weaver (semantic convention model -- Go constants + docs)
- golangci-lint v2, MegaLinter, SonarCloud, gaze (quality
  tooling)
- Podman + podman-compose (container runtime)
- Registries: `ghcr.io` (primary), Quay.io (secondary)

## Behavioral Rules

These rules are non-negotiable. Violations are CRITICAL severity.

- **Gatekeeping**: MUST NOT modify quality/governance gates
  (coverage thresholds, CRAP scores, severity definitions,
  CI flags, agent settings, constitution MUST rules, review
  limits, workflow markers). Stop and report instead.
- **Phase boundaries**: MUST NOT cross workflow phase boundaries.
  Spec phases: spec artifacts only. Implement: source code.
  Review: fixes only. Violation = process error, stop immediately.
- **CI parity**: MUST replicate CI checks locally before marking
  tasks complete. Derive commands from `.github/workflows/`.
- **Review council**: MUST run `/review-council` before PR
  submission. Resolve all REQUEST CHANGES. No code changes
  between APPROVE and PR. Exempt: constitution amendments,
  docs-only, emergency hotfixes.
- **Branch protection**: MUST NOT commit directly to `main`.
  All changes via feature branches and PRs.
- **Documentation gate**: Before marking a task complete,
  assess documentation impact: `CHANGELOG.md` for change
  entries, `AGENTS.md` for structural updates (project
  structure, conventions, build commands), `README.md` for
  description changes.
- **Website gate**: MUST file `unbound-force/website` issue
  for user-facing changes before PR merge. Exempt: internal
  refactoring, test-only, CI-only, spec artifacts.
- **Zero-waste**: No orphaned specs, unused standards, or
  aspirational documents that do not map to actionable work.

### PR Review Commands

| Command | When | Scope |
|---------|------|-------|
| `/review-council` | Pre-PR (local) | 5+ Divisor agents |
| `/review-pr [N]` | Post-PR (GitHub) | Single agent, CI analysis |

## Specification Workflow

All non-trivial changes MUST be preceded by a spec workflow.

| Tier | Tool | When | Artifacts |
|------|------|------|-----------|
| Strategic | Speckit | >= 3 stories, cross-repo | `specs/NNN-*/` |
| Tactical | OpenSpec | < 3 stories, single-repo | `openspec/changes/*/` |

Pipeline: `constitution -> specify -> clarify -> plan -> tasks ->
analyze -> checklist -> implement`

**Ordering**: Constitution before specs. Spec before plan. Plan
before tasks. Tasks before implementation. Spec artifacts MUST
be committed/pushed before implementation begins.

**Branches**: Speckit: `NNN-<name>`. OpenSpec: `opsx/<name>`.

**Task bookkeeping**: Mark checkboxes `[x]` immediately on
completion. `[P]` marks parallel-eligible tasks.

**When in doubt**: Start with OpenSpec. Escalate to Speckit if
scope grows beyond 3 stories or crosses repo boundaries.

**What requires a spec**: New features, refactoring that changes
signatures, test additions across multiple functions, agent
changes, CI changes, data model changes.

**Exempt**: Constitution amendments, typo fixes, emergency
hotfixes (retroactively documented).

## Knowledge Retrieval

Prefer Dewey MCP tools over grep/glob/read for cross-repo
context and architectural patterns.

| Intent | Tool |
|--------|------|
| Conceptual | `dewey_semantic_search` |
| Keyword | `dewey_search` |
| Navigation | `dewey_traverse`, `dewey_get_page` |
| Discovery | `dewey_find_connections`, `dewey_similar` |

**Fallback**: Use Read/Grep/Glob when Dewey is unavailable,
for exact string matching, known file paths, or non-Markdown
content (Go source, JSON, YAML).

## Convention Packs

This repository uses convention packs scaffolded by
unbound-force. Agents MUST read the applicable pack(s)
before writing or reviewing code.

- `.opencode/uf/packs/default.md`
- `.opencode/uf/packs/default-custom.md`
- `.opencode/uf/packs/severity.md`
- `.opencode/uf/packs/content.md`
- `.opencode/uf/packs/content-custom.md`

## Architecture

ComplyBeacon follows a pipeline architecture built on the
OpenTelemetry Collector framework:

- **Multi-module Go workspace**: `proofwatch` and `truthbeam`
  are independent Go modules linked via `go.work`, each with
  its own `go.mod` and test coverage
- **OTel Collector extension**: TruthBeam is a custom processor
  plugin implementing the `processor.Logs` interface
- **Semantic conventions model**: Domain attributes defined in
  `model/` using OTel Weaver, with code generation into both
  modules via `make weaver-codegen`
- **Options pattern**: Configuration uses functional options
  (e.g., `WithTracerProvider`, `WithMeterProvider`)
- **Version-locked distribution**: `beacon-distro/manifest.yaml`
  is version-synced from `truthbeam/go.mod` via
  `make sync-otel-versions` with CI drift detection
- **External enrichment**: Compass service consumed as a
  pre-built container image, accessed via HTTP client with
  caching (`otter`) and TLS support
