# ComplyBeacon

Open-source observability toolkit that collects, normalizes, and enriches compliance evidence by extending the OpenTelemetry standard. Uses a Go workspace monorepo with two active modules (`proofwatch`, `truthbeam`) and an OTel Collector distribution (`beacon-distro`).

## Structure

```text
proofwatch/              # Go module — evidence collection & emission library
  internal/metrics/      # OTel metrics observer (evidence counters)
  cmd/validate-logs/     # CLI tool for validating log output
truthbeam/               # Go module — OTel Collector enrichment processor
  internal/applier/      # Attribute application logic
  internal/client/       # Generated OpenAPI client + otter cache
  internal/metadata/     # Component metadata + test fixtures
beacon-distro/           # OTel Collector distribution (manifest.yaml + Containerfile)
model/                   # Weaver semantic convention definitions (source of truth for attributes)
templates/               # Weaver Jinja2 code generation templates
configs/                 # Collector and Loki deployment configs (base, storage, enrichment)
certs/                   # Generated TLS certificates (gitignored, created by task infra:generate-self-signed-cert)
deploy/                  # Deployment infrastructure (Terraform)
tests/                   # Test infrastructure
  integration/           # E2E Ginkgo tests, mock Compass, evidence fixtures
docs/                    # Architecture (DESIGN.md), dev guide (DEVELOPMENT.md), attribute docs
openspec/                # OpenSpec change proposals and specs
.specify/memory/         # ComplyTime constitution (org-wide standards)
Taskfile.yml             # Build automation entry point (canonical)
.taskfiles/              # Task modules (dev, codegen, infra, quality, version, integration) + scripts
```

## Commands

```bash
task                     # List all available targets
task build               # Build the collector container image
task test                # Unit tests with coverage (proofwatch + truthbeam)
task test-race           # Tests with race detection
task lint                # Lint all modules (golangci-lint v2)
task check               # Run all quality gates (lint + test)
task deps                # Tidy, verify, download deps for all modules
task clean               # Remove build artifacts and test output
task version:sync        # Sync Go + OTel versions across all modules, Containerfiles, CI, docs
task version:check       # Validate version alignment (read-only, for CI)
task infra:deploy        # Start local stack (podman-compose)
task infra:undeploy      # Stop local stack — DESTRUCTIVE
```

## Constraints

- **Go workspace, no root go.mod**: This repo uses `go.work` to link modules. All module-level commands iterate over `MODULES := ./proofwatch ./truthbeam`. Running `go test ./...` from root will not work — use `task test`.
- **Generated files — DO NOT EDIT**:
  - `proofwatch/attributes.go` — regenerate with `task codegen:weaver-codegen`
  - `truthbeam/internal/applier/attributes.go` — regenerate with `task codegen:weaver-codegen`
  - `truthbeam/internal/client/client.gen.go` — regenerate with `task codegen:api-codegen`
  - `docs/attributes/*.md` — regenerate with `task codegen:weaver-docsgen`
- **Build automation**: Use `task` (taskfile.dev), not `make`. A deprecated Makefile exists but is not maintained.
- **External tools**: Install development tools with `task tools:install-all` or `task tools:install-weaver`. SHA256 checksums are pinned in `.tool_checksums` for supply chain security. Ginkgo CLI is managed as a `tool` directive in the root `go.mod` and invoked via `go tool ginkgo`.
- **Podman, not Docker**: Container operations use `podman` and `podman-compose`. Do not reference `docker` commands.
- **Lint**: Go linting uses `.golangci.yml` (v2 format). Multi-language CI linting uses `.mega-linter.yml`. No pre-commit hooks — run `task lint` locally.
- **Integration tests**: `tests/integration/` contains Ginkgo E2E tests and a mock Compass HTTP server. Run with `task integration:test` (all layers) or `task integration:test-profile PROFILE=base|storage|enrichment`. Do not recreate mock Compass — it already serves fixture-driven `/v1/enrich` responses.
- **Standards**: All coding standards are in `.specify/memory/constitution.md`. For architecture context, see `docs/DESIGN.md`. For dev setup, see `docs/DEVELOPMENT.md`.

