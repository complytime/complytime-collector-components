## Why

The repo has four layers of runtime complexity that exist to bridge a gap proofwatch can close at the source:

1. **OCSF support** -- no downstream consumers. All evidence producers use Gemara. Dead code.
2. **truthbeam processor** -- extracts `policy.*` attributes from log records, calls Compass for compliance enrichment, writes `compliance.*` attributes back. If proofwatch emits fully-mapped OTel log records directly from Gemara types, truthbeam has nothing left to do.
3. **Compass / gemara-content-service** -- HTTP service that maps `(engineName, ruleId)` to compliance metadata from a static catalog. The four fields it provides (`category`, `frameworks`, `requirements`, `risk.level`) are catalog reference data better joined at query time. `compliance.status` is derivable locally from `Result`. No runtime service needed.
4. **mTLS cert infrastructure** -- exists solely to secure the truthbeam-to-compass link.

Additionally, the repository was renamed to `complytime-collector-components` but internal references still use `complybeacon` / `beacon`.

### Evidence model insight

complyctl plugins operate at the **AssessmentLog + Metadata** level, not the full `EvaluationLog` hierarchy. Plugins do not have access to `Target` or `ControlEvaluation` parent context. The existing `GemaraEvidence` type (flat struct: `Metadata` + `AssessmentLog`) is the correct abstraction for evidence emission. An `EvaluationLog` fan-out is not appropriate at the proofwatch layer because:

- Plugins produce individual assessment records, not full evaluation hierarchies
- The `Target` field on `EvaluationLog` is not populated by complyctl today (zero-value)
- complyctl merges multiple targets into a single `EvaluationLog`, making the hierarchy unreliable for per-target attribution

`GemaraEvidence` already carries every field needed for a complete OTel log record. The gap is only `compliance.status`, which is a pure function of `Result`.

## What Changes

**Remove:**
- `truthbeam/` module (processor, applier, client, cache, codegen, config, factory, tests)
- `proofwatch/ocsf.go`, `proofwatch/ocsf_test.go`, `go-ocsf` dependency
- `transform/ocsf` processor from all collector configs
- `compass` service from `compose.yaml`
- Cert generation from `Makefile` and `hack/self-signed-cert/`
- truthbeam from collector distro manifest
- All truthbeam/OCSF config from collector configs
- `oapi-codegen` from CI verify-codegen job (only needed for truthbeam's generated client)
- `simulateTruthBeamEnrichment` from `validate-logs` CLI

**Add:**
- `compliance.status` derivation in `GemaraEvidence.Attributes()` via ported `MapResult()` function
- `proofwatch/status.go`: `ComplianceStatus` type and `MapResult()` from truthbeam `internal/applier/status.go`
- ClickHouse exporter to collector distro manifest

**Rename (repo alignment):**
- Go module: `complybeacon` → `complytime-collector-components`
- Directory: `beacon-distro/` → `collector-distro/`
- Distro binary: `otelcol-beacon` → `otelcol-complytime`
- Container image: `complybeacon-beacon-distro` → `complytime-collector-distro`
- Semantic convention entity: `beacon.evidence` → `complytime.evidence`
- Registry manifest name: `beacon` → `complytime`
- All CI workflow job names, image references, and docs updated accordingly

**Update:**
- `collector-distro/manifest.yaml`: remove truthbeam module, add clickhouseexporter
- `collector-distro/config.yaml`: logs pipeline = `batch` only
- `hack/demo/demo-config.yaml`: remove truthbeam and transform/ocsf from pipeline
- `compose.yaml`: remove compass service, cert mounts, `depends_on`; update distro context path
- `Makefile`: remove cert generation, truthbeam from MODULES, update semantic check targets, remove truthbeam codegen
- CI: update workflows, dependabot, sonar config
- Docs: `DESIGN.md`, `DEVELOPMENT.md`, `README.md`, `proofwatch/README.md`, publish_image docs, Hyperproof integration docs

## Capabilities

### New Capabilities

- **Self-contained attribute mapping**: Proofwatch derives `compliance.status` locally from `gemara.Result` via `MapResult()`, eliminating the need for in-pipeline enrichment
- **ClickHouse export**: Collector distro includes ClickHouse exporter for analytics and long-term storage

### Removed Capabilities

- **In-pipeline enrichment**: truthbeam's `Extract → Enrich → Apply` pipeline removed entirely
- **Catalog-derived attributes**: `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level` no longer populated at pipeline time. Deferred to query-time enrichment if needed.
- **`compliance.enrichment.status`**: Removed -- no enrichment, no enrichment status
- **OCSF evidence format**: Removed -- no consumers

### Changed Capabilities

- **`GemaraEvidence.Attributes()`**: Now includes `compliance.status` in every emitted attribute set
- **`validate-logs` CLI**: Gemara-only, no format argument, no simulated enrichment

## Impact

- **proofwatch module**: `ocsf.go` deleted; `go-ocsf` dropped; `MapResult()` added in `status.go`; `GemaraEvidence.Attributes()` gains `compliance.status`; tests updated
- **truthbeam module**: Deleted entirely (~5,000 lines removed)
- **collector-distro**: Renamed from `beacon-distro`; manifest removes truthbeam, adds clickhouseexporter; custom distro retained for non-standard components but carries no custom processor code
- **Deployment**: Single collector container, no sidecar. No mTLS. `docker compose up` starts loki + grafana + collector only.
- **Attribute model**: `compliance.enrichment.status` removed; four catalog-derived attributes downgraded
- **Entity model**: `beacon.evidence` → `complytime.evidence`
- **No downstream impact**: No external OCSF consumers; no known consumers of catalog-derived attributes
