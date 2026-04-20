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

Meanwhile, complyctl already produces full `gemara.EvaluationLog` objects with the complete hierarchy:

```
EvaluationLog
  ├── Metadata (Author/Actor, Id)
  ├── Target (Resource: Name, Id, Type, Environment)
  └── Evaluations[]
        ├── Control (EntryMapping: EntryId, ReferenceId)
        └── AssessmentLogs[]
              ├── Requirement (EntryMapping)
              ├── Plan (*EntryMapping, optional)
              ├── Result, Message, Recommendation
              └── Applicability, ConfidenceLevel, Start/End
```

This hierarchy carries every field needed for complete OTel semantic convention mapping. The `Target` and `Control` context on the outer objects propagate naturally to each `AssessmentLog`. There is no data that requires a runtime catalog lookup -- everything is available at the point of evidence emission.

### Data mapping analysis

**Fields available from Gemara EvaluationLog hierarchy (no external lookup):**

| Source | Gemara Field | OTel Attribute |
|:--|:--|:--|
| EvaluationLog | `Target.Name` | `policy.target.name` |
| EvaluationLog | `Target.Id` | `policy.target.id` |
| EvaluationLog | `Target.Type` | `policy.target.type` |
| EvaluationLog | `Target.Environment` | `policy.target.environment` |
| EvaluationLog | `Metadata.Author.Name` | `policy.engine.name` |
| EvaluationLog | `Metadata.Author.Version` | `policy.engine.version` |
| EvaluationLog | `Metadata.Author.Uri` | `policy.rule.uri` |
| EvaluationLog | `Metadata.Id` | `compliance.assessment.id` |
| ControlEvaluation | `Control.EntryId` | `compliance.control.id` |
| ControlEvaluation | `Control.ReferenceId` | `compliance.control.catalog.id` |
| AssessmentLog | `Result` | `policy.evaluation.result` |
| AssessmentLog | `Plan.EntryId` | `policy.rule.id` (opt_in, nil-guarded) |
| AssessmentLog | `Message` | `policy.evaluation.message` |
| AssessmentLog | `Recommendation` | `compliance.remediation.description` |
| AssessmentLog | `Applicability` | `compliance.control.applicability` |
| *(derived)* | `mapResult(Result)` | `compliance.status` |

**Fields that required Compass (catalog-only, no longer populated):**

| Field | OTel Attribute | Disposition |
|:--|:--|:--|
| `Control.Category` | `compliance.control.category` | Deferred to query-time enrichment |
| `Frameworks.Frameworks` | `compliance.frameworks` | Deferred to query-time enrichment |
| `Frameworks.Requirements` | `compliance.requirements` | Deferred to query-time enrichment |
| `Risk.Level` | `compliance.risk.level` | Deferred to query-time enrichment |
| `EnrichmentStatus` | `compliance.enrichment.status` | Removed |

Every attribute that matters for auditor workflows (control ID, catalog ID, target name, pass/fail status) is available from the Gemara hierarchy. The catalog-derived cross-references are reference data suitable for query-time joins.

## Goals / Non-Goals

**Goals:**
- Proofwatch accepts `gemara.EvaluationLog` and fans out to one OTel log record per `AssessmentLog` with full semantic convention attributes
- Remove truthbeam processor, OCSF support, Compass service, and cert infrastructure
- Collector retains custom OCB distro (non-standard components) but drops all custom processor code
- Maintain fail-open behavior (malformed records pass through with warning)

**Non-Goals:**
- Adding new Gemara fields or modifying the `go-gemara` SDK
- Query-time catalog enrichment tooling (future work)
- Changing complyctl's `EvaluationLog` generation (upstream dependency, tracked separately)
- Supporting non-Gemara evidence formats

## Decisions

### 1. Proofwatch owns the Gemara-to-OTel mapping

Move all semantic convention attribute mapping to proofwatch. The `EvaluationLog` fan-out works as follows:

For each `ControlEvaluation` in `EvaluationLog.Evaluations`:
  For each `AssessmentLog` in `ControlEvaluation.AssessmentLogs`:
    Emit one OTel log record with:
    - `Target` attributes from `EvaluationLog.Target`
    - `Control` attributes from `ControlEvaluation.Control`
    - Assessment attributes from the `AssessmentLog` itself
    - `compliance.status` derived from `AssessmentLog.Result`

This replaces the current `GemaraEvidence` type (flat struct) with a richer input that preserves the full hierarchy context.

**Alternatives considered:**
- *Keep truthbeam as mapper, proofwatch does fan-out only*: proofwatch would emit records with Gemara fields as raw attributes, truthbeam would rename them to semantic conventions. Rejected -- adds a pass-through processor for pure field renaming that proofwatch can do at the source.
- *Fan out in a collector processor instead of proofwatch*: Would require truthbeam to parse structured Gemara JSON from log bodies. Rejected -- moves complexity into the pipeline that belongs at the evidence producer.

### 2. Delete truthbeam entirely

With proofwatch emitting fully-mapped records, truthbeam has zero remaining function. A pass-through processor adds build complexity (custom collector distro), config surface, and a test surface for no value.

The collector remains a custom OCB-built distro (non-standard receivers, exporters, and extensions are still needed), but `beacon-distro/manifest.yaml` drops the truthbeam module. No custom Go processor code runs in the pipeline.

**Alternatives considered:**
- *Keep truthbeam as optional enrichment processor for future Path B*: Would require maintaining the module, tests, and custom distro build even when unused. Rejected -- if Path B is needed later, it can be re-added. Dead code has carrying cost.

### 3. Port `mapResult()` to proofwatch

