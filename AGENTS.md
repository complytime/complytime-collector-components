# ComplyBeacon

Open-source observability toolkit that collects, normalizes, and exports compliance evidence by extending the OpenTelemetry standard. Uses a Go workspace monorepo with one active module (`proofwatch`) and an OTel Collector distribution (`beacon-distro`).

## Structure

```text
proofwatch/              # Go module — evidence collection & emission library
  internal/metrics/      # OTel metrics observer (evidence counters)
  cmd/validate-logs/     # CLI tool for validating log output
beacon-distro/           # OTel Collector distribution (manifest.yaml + Containerfile)
model/                   # Weaver semantic convention definitions (source of truth for attributes)
templates/               # Weaver Jinja2 code generation templates
configs/                 # Collector and Loki deployment configs (base, storage, auth)
certs/                   # Generated TLS certificates (gitignored, created by task infra:generate-self-signed-cert)
deploy/                  # Deployment infrastructure (Terraform)
tests/                   # Test infrastructure
  integration/           # E2E Ginkgo tests, evidence fixtures
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
task test                # Unit tests with coverage (proofwatch)
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

- **Go workspace, no root go.mod**: This repo uses `go.work` to link modules. All module-level commands iterate over `MODULES := ./proofwatch`. Running `go test ./...` from root will not work — use `task test`.
- **Generated files — DO NOT EDIT**:
  - `proofwatch/attributes.go` — regenerate with `task codegen:weaver-codegen`
  - `docs/attributes/*.md` — regenerate with `task codegen:weaver-docsgen`
- **Build automation**: Use `task` (taskfile.dev), not `make`. A deprecated Makefile exists but is not maintained.
- **External tools**: Install development tools with `task tools:install-all` or `task tools:install-weaver`. SHA256 checksums are pinned in `.tool_checksums` for supply chain security. Ginkgo CLI is managed as a `tool` directive in the root `go.mod` and invoked via `go tool ginkgo`.
- **Podman, not Docker**: Container operations use `podman` and `podman-compose`. Do not reference `docker` commands.
- **Lint**: Go linting uses `.golangci.yml` (v2 format). Multi-language CI linting uses `.mega-linter.yml`. No pre-commit hooks — run `task lint` locally.
- **Integration tests**: `tests/integration/` contains Ginkgo E2E tests. Run with `task integration:test` (all layers) or `task integration:test-profile PROFILE=base|storage|storage-tls|auth`.
- **Standards**: All coding standards are in `.specify/memory/constitution.md`. For architecture context, see `docs/DESIGN.md`. For dev setup, see `docs/DEVELOPMENT.md`.

## Local Dev Stack

The compose stack (`compose.yaml`) runs the full evidence pipeline locally:

- **Base** (always on): **Loki** (log aggregation), **collector** (OTel Collector with evidence processing)
- **Storage** (`--profile storage`): adds **rustfs** (S3-compatible object storage, API <http://localhost:9000>, console <http://localhost:9001>, credentials `rustfsadmin`/`rustfsadmin`)
- **Storage-TLS** (`--profile storage-tls`): storage profile with TLS-secured S3 export
- **Auth** (`--profile auth`): adds **Dex** (OIDC provider) and configures the collector with OIDC-secured OTLP receivers
- **Debug** (`--profile debug`): adds **Grafana** (visualization, <http://localhost:3000>)

Collector config is selected via `COLLECTOR_CONFIG` env var (defaults to `configs/collector-base.yaml`).

All S3 environment variables have inline defaults that target rustfs. The stack works out-of-the-box with `task infra:deploy` — no AWS credentials needed. Override env vars in your shell to target real AWS S3 instead.

## Commits

All commits MUST use Conventional Commits, the `-s` flag (Signed-off-by), and include an `Assisted-by` trailer when AI-assisted. See `.specify/memory/constitution.md` for full details.

## Changelog

`CHANGELOG.md` follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. The audience is **users and adopting developers**, not contributors or CI.

**Principles:**

- Write entries that answer "what can I do now?" or "what changed for me?", not "what did we refactor internally?"
- Group by component scope (`proofwatch`, `beacon-distro`) when an entry is module-specific
- Use descriptive, multi-line entries that give enough context for someone unfamiliar with the codebase to understand the value — avoid terse one-liners that require reading source to interpret
- Internal tooling changes (CI actions, linter upgrades, dependency bumps, code quality tooling) are **not** changelog entries unless they affect the user experience or public API
- Do not use `BREAKING` labels until there is a prior release to break against
- `Changed` and `Removed` sections only apply when a released feature is modified or dropped — pre-release iteration is not a "change"
- Reference PR numbers only when they add traceability for significant features

## Integration Test Gotchas

- **Auth profile uses Dex as OIDC provider**: The `auth` profile adds a Dex container (`dexidp/dex:v2.45.1`) and a dedicated `collector-auth` service with `depends_on: dex (service_healthy)` to ensure OIDC discovery succeeds at startup. Dex provides a static test user (`test@complybeacon.dev` / `testpassword`). The collector's OTLP receivers require a valid JWT from Dex.
- **Compose profiles are mutually exclusive for collector variants**: Each collector variant (`collector`, `collector-tls`, `collector-auth`) has its own profile to prevent port conflicts. The default `collector` has `profiles: [base, storage]`. The Taskfile handles profile selection — always use `task integration:test PROFILE=<name>`, not raw `podman-compose` commands.
- **RustFS is not MinIO**: The health endpoint is `/health`, not `/minio/health/live` (returns 403). The S3 API is compatible but administrative endpoints differ. The `rustfs-init` container uses `rc` (RustFS CLI), not `mc` (MinIO client).
- **Loki 3.x OTLP labels**: Loki no longer auto-creates `{exporter="OTLP"}` from OTLP ingestion. Query indexed labels from `otlp_config.resource_attributes` in `configs/loki.yaml` directly (e.g. `{policy_rule_id="..."}`). Log attributes are stored as structured metadata and can be filtered with `| key="value"` after the stream selector.
- **Podman rootless volume permissions**: The collector runs as UID 10001. Host-mounted volumes need the `:U` flag (podman ownership remapping) or the container can't write. The `:Z` flag alone only handles SELinux relabeling.
- **S3 test portability**: Don't depend on host-installed CLI tools (`aws`, `mc`). The test bucket is configured with anonymous public access via `rc anonymous set public` in `rustfs-init`, so tests can query the S3 ListObjectsV2 API with plain HTTP — no auth headers needed.
- **awss3 exporter partitioning**: When `resource_attrs_to_s3.s3_prefix` is set, it replaces (not appends to) the static `s3_prefix`. Objects land at `{resource_attr_value}/evidence_logs_{uuid}.json`, not `{s3_prefix}/{resource_attr_value}/...`.

## CI Workflow Gotchas

- **`artifact-metadata` is not a valid GitHub Actions permission** scope. The correct scope for build provenance attestations (actions/attest) is `attestations: write`. Using an invalid scope causes a startup_failure because GitHub validates all permission blocks including in reusable workflows before any job runs.
- **Reusable workflow permission validation is global**: When `ci_local.yml` calls `ci_publish_ghcr.yml` via `uses:`, GitHub parses and validates the called workflow's permissions block even if the calling job wouldn't run it. An invalid permission in any reusable workflow breaks the entire caller.
- **Required permissions for GHCR publish with attestations**: The `build-beacon-distro` job in `ci_publish_ghcr.yml` needs exactly: `contents: read`, `packages: write`, `id-token: write`, `actions: read`, `attestations: write`. Use `actions: read` (not `write`) for least privilege.
- **Quay publish has no attestation support**: `ci_publish_quay.yml` does not use `actions/attest` and does not need `attestations` or `id-token` permissions. Quay authentication uses `QUAY_USERNAME` and `QUAY_PASSWORD` repository secrets.
- **Validate with `actionlint`**: Run `actionlint .github/workflows/*.yml` locally before pushing workflow changes. It catches invalid permission scopes, syntax errors, and expression issues that GitHub only reports at runtime.

## Active Technologies

- Go 1.26.4, multi-module workspace (`go.work`)
- OpenTelemetry Collector SDK v1.61.0 / v0.155.0 (stable + experimental series, component framework, pipeline data, processor interfaces)
- `github.com/gemaraproj/go-gemara` v0.7.0 (compliance evidence model — Gemara v1 schema)
- `github.com/telophasehq/go-ocsf` v0.2.1 (OCSF cybersecurity schema types)
- `github.com/stretchr/testify` (test assertions, proofwatch)
- OTel Weaver (semantic convention model — Go constants + docs)
- OTel Collector Builder v0.155.0 (beacon-distro binary)
- golangci-lint v2, MegaLinter, SonarCloud, gaze (quality tooling)
- Task v3 ([taskfile.dev](https://taskfile.dev)) — build automation
- Podman + podman-compose (container runtime)
- Container: UBI10 Minimal (runtime), golang:1.26.4 (build)
- Registries: `ghcr.io` (primary), Quay.io (secondary)
