## Why

The repo has four layers of runtime complexity that exist to bridge a gap proofwatch can close at the source:

1. **OCSF support** -- no downstream consumers. All evidence producers use Gemara. Dead code.
2. **truthbeam processor** -- extracts `policy.*` attributes from log records, calls Compass for compliance enrichment, writes `compliance.*` attributes back. If proofwatch emits fully-mapped OTel log records directly from Gemara `EvaluationLog` types, truthbeam has nothing left to do.
3. **Compass / gemara-content-service** -- HTTP service that maps `(engineName, ruleId)` to compliance metadata from a static catalog. The four fields it provides (`category`, `frameworks`, `requirements`, `risk.level`) are catalog reference data better joined at query time. `compliance.status` is derivable locally from `Result`. No runtime service needed.
4. **mTLS cert infrastructure** -- exists solely to secure the truthbeam-to-compass link.

The Gemara `EvaluationLog` type already carries all the data needed for complete OTel semantic convention mapping:
- `EvaluationLog.Target` → `policy.target.*` attributes
- `EvaluationLog.Metadata.Author` → `policy.engine.*` attributes
- `ControlEvaluation.Control` → `compliance.control.*` attributes
- `AssessmentLog.Result` → `policy.evaluation.result` + derived `compliance.status`
- `AssessmentLog.Plan` → `policy.rule.id`
- `AssessmentLog.Message` / `Recommendation` → `policy.evaluation.message` / `compliance.remediation.description`

Proofwatch should accept an `EvaluationLog`, fan out to one OTel log record per `AssessmentLog`, propagate `Target` and `Control` context from parent objects, and emit fully-mapped semantic convention attributes. The collector becomes a stock distro: receive, batch, export.

## What Changes

**Remove:**
- `truthbeam/` module (processor, applier, client, cache, codegen, config, factory, tests)
- `proofwatch/ocsf.go`, `proofwatch/ocsf_test.go`, `go-ocsf` dependency
- `transform/ocsf` processor from all collector configs
- `compass` service from `compose.yaml`
- Compass cert generation from `Makefile` and `hack/self-signed-cert/`
- truthbeam from `beacon-distro/manifest.yaml` (OCB manifest)
- All truthbeam config from collector configs

**Add:**
- `EvaluationLog` fan-out in proofwatch: accept `gemara.EvaluationLog`, walk `Evaluations[] → AssessmentLogs[]`, emit one OTel log record per assessment with inherited `Target` and `Control` context
- `compliance.status` local derivation in proofwatch (port `mapResult()` from truthbeam `internal/applier/status.go`)
- Full Gemara-to-OTel semantic convention attribute mapping in proofwatch

**Update:**
- `beacon-distro/manifest.yaml`: remove truthbeam module
- `beacon-distro/config.yaml`: remove truthbeam and transform/ocsf processors; logs pipeline = `batch` only
- `hack/demo/demo-config.yaml`: same pipeline simplification
- `compose.yaml`: remove compass service, compass cert mounts, `depends_on`
- `Makefile`: remove cert generation, update semantic check targets
- `model/attributes.yaml`: remove `compliance.enrichment.status`; downgrade catalog-only attributes
- Docs: `DESIGN.md`, `DEVELOPMENT.md`, `README.md`

## Capabilities

### New Capabilities

- **EvaluationLog fan-out**: Proofwatch accepts a Gemara `EvaluationLog` and produces one OTel log record per `AssessmentLog`, with `Target` and `Control` context propagated from parent objects
- **Self-contained attribute mapping**: Proofwatch maps all Gemara fields to OTel semantic conventions at the source, including local `compliance.status` derivation

### Removed Capabilities

- **In-pipeline enrichment**: truthbeam's `Extract → Enrich → Apply` pipeline removed entirely
- **Catalog-derived attributes**: `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level` no longer populated at pipeline time. Deferred to query-time enrichment if needed.
- **`compliance.enrichment.status`**: Removed -- no enrichment, no enrichment status
- **OCSF evidence format**: Removed -- no consumers

## Impact

- **proofwatch module**: New `EvaluationLog` fan-out capability; `ocsf.go` deleted; `go-ocsf` dropped; `mapResult()` ported from truthbeam
- **truthbeam module**: Deleted entirely
- **beacon-distro**: `manifest.yaml` removes truthbeam module; custom distro retained for non-standard receivers/exporters/extensions but carries no custom processor code
- **Deployment**: Single collector container, no sidecar. No mTLS. `docker compose up` starts loki + grafana + collector only.
- **Attribute model**: `compliance.enrichment.status` removed; four catalog-derived attributes downgraded
- **Upstream dependency**: complyctl must populate `EvaluationLog.Target` (currently zero-valued in `evaluator.go` `GemaraLog()`)
- **No downstream impact**: No external OCSF consumers; no known consumers of catalog-derived attributes
