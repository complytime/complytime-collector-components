## ADDED Requirements

### Requirement: Conditional policy.rule.id emission

The `GemaraEvidence.Attributes()` method SHALL emit the `policy.rule.id` OTel attribute only when the assessment's `Plan` field is non-nil. When `Plan` is nil, the attribute MUST be omitted entirely from the returned attribute set.

#### Scenario: Plan is present
- **WHEN** a `GemaraEvidence` has a non-nil `Plan` with `EntryId` set to `"deny-root-user"`
- **THEN** the returned attributes SHALL include `policy.rule.id` with value `"deny-root-user"`

#### Scenario: Plan is absent
- **WHEN** a `GemaraEvidence` has a nil `Plan`
- **THEN** the returned attributes SHALL NOT include a `policy.rule.id` entry

### Requirement: Gemara v1 SDK type alignment

All Gemara type references SHALL use the official `gemaraproj/go-gemara` v0.3.0 SDK types: `Actor` (not `Author`), `EntryMapping` (not `Mapping`), `Plan` (not `Procedure`), and `Metadata`/`AssessmentLog` from the top-level package (not `layer4` sub-package).

#### Scenario: Import path correctness
- **WHEN** any Go source file in `proofwatch` imports Gemara types
- **THEN** the import path SHALL be `github.com/gemaraproj/go-gemara` and SHALL NOT reference `github.com/ossf/gemara/layer4`

### Requirement: policy.rule.id requirement level alignment

The attribute model (`model/attributes.yaml`) SHALL declare `policy.rule.id` with `requirement_level: opt_in` to reflect that this attribute is conditionally emitted and not guaranteed to be present.

#### Scenario: Attribute model requirement level
- **WHEN** `policy.rule.id` is defined in `model/attributes.yaml`
- **THEN** its `requirement_level` SHALL be `opt_in`

### Requirement: Correct GEMARA layer numbering in attribute model

The attribute model (`model/attributes.yaml`) SHALL reference GEMARA layers according to the v1 specification: compliance attributes map to Layer 6 (Enforcement) and policy evaluation attributes map to Layer 5 (Evaluation).

#### Scenario: Compliance group layer reference
- **WHEN** the compliance attribute group description references a GEMARA layer
- **THEN** it SHALL state "GEMARA Layer 6 (Enforcement)"

#### Scenario: Policy group layer reference
- **WHEN** the policy attribute group description references a GEMARA layer
- **THEN** it SHALL state "GEMARA Layer 5 (Evaluation)"