## Local Dev Stack

The compose stack (`compose.yaml`) runs the full evidence pipeline locally:

- **Base** (always on): **Loki** (log aggregation), **collector** (OTel Collector with evidence processing)
- **Storage** (`--profile storage`): adds **rustfs** (S3-compatible object storage, API <http://localhost:9000>, console <http://localhost:9001>, credentials `rustfsadmin`/`rustfsadmin`)
- **Enrichment** (`--profile enrichment`): adds **Compass** (mock enrichment service for TruthBeam)
- **Debug** (`--profile debug`): adds **Grafana** (visualization, <http://localhost:3000>)

Collector config is selected via `COLLECTOR_CONFIG` env var (defaults to `configs/collector-base.yaml`).

All S3 environment variables have inline defaults that target rustfs. The stack works out-of-the-box with `task infra:deploy` — no AWS credentials needed. Override env vars in your shell to target real AWS S3 instead.

## Commits

All commits MUST use Conventional Commits, the `-s` flag (Signed-off-by), and include an `Assisted-by` trailer when AI-assisted. See `.specify/memory/constitution.md` for full details.

## Integration Test Gotchas

- **RustFS is not MinIO**: The health endpoint is `/health`, not `/minio/health/live` (returns 403). The S3 API is compatible but administrative endpoints differ. The `rustfs-init` container uses `rc` (RustFS CLI), not `mc` (MinIO client).
- **Loki 3.x OTLP labels**: Loki no longer auto-creates `{exporter="OTLP"}` from OTLP ingestion. Query indexed labels from `otlp_config.resource_attributes` in `configs/loki.yaml` directly (e.g. `{policy_rule_id="..."}`). Log attributes are stored as structured metadata and can be filtered with `| key="value"` after the stream selector.
- **Podman rootless volume permissions**: The collector runs as UID 10001. Host-mounted volumes need the `:U` flag (podman ownership remapping) or the container can't write. The `:Z` flag alone only handles SELinux relabeling.
- **S3 test portability**: Don't depend on host-installed CLI tools (`aws`, `mc`). The test bucket is configured with anonymous public access via `rc anonymous set public` in `rustfs-init`, so tests can query the S3 ListObjectsV2 API with plain HTTP — no auth headers needed.
- **awss3 exporter partitioning**: When `resource_attrs_to_s3.s3_prefix` is set, it replaces (not appends to) the static `s3_prefix`. Objects land at `{resource_attr_value}/evidence_logs_{uuid}.json`, not `{s3_prefix}/{resource_attr_value}/...`.

## Active Technologies

- Go 1.26.3, multi-module workspace (`go.work`)
- OpenTelemetry Collector SDK v1.58.0 / v0.152.0 (stable + experimental series, component framework, pipeline data, processor interfaces)
- `github.com/gemaraproj/go-gemara` v0.5.0 (compliance evidence model — Gemara v1 schema)
- `github.com/telophasehq/go-ocsf` v0.2.1 (OCSF cybersecurity schema types)
- `github.com/maypok86/otter/v2` (in-memory cache, truthbeam)
- `github.com/oapi-codegen` (OpenAPI client generation, truthbeam)
- `go.uber.org/zap` (structured logging, truthbeam)
- `github.com/stretchr/testify` (test assertions, both modules)
- OTel Weaver (semantic convention model — Go constants + docs)
- OTel Collector Builder v0.152.0 (beacon-distro binary)
- golangci-lint v2, MegaLinter, SonarCloud, gaze (quality tooling)
- Task v3 ([taskfile.dev](https://taskfile.dev)) — build automation
- Podman + podman-compose (container runtime)
- Container: UBI10 Minimal (certs), golang:1.26.3 (build), UBI10 Micro (runtime)
- Registries: `ghcr.io` (primary), Quay.io (secondary)
