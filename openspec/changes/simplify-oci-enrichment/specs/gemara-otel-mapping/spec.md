## REMOVED Requirements

### Requirement: OCSF evidence support

The `OCSFEvidence` type, `go-ocsf` dependency, and `transform/ocsf` collector processor SHALL be removed. Proofwatch SHALL only support Gemara-format evidence.

#### Scenario: OCSF code removed
- **WHEN** the proofwatch module is built
- **THEN** no source file SHALL import `github.com/Santiago-Labs/go-ocsf`
- **AND** no collector config SHALL reference a `transform/ocsf` processor

### Requirement: truthbeam processor

The `truthbeam/` module SHALL be deleted entirely. The collector SHALL NOT include a custom truthbeam processor.

#### Scenario: No custom processor in distro
- **WHEN** the collector distro is built from `beacon-distro/manifest.yaml`
- **THEN** the manifest SHALL NOT reference the truthbeam module
- **AND** no collector config SHALL reference a `truthbeam` processor

### Requirement: Compass / HTTP enrichment service

The `compass` container, mTLS cert generation, generated OpenAPI client, and all Compass-related deployment infrastructure SHALL be removed.

#### Scenario: Self-contained collector
- **WHEN** the collector is deployed
- **THEN** it SHALL NOT require a `compass` or `gemara-content-service` sidecar
- **AND** `compose.yaml` SHALL NOT define a `compass` service

### Requirement: Catalog-derived enrichment attributes

The following attributes SHALL NOT be populated by any pipeline component in the default configuration: `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level`, `compliance.enrichment.status`.

#### Scenario: No catalog-only attributes in output
- **WHEN** a Gemara evidence log record is exported by the collector
- **THEN** it SHALL NOT contain `compliance.enrichment.status`

---

## ADDED Requirements

### Requirement: EvaluationLog fan-out

Proofwatch SHALL accept a `gemara.EvaluationLog` and produce one OTel log record per `AssessmentLog` entry. Each record SHALL carry context propagated from the parent `EvaluationLog` and `ControlEvaluation` objects.

#### Scenario: Fan-out produces correct record count
- **WHEN** an `EvaluationLog` contains 2 `ControlEvaluation` entries, each with 3 `AssessmentLog` entries
- **THEN** proofwatch SHALL emit exactly 6 OTel log records

#### Scenario: Target context propagation
- **WHEN** an `EvaluationLog` has `Target.Name` = `"org-infra"` and `Target.Id` = `"complytime/org-infra"`
- **THEN** every emitted log record SHALL contain `policy.target.name` = `"org-infra"` and `policy.target.id` = `"complytime/org-infra"`

#### Scenario: Control context propagation
- **WHEN** a `ControlEvaluation` has `Control.EntryId` = `"OSPS-QA-07"` and `Control.ReferenceId` = `"OSPS-B"`
- **THEN** every log record emitted from that evaluation's `AssessmentLogs` SHALL contain `compliance.control.id` = `"OSPS-QA-07"` and `compliance.control.catalog.id` = `"OSPS-B"`

#### Scenario: Zero-value Target handled gracefully
- **WHEN** an `EvaluationLog` has a zero-value `Target` (empty `Name`, `Id`, etc.)
- **THEN** the emitted log records SHALL omit `policy.target.*` attributes rather than emitting empty strings

### Requirement: Complete Gemara-to-OTel semantic convention mapping

Each emitted log record SHALL contain the following OTel attributes mapped from the Gemara hierarchy:

| Source | Gemara Field | OTel Attribute |
|:--|:--|:--|
| EvaluationLog | `Target.Name` | `policy.target.name` |
| EvaluationLog | `Target.Id` | `policy.target.id` |
| EvaluationLog | `Target.Type` | `policy.target.type` |
| EvaluationLog | `Target.Environment` | `policy.target.environment` |
| EvaluationLog | `Metadata.Author.Name` | `policy.engine.name` |
| EvaluationLog | `Metadata.Author.Version` | `policy.engine.version` |
| EvaluationLog | `Metadata.Id` | `compliance.assessment.id` |
| ControlEvaluation | `Control.EntryId` | `compliance.control.id` |
| ControlEvaluation | `Control.ReferenceId` | `compliance.control.catalog.id` |
| AssessmentLog | `Result` | `policy.evaluation.result` |
| AssessmentLog | `Plan.EntryId` | `policy.rule.id` |
| AssessmentLog | `Message` | `policy.evaluation.message` |
| AssessmentLog | `Recommendation` | `compliance.remediation.description` |
| AssessmentLog | `Applicability` | `compliance.control.applicability` |
| *(derived)* | `mapResult(Result)` | `compliance.status` |

#### Scenario: Nil Plan omits policy.rule.id
- **WHEN** an `AssessmentLog` has `Plan` = nil
- **THEN** the emitted log record SHALL NOT contain `policy.rule.id`

#### Scenario: Empty optional fields omitted
- **WHEN** `AssessmentLog.Message` is empty
- **THEN** `policy.evaluation.message` SHALL NOT be present on the log record

### Requirement: Local compliance.status derivation

Proofwatch SHALL derive `compliance.status` from the Gemara `Result` field using a local mapping function.

#### Scenario: Result-to-status mapping
- **WHEN** `Result` = `Passed` **THEN** `compliance.status` = `"Compliant"`
- **WHEN** `Result` = `Failed` **THEN** `compliance.status` = `"Non-Compliant"`
- **WHEN** `Result` = `NotApplicable` or `NotRun` **THEN** `compliance.status` = `"Not Applicable"`
- **WHEN** `Result` = `Unknown` **THEN** `compliance.status` = `"Unknown"`

### Requirement: GemaraEvidence backward compatibility

The existing `GemaraEvidence` type and its `Attributes()` method SHALL be retained for callers that produce individual assessment records. The `EvaluationLog` fan-out is an additional capability.

#### Scenario: Existing GemaraEvidence still functional
- **WHEN** a caller constructs a `GemaraEvidence` and calls `Attributes()`
- **THEN** the returned attributes SHALL match the current behavior (per migrate-gemara-sdk spec)

### Requirement: No custom processor code in collector distro

The collector distro SHALL NOT include custom processor code. The OCB manifest SHALL only reference community-maintained processors. Non-standard receivers, exporters, and extensions are retained.

#### Scenario: Minimal logs pipeline
- **WHEN** the collector config defines a logs pipeline
- **THEN** the processors list SHALL NOT include `truthbeam` or `transform/ocsf`

---

## DEFERRED Requirements

### Requirement: Query-time catalog enrichment

If consumers require `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, or `compliance.risk.level`, these SHALL be resolved at query time (e.g., Grafana variable, Loki structured metadata join) rather than at pipeline time. Implementation is out of scope for this change.

### Requirement: complyctl Target population

complyctl SHALL populate `EvaluationLog.Target` with the scanned repository's `Resource` data (`Name`, `Id`, `Type`). This is an upstream dependency tracked in the complyctl repository.
