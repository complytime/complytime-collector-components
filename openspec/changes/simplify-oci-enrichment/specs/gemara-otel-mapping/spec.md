## REMOVED Requirements

### Requirement: OCSF evidence support

The `OCSFEvidence` type, `go-ocsf` dependency, and `transform/ocsf` collector processor SHALL be removed. Proofwatch SHALL only support Gemara-format evidence.

#### Scenario: OCSF code removed
- **WHEN** the proofwatch module is built
- **THEN** no source file SHALL import `github.com/Santiago-Labs/go-ocsf`
- **AND** no collector config SHALL reference a `transform/ocsf` processor

#### Scenario: OCSF test helpers removed
- **WHEN** `ocsf_test.go` is deleted
- **THEN** `proofwatch_test.go` SHALL use `createTestGemaraEvidence()` (from `gemara_test.go`) instead of `createTestEvidence()`

### Requirement: truthbeam processor

The `truthbeam/` module SHALL be deleted entirely. The collector SHALL NOT include a custom truthbeam processor.

#### Scenario: No custom processor in distro
- **WHEN** the collector distro is built from `collector-distro/manifest.yaml`
- **THEN** the manifest SHALL NOT reference the truthbeam module
- **AND** no collector config SHALL reference a `truthbeam` processor

### Requirement: Compass / HTTP enrichment service

The `compass` container, mTLS cert generation, generated OpenAPI client, and all Compass-related deployment infrastructure SHALL be removed.

#### Scenario: Self-contained collector
- **WHEN** the collector is deployed
- **THEN** it SHALL NOT require a `compass` or `gemara-content-service` sidecar
- **AND** `compose.yaml` SHALL NOT define a `compass` service

#### Scenario: Cert infrastructure removed
- **WHEN** the change is complete
- **THEN** `hack/self-signed-cert/openssl.cnf` SHALL be deleted
- **AND** the `generate-self-signed-cert` Makefile target SHALL be removed
- **AND** no cert volume mounts SHALL exist in `compose.yaml`

### Requirement: Catalog-derived enrichment attributes

The following attributes SHALL NOT be populated by any pipeline component in the default configuration: `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level`, `compliance.enrichment.status`.

#### Scenario: No catalog-only attributes in output
- **WHEN** a Gemara evidence log record is exported by the collector
- **THEN** it SHALL NOT contain `compliance.enrichment.status`

### Requirement: EvaluationLog fan-out

The `EvaluationLog` fan-out described in the initial draft is REMOVED from scope. complyctl plugins operate at the `AssessmentLog` + `Metadata` level and do not have access to the full `EvaluationLog` hierarchy. The existing `GemaraEvidence` type is the correct abstraction.

#### Rationale
- Plugins produce individual assessment records, not full evaluation hierarchies
- `EvaluationLog.Target` is not populated by complyctl (zero-value)
- complyctl merges multiple targets into a single `EvaluationLog`

---

## ADDED Requirements

### Requirement: Local compliance.status derivation on GemaraEvidence

Proofwatch SHALL derive `compliance.status` from the Gemara `Result` field and include it in `GemaraEvidence.Attributes()`.

#### Implementation
A `ComplianceStatus` type and `MapResult()` function SHALL be added in `proofwatch/status.go`, ported from `truthbeam/internal/applier/status.go`.

#### Scenario: Result-to-status mapping
- **WHEN** `Result` = `Passed` **THEN** `compliance.status` = `"Compliant"`
- **WHEN** `Result` = `Failed` **THEN** `compliance.status` = `"Non-Compliant"`
- **WHEN** `Result` = `NotApplicable` or `NotRun` **THEN** `compliance.status` = `"Not Applicable"`
- **WHEN** `Result` is any other value **THEN** `compliance.status` = `"Unknown"`

#### Scenario: compliance.status present in GemaraEvidence attributes
- **WHEN** a caller constructs a `GemaraEvidence` with `Result` = `Passed` and calls `Attributes()`
- **THEN** the returned attributes SHALL include `compliance.status` = `"Compliant"` alongside `policy.evaluation.result` = `"Passed"`

### Requirement: Complete Gemara-to-OTel semantic convention mapping

Each `GemaraEvidence` log record SHALL contain the following OTel attributes:

