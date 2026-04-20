## 1. Proofwatch: EvaluationLog Fan-out

- [ ] 1.1 Add `EvaluationLogEmitter` (or similar) that accepts `gemara.EvaluationLog` and produces `[]plog.LogRecord` (or equivalent OTel log output)
- [ ] 1.2 Implement fan-out: walk `Evaluations[] → AssessmentLogs[]`, emit one record per assessment
- [ ] 1.3 Propagate `EvaluationLog.Target` fields to `policy.target.*` attributes on each record; omit when zero-value
- [ ] 1.4 Propagate `ControlEvaluation.Control` fields to `compliance.control.id` and `compliance.control.catalog.id` on each record
- [ ] 1.5 Map `EvaluationLog.Metadata.Author` to `policy.engine.name`, `policy.engine.version`
- [ ] 1.6 Map `EvaluationLog.Metadata.Id` to `compliance.assessment.id`
- [ ] 1.7 Map assessment-level fields: `Result`, `Plan`, `Message`, `Recommendation`, `Applicability`
- [ ] 1.8 Port `mapResult()` from `truthbeam/internal/applier/status.go` to proofwatch; derive `compliance.status` locally
- [ ] 1.9 Nil-guard `Plan` -- omit `policy.rule.id` when nil (per migrate-gemara-sdk spec)
- [ ] 1.10 Omit attributes for empty optional fields (`Message`, `Recommendation`, `Target.*`)
- [ ] 1.11 Write tests: fan-out record count, target propagation, control propagation, nil Plan, zero-value Target, status derivation
- [ ] 1.12 Retain existing `GemaraEvidence` type and `Attributes()` for backward compatibility

## 2. Remove OCSF Support

- [ ] 2.1 Delete `proofwatch/ocsf.go` and `proofwatch/ocsf_test.go`
- [ ] 2.2 Remove `ocsf` and `both` modes from `proofwatch/cmd/validate-logs/main.go`; keep `gemara` only
- [ ] 2.3 Remove `go-ocsf` dependency from `proofwatch/go.mod`; run `go mod tidy`
- [ ] 2.4 Remove OCSF test helpers from `proofwatch/proofwatch_test.go` if any remain
- [ ] 2.5 Remove `transform/ocsf` processor from `beacon-distro/config.yaml`
- [ ] 2.6 Remove `transform/ocsf` processor from `hack/demo/demo-config.yaml`
- [ ] 2.7 Remove `transform/ocsf` from logs pipeline in `beacon-distro/config.yaml` service section

## 3. Delete truthbeam Module

- [ ] 3.1 Delete `truthbeam/` directory entirely (config, processor, factory, applier, client, metadata, tests, go.mod, go.sum)
- [ ] 3.2 Remove truthbeam module from `beacon-distro/manifest.yaml`
- [ ] 3.3 Remove `truthbeam` processor from `beacon-distro/config.yaml` (both processor definition and pipeline reference)
- [ ] 3.4 Remove `truthbeam` processor from `hack/demo/demo-config.yaml`
- [ ] 3.5 Remove truthbeam references from `sonar-project.properties`
- [ ] 3.6 Remove truthbeam references from `.github/workflows/ci_local.yml`
- [ ] 3.7 Remove truthbeam references from `.github/workflows/ci_crapload.yml` if present
- [ ] 3.8 Remove truthbeam test/build targets from `Makefile`
- [ ] 3.9 Remove truthbeam references from `.mega-linter.yml` if present

## 4. Remove Compass Service and Cert Infrastructure

- [ ] 4.1 Remove `compass` service from `compose.yaml`
- [ ] 4.2 Remove `depends_on: compass` from `collector` service in `compose.yaml`
- [ ] 4.3 Remove compass cert volume mounts from `collector` service in `compose.yaml`
- [ ] 4.4 Remove compass cert generation from `Makefile` (`generate-self-signed-cert` target: compass key, CSR, cert steps)
- [ ] 4.5 Clean `hack/self-signed-cert/openssl.cnf` of compass-specific entries
- [ ] 4.6 Remove `hack/demo/config.yaml` if it was compass-specific
- [ ] 4.7 Remove compass-related sample data (`hack/sampledata/osps.yaml` if still present)

## 5. Update Collector Configs

- [ ] 5.1 Simplify `beacon-distro/config.yaml` logs pipeline to `receivers: [otlp] → processors: [batch] → exporters: [debug]`
- [ ] 5.2 Simplify `hack/demo/demo-config.yaml` logs pipeline (remove truthbeam and transform/ocsf)
- [ ] 5.3 Remove truthbeam endpoint/TLS/cache config from all collector configs
- [ ] 5.4 Update `beacon-distro/Containerfile.collector` if it references compass certs or truthbeam config

## 6. Update Makefile

- [ ] 6.1 Remove or simplify `generate-self-signed-cert` target (may still need truthbeam CA cert for other TLS uses, or remove entirely)
- [ ] 6.2 Update `weaver-semantic-check` target: change `both` to `gemara`, update echo messages
- [ ] 6.3 Update `weaver-semantic-check-verbose` target: same changes
- [ ] 6.4 Remove any truthbeam-specific build/test/lint targets

## 7. Update Attribute Model

- [ ] 7.1 Remove `compliance.enrichment.status` from `model/attributes.yaml`
- [ ] 7.2 Downgrade or annotate `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level` as not populated in default mode
- [ ] 7.3 Regenerate attribute docs with `make weaver-docsgen`

## 8. Documentation

- [ ] 8.1 Update `docs/DESIGN.md`: remove truthbeam/compass architecture, update pipeline diagram, update `GemaraEvidence` example
- [ ] 8.2 Update `docs/DEVELOPMENT.md`: remove compass setup instructions, simplify collector startup
- [ ] 8.3 Update `README.md`: update architecture overview, remove compass/truthbeam references
- [ ] 8.4 Update `docs/publish_image/publish_image.md`: remove compass image references
- [ ] 8.5 Update `truthbeam/README.md` → delete (module is gone)
- [ ] 8.6 Update `proofwatch/README.md`: document EvaluationLog fan-out capability
- [ ] 8.7 Update `docs/integration/Sync_Evidence2Hyperproof.md` if it references OCSF or compass

## 9. Verification

- [ ] 9.1 `go build ./...` passes for proofwatch
- [ ] 9.2 `go test ./...` passes for proofwatch
- [ ] 9.3 `make weaver-semantic-check` passes with Gemara-only mode
- [ ] 9.4 Collector distro builds successfully with `beacon-distro/Containerfile.collector`
- [ ] 9.5 `docker compose up` starts without compass dependency
- [ ] 9.6 No remaining references to `ocsf`, `compass`, or `truthbeam` in source files (grep verification)

## 10. Upstream (complyctl -- tracked separately)

- [ ] 10.1 Update `complyctl/internal/output/evaluator.go` `GemaraLog()` to populate `EvaluationLog.Target` with scanned repository `Resource` data
- [ ] 10.2 Consider per-target `EvaluationLog` generation (one log per target instead of merging)