The `Result` → `compliance.status` mapping logic (`Passed` → `Compliant`, `Failed` → `Non-Compliant`, etc.) currently lives in `truthbeam/internal/applier/status.go`. Port this function to proofwatch where it will be used during the fan-out.

### 4. Retain `GemaraEvidence` for backward compatibility

Keep the existing `GemaraEvidence` type and its `Attributes()` method for callers that produce individual assessment records (not full EvaluationLogs). The new `EvaluationLog` fan-out is an additional capability, not a replacement. `cmd/validate-logs` uses `GemaraEvidence` directly for semantic convention validation.

### 5. Remove OCSF, Compass, and cert infrastructure

No migration path. No consumers. Clean delete.

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

### 7. Upstream dependency: complyctl Target population

complyctl's `Evaluator.GemaraLog()` currently returns `EvaluationLog` with a zero-value `Target` field. For the fan-out to produce useful `policy.target.*` attributes, complyctl must populate `Target` with the scanned repository's `Resource` data. This is tracked as an upstream dependency, not blocked by this change -- proofwatch handles a zero-value `Target` gracefully by omitting `policy.target.*` attributes.

## Open Decisions

### 8. Attribute namespace: keep policy/compliance dichotomy or unify under Gemara-native naming?

The current attribute model splits into two namespaces:
- `policy.*` -- engine, rule, evaluation result, target
- `compliance.*` -- control, catalog, status, remediation, assessment ID

This split originated from the old architecture where `policy.*` attributes came from the evidence producer and `compliance.*` attributes were added by Compass enrichment. With enrichment removed and all data sourced from a single Gemara `EvaluationLog`, the boundary is an artifact of the pipeline, not the data model.

Examples of the mismatch:

| Observation | Issue |
|:--|:--|
| `AssessmentLog.Result` → `policy.evaluation.result` but `mapResult(Result)` → `compliance.status` | Same field, two namespaces |
| `Metadata.Id` → `compliance.assessment.id` but `Metadata.Author.Name` → `policy.engine.name` | Same parent struct, two namespaces |
| `Control.EntryId` → `compliance.control.id` but `Target.Name` → `policy.target.name` | Parallel Gemara concepts, different namespaces |

**Option A: Keep the dichotomy**

- No breaking change to existing configs, queries, dashboards
- `policy` and `compliance` are arguably two semantic domains even if the data source is one
- Aligns with OTel convention of domain-based namespacing (`http.*`, `db.*`, `rpc.*`)
- Cost: new contributors must learn an arbitrary split that doesn't match the Gemara type hierarchy

**Option B: Unify under Gemara-native namespace**

Example mapping:

| Gemara Source | Attribute |
|:--|:--|
| `EvaluationLog.Metadata.Author.Name` | `gemara.author.name` |
| `EvaluationLog.Metadata.Id` | `gemara.evaluation.id` |
| `EvaluationLog.Target.Name` | `gemara.target.name` |
| `ControlEvaluation.Control.EntryId` | `gemara.control.id` |
| `ControlEvaluation.Control.ReferenceId` | `gemara.control.catalog_id` |
| `AssessmentLog.Result` | `gemara.assessment.result` |
| `AssessmentLog.Plan.EntryId` | `gemara.plan.id` |
| `AssessmentLog.Message` | `gemara.assessment.message` |
| `AssessmentLog.Recommendation` | `gemara.assessment.recommendation` |
| `AssessmentLog.Applicability` | `gemara.assessment.applicability` |
| *(derived)* | `gemara.assessment.status` |

- Attribute names mirror the Gemara SDK type hierarchy directly
- Single namespace, simpler mental model
- Eliminates the artificial policy/compliance boundary
- Cost: breaks every existing query, config, and dashboard that references `policy.*` or `compliance.*`
- Cost: `gemara.*` namespace is project-specific, not domain-generic -- may limit adoption by tools that don't use Gemara

**Option C: Hybrid -- domain namespace, Gemara-aligned structure**

Keep domain-generic names (`assessment.*`, `control.*`, `target.*`) but restructure to follow the Gemara hierarchy rather than the old pipeline boundary:

| Gemara Source | Attribute |
|:--|:--|
| `EvaluationLog.Metadata.Author.Name` | `assessment.engine.name` |
| `EvaluationLog.Metadata.Id` | `assessment.id` |
| `EvaluationLog.Target.Name` | `target.name` |
| `ControlEvaluation.Control.EntryId` | `control.id` |
| `AssessmentLog.Result` | `assessment.result` |
| `AssessmentLog.Plan.EntryId` | `assessment.plan.id` |
| *(derived)* | `assessment.status` |

- Domain-generic (not Gemara-specific), could be adopted by non-Gemara tools
- Follows the data model, not the pipeline
- Still a breaking change to existing configs

**Status:** Open for discussion. This decision affects the attribute model (`model/attributes.yaml`), all generated docs, proofwatch's attribute constants, and any downstream query or dashboard. Should be resolved before implementation begins.

## Risks / Trade-offs

- **[Catalog-derived attributes dropped]** → `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level` no longer populated. No known consumers. If needed, query-time enrichment is the appropriate pattern for static reference data.
- **[Custom distro still needed]** → Even without truthbeam, the collector distro may include non-standard receivers/exporters (webhook, S3, signaltometrics). The custom build remains, but with no custom processor code.
- **[complyctl Target not populated]** → Until complyctl is updated, `policy.target.*` attributes will be absent. Fan-out still works -- records just lack target context. Proofwatch does not fail on zero-value Target.
- **[Two evidence APIs in proofwatch]** → `GemaraEvidence` (flat) and `EvaluationLog` fan-out (hierarchical) coexist. This is intentional -- different callers have different needs. The flat API remains useful for simple single-assessment evidence and for `cmd/validate-logs`.
