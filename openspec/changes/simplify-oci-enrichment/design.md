## Context

ComplyBeacon's current architecture has four runtime components in the log pipeline:

```
proofwatch (client lib) → OTLP → collector:
  transform/ocsf → truthbeam → [exporters]
                       ↓
              compass (HTTP, mTLS)
```

Proofwatch emits flat `GemaraEvidence` records (an `AssessmentLog` + `Metadata`) with a subset of OTel semantic convention attributes. The collector's `transform/ocsf` processor handles OCSF-format logs (no consumers exist). Truthbeam extracts `policy.*` attributes, calls Compass for compliance enrichment, and writes `compliance.*` attributes back.

Compass is a stateless HTTP wrapper around a Gemara catalog YAML file. The four fields it adds (`category`, `frameworks`, `requirements`, `risk.level`) are static catalog reference data. `compliance.status` is a pure function of `Result`.

### Evidence model: GemaraEvidence is the right abstraction

complyctl plugins operate at the **AssessmentLog + Metadata** level. They do not have access to the full `EvaluationLog` hierarchy (`Target`, `ControlEvaluation` parent context). An `EvaluationLog` fan-out at the proofwatch layer was considered and rejected for three reasons:

1. **Plugin scope**: Plugins produce individual assessment records, not full evaluation hierarchies. The `GemaraEvidence` type (flat: `Metadata` + `AssessmentLog`) matches what producers actually have.
2. **Target reliability**: complyctl merges multiple targets into a single `EvaluationLog` with a zero-value `Target` field. The hierarchy is not reliable for per-target attribution today.
3. **Simplicity**: Adding `compliance.status` to `GemaraEvidence.Attributes()` is a one-line change. A fan-out adds a new type, new tests, and a new API surface for no gain given the current producer constraints.

If complyctl evolves to produce per-target `EvaluationLog` objects with populated `Target` fields, a fan-out API can be added later as an additional capability.

### Data mapping analysis

**Fields available from GemaraEvidence (no external lookup):**

| Source | Gemara Field | OTel Attribute |
|:--|:--|:--|
| Metadata | `Author.Name` | `policy.engine.name` |
| Metadata | `Author.Version` | `policy.engine.version` |
| Metadata | `Author.Uri` | `policy.rule.uri` |
| Metadata | `Id` | `compliance.assessment.id` |
| AssessmentLog | `Requirement.EntryId` | `compliance.control.id` |
| AssessmentLog | `Requirement.ReferenceId` | `compliance.control.catalog.id` |
| AssessmentLog | `Result` | `policy.evaluation.result` |
| AssessmentLog | `Plan.EntryId` | `policy.rule.id` (opt_in, nil-guarded) |
| AssessmentLog | `Message` | `policy.evaluation.message` |
| AssessmentLog | `Recommendation` | `compliance.remediation.description` |
| AssessmentLog | `Applicability` | `compliance.control.applicability` |
| *(derived)* | `MapResult(Result)` | `compliance.status` |

**Fields that required Compass (catalog-only, no longer populated):**

| Field | OTel Attribute | Disposition |
|:--|:--|:--|
| `Control.Category` | `compliance.control.category` | Deferred to query-time enrichment |
| `Frameworks.Frameworks` | `compliance.frameworks` | Deferred to query-time enrichment |
| `Frameworks.Requirements` | `compliance.requirements` | Deferred to query-time enrichment |
| `Risk.Level` | `compliance.risk.level` | Deferred to query-time enrichment |
| `EnrichmentStatus` | `compliance.enrichment.status` | Removed |

Every attribute that matters for auditor workflows (control ID, catalog ID, pass/fail status) is available from `GemaraEvidence`. The catalog-derived cross-references are reference data suitable for query-time joins.

## Goals / Non-Goals

**Goals:**
- Proofwatch emits `compliance.status` on every `GemaraEvidence` log record via local `MapResult()` derivation
- Remove truthbeam processor, OCSF support, Compass service, and cert infrastructure
- Rename all internal references from `complybeacon`/`beacon` to `complytime-collector-components`/`complytime`
- Collector retains custom OCB distro (non-standard components) but drops all custom processor code
- Add ClickHouse exporter to distro manifest
- Maintain fail-open behavior (malformed records pass through with warning)