| Source | Gemara Field | OTel Attribute |
|:--|:--|:--|
| Metadata | `Author.Name` | `policy.engine.name` |
| Metadata | `Author.Version` | `policy.engine.version` |
| Metadata | `Author.Uri` | `policy.rule.uri` |
| Metadata | `Id` | `compliance.assessment.id` |
| AssessmentLog | `Requirement.EntryId` | `compliance.control.id` |
| AssessmentLog | `Requirement.ReferenceId` | `compliance.control.catalog.id` |
| AssessmentLog | `Result` | `policy.evaluation.result` |
| AssessmentLog | `Plan.EntryId` | `policy.rule.id` |
| AssessmentLog | `Message` | `policy.evaluation.message` |
| AssessmentLog | `Recommendation` | `compliance.remediation.description` |
| AssessmentLog | `Applicability` | `compliance.control.applicability` |
| *(derived)* | `MapResult(Result)` | `compliance.status` |

#### Scenario: Nil Plan omits policy.rule.id
- **WHEN** an `AssessmentLog` has `Plan` = nil
- **THEN** the emitted log record SHALL NOT contain `policy.rule.id`

#### Scenario: Empty optional fields omitted
- **WHEN** `AssessmentLog.Message` is empty
- **THEN** `policy.evaluation.message` SHALL NOT be present on the log record

### Requirement: GemaraEvidence backward compatibility

The existing `GemaraEvidence` type and its `Attributes()` method SHALL be retained and enhanced with `compliance.status`. No new evidence types are introduced.

#### Scenario: Existing GemaraEvidence still functional
- **WHEN** a caller constructs a `GemaraEvidence` and calls `Attributes()`
- **THEN** the returned attributes SHALL include all previously-emitted attributes plus `compliance.status`

### Requirement: No custom processor code in collector distro

The collector distro SHALL NOT include custom processor code. The OCB manifest SHALL only reference community-maintained processors. Non-standard receivers, exporters, extensions, and connectors are retained.

#### Scenario: Minimal logs pipeline
- **WHEN** the collector config defines a logs pipeline
- **THEN** the processors list SHALL NOT include `truthbeam` or `transform/ocsf`

### Requirement: Rename from complybeacon to complytime-collector-components

All internal references SHALL be updated to match the renamed GitHub repository.

#### Scenario: Go module path
- **WHEN** proofwatch is imported
- **THEN** the module path SHALL be `github.com/complytime/complytime-collector-components/proofwatch`

#### Scenario: Collector distro
- **WHEN** the collector distro is built
- **THEN** the directory SHALL be `collector-distro/`
- **AND** the binary SHALL be named `otelcol-complytime`
- **AND** the container image SHALL be `complytime-collector-distro`

#### Scenario: Entity name
- **WHEN** the semantic convention entity is referenced
- **THEN** the entity id SHALL be `entity.complytime.evidence` with name `complytime.evidence`

### Requirement: ClickHouse exporter

The collector distro manifest SHALL include the ClickHouse exporter component.

#### Scenario: ClickHouse in manifest
- **WHEN** the collector distro is built from `collector-distro/manifest.yaml`
- **THEN** the exporters list SHALL include `clickhouseexporter` at the same version as other contrib components

### Requirement: validate-logs CLI simplification

The `validate-logs` CLI SHALL be simplified to Gemara-only mode.

#### Scenario: No format argument
- **WHEN** `validate-logs` is invoked
- **THEN** it SHALL accept an optional output file path as its only argument
- **AND** it SHALL generate Gemara evidence logs only (no OCSF, no format flag)
- **AND** it SHALL NOT simulate truthbeam enrichment

---

## DEFERRED Requirements

### Requirement: Query-time catalog enrichment

If consumers require `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, or `compliance.risk.level`, these SHALL be resolved at query time (e.g., Grafana variable, Loki structured metadata join) rather than at pipeline time. Implementation is out of scope for this change.

### Requirement: complyctl Target population

complyctl SHALL populate `EvaluationLog.Target` with the scanned repository's `Resource` data (`Name`, `Id`, `Type`). This is an upstream dependency tracked in the complyctl repository.

### Requirement: EvaluationLog fan-out (future)

If complyctl evolves to produce per-target `EvaluationLog` objects with populated `Target` fields, proofwatch MAY add a `LogEvaluationLog()` method as an additional capability. This would walk `Evaluations[] → AssessmentLogs[]` and emit one OTel log record per assessment with inherited `Target` and `Control` context. This is deferred until the upstream data model supports it.