**Non-Goals:**
- `EvaluationLog` fan-out in proofwatch (producers don't have that context today)
- Adding new Gemara fields or modifying the `go-gemara` SDK
- Query-time catalog enrichment tooling (future work)
- Changing complyctl's `EvaluationLog` generation (upstream dependency, tracked separately)
- Supporting non-Gemara evidence formats

## Decisions

### 1. Proofwatch adds compliance.status to GemaraEvidence

Add `MapResult()` to proofwatch (ported from `truthbeam/internal/applier/status.go`) and call it in `GemaraEvidence.Attributes()`. This is a single new attribute on an existing type -- no new types, no new API surface.

The `MapResult()` function maps:
- `Passed` → `Compliant`
- `Failed` → `Non-Compliant`
- `NotApplicable`, `NotRun` → `Not Applicable`
- default → `Unknown`

Lives in `proofwatch/status.go` as an exported function with an exported `ComplianceStatus` type.

**Alternatives considered:**
- *EvaluationLog fan-out*: Rejected. complyctl plugins only produce AssessmentLog + Metadata. The full hierarchy is not available to evidence producers. If this changes upstream, a fan-out can be added as an additional capability without breaking the flat API.
- *Keep truthbeam for status derivation only*: Rejected. A custom collector processor for one `switch` statement is not justified.

### 2. Delete truthbeam entirely

With proofwatch emitting `compliance.status` directly, truthbeam has zero remaining function. The collector remains a custom OCB-built distro (non-standard receivers, exporters, and extensions are still needed), but the manifest drops the truthbeam module.

### 3. Retain GemaraEvidence as the primary evidence type

The existing `GemaraEvidence` struct (explicit `Metadata` field + embedded `AssessmentLog`) is the right level of abstraction. It now emits `compliance.status` in addition to its existing attributes. The `validate-logs` CLI uses it directly for semantic convention validation.

### 4. Remove OCSF, Compass, and cert infrastructure

No migration path needed. No consumers. Clean delete.

**Implementation detail**: `proofwatch/proofwatch_test.go` shares a `createTestEvidence()` helper defined in `ocsf_test.go`. This must be replaced with `createTestGemaraEvidence()` (from `gemara_test.go`) before deleting OCSF test files.

### 5. Rename from complybeacon to complytime-collector-components

The GitHub repo was renamed. All internal references update:

| Old | New |
|:--|:--|
| `github.com/complytime/complybeacon/proofwatch` | `github.com/complytime/complytime-collector-components/proofwatch` |
| `beacon-distro/` | `collector-distro/` |
| `otelcol-beacon` | `otelcol-complytime` |
| `complybeacon-beacon-distro` (image) | `complytime-collector-distro` (image) |
| `beacon.evidence` (entity) | `complytime.evidence` |
| `ComplyBeacon` (prose) | `ComplyTime Collector` |

This touches Go module paths, ScopeName constants, imports, CI workflow job names, image names (GHCR + Quay), Containerfile labels, compose context paths, dependabot config, sonar config, and all docs.

### 6. Collector pipeline simplification

Before:
```yaml
logs:
  receivers: [otlp]
  processors: [batch, transform/ocsf, truthbeam]
  exporters: [debug]
```

After:
```yaml
logs:
  receivers: [otlp]
  processors: [batch]
  exporters: [debug]
```

### 7. Add ClickHouse exporter

Add `clickhouseexporter` to the OCB manifest at the same version as other contrib components (`v0.144.0`). Enables direct log export to ClickHouse for analytics and long-term storage.

### 8. validate-logs CLI simplification

Before: accepts `gemara|ocsf|both` format argument, simulates truthbeam enrichment.
After: no format argument (Gemara-only), no enrichment simulation, simplified main function.

## Open Decisions

### 9. Attribute namespace: keep policy/compliance dichotomy or unify under Gemara-native naming?

The current attribute model splits into two namespaces:
- `policy.*` -- engine, rule, evaluation result, target
- `compliance.*` -- control, catalog, status, remediation, assessment ID

This split originated from the old architecture where `policy.*` attributes came from the evidence producer and `compliance.*` attributes were added by Compass enrichment. With enrichment removed and all data sourced from `GemaraEvidence`, the boundary is an artifact of the pipeline, not the data model.

Examples of the mismatch:

| Observation | Issue |
|:--|:--|
| `AssessmentLog.Result` → `policy.evaluation.result` but `MapResult(Result)` → `compliance.status` | Same field, two namespaces |
| `Metadata.Id` → `compliance.assessment.id` but `Metadata.Author.Name` → `policy.engine.name` | Same parent struct, two namespaces |

**Option A: Keep the dichotomy** -- No breaking change. `policy` and `compliance` are arguably two semantic domains. Aligns with OTel domain-based namespacing.

**Option B: Unify under Gemara-native namespace** (`gemara.*`) -- Mirrors SDK hierarchy. Breaking change to all queries/dashboards.

**Option C: Hybrid domain-generic** (`assessment.*`, `control.*`, `target.*`) -- Follows data model, not pipeline. Still a breaking change.

**Status:** Open for discussion. The implementation proceeds with the existing dichotomy (Option A) since `proofwatch/attributes.go` constants are centralized and easily changed.

## Risks / Trade-offs

- **[Catalog-derived attributes dropped]** → No known consumers. Query-time enrichment is the appropriate pattern for static reference data.
- **[Custom distro still needed]** → Non-standard receivers/exporters (webhook, S3, ClickHouse, signaltometrics). Custom build remains, but with no custom processor code.
- **[Two evidence APIs not needed]** → The original spec proposed `GemaraEvidence` (flat) + `EvaluationLog` fan-out (hierarchical). Implementation showed only the flat API is needed given current producer constraints. Scope reduced.
- **[Test helper dependency]** → `proofwatch_test.go` uses `createTestEvidence()` from `ocsf_test.go`. Must replace with `createTestGemaraEvidence()` before deleting OCSF files.
- **[Vendor directory staleness]** → After removing `go-ocsf`, the vendor directory is stale. Use `-mod=readonly` for builds/tests, or run `go mod vendor` to sync.
- **[Entity rename]** → `beacon.evidence` → `complytime.evidence` in `model/entities.yaml`. No external consumers of this entity ID.
