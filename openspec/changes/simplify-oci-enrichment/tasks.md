## 1. Proofwatch: Add compliance.status to GemaraEvidence

- [ ] 1.1 Create `proofwatch/status.go`: port `ComplianceStatus` type and `MapResult()` from `truthbeam/internal/applier/status.go`
- [ ] 1.2 Add `compliance.status` attribute to `GemaraEvidence.Attributes()` via `MapResult(g.Result).String()`
- [ ] 1.3 Add `compliance.status` assertion to `TestGemaraEvidenceAttributes` in `gemara_test.go`
- [ ] 1.4 Update `TestGemaraEvidenceAttributesDifferentResults` to verify both `policy.evaluation.result` and `compliance.status` for each result type

## 2. Remove OCSF Support

- [ ] 2.1 Delete `proofwatch/ocsf.go` and `proofwatch/ocsf_test.go`
- [ ] 2.2 Replace `createTestEvidence()` calls in `proofwatch/proofwatch_test.go` with `createTestGemaraEvidence()`
- [ ] 2.3 Remove stale comment referencing `createTestEvidence` / `ocsf_test.go` from `proofwatch_test.go`
- [ ] 2.4 Simplify `proofwatch/cmd/validate-logs/main.go`: remove OCSF imports, `generateOCSFLog()`, `simulateTruthBeamEnrichment()`, `stringPtr()`, format argument; CLI takes optional output file path only
- [ ] 2.5 Remove `go-ocsf` dependency: `go mod tidy` (removes `github.com/Santiago-Labs/go-ocsf` and transitive deps like `arrow-go`, `flatbuffers`, etc.)
- [ ] 2.6 Remove `transform/ocsf` processor from `beacon-distro/config.yaml` (both definition and pipeline reference)
- [ ] 2.7 Remove `transform/ocsf` processor from `hack/demo/demo-config.yaml` (extensive -- includes OCSF field extraction, severity mapping, S3 partitioning)

## 3. Delete truthbeam Module

- [ ] 3.1 Delete `truthbeam/` directory entirely (config, processor, factory, applier, client, metadata, tests, go.mod, go.sum, `.gaze/baseline.json`, README)
- [ ] 3.2 Remove truthbeam module from `beacon-distro/manifest.yaml`
- [ ] 3.3 Remove `truthbeam` processor config and pipeline reference from `beacon-distro/config.yaml`
- [ ] 3.4 Remove `truthbeam` processor config (endpoint, compression, TLS) and pipeline reference from `hack/demo/demo-config.yaml`
- [ ] 3.5 Remove truthbeam from `sonar-project.properties` (`sonar.modules`, `truthbeam.sonar.*` properties)
- [ ] 3.6 Update `.github/workflows/ci_sonarcloud.yml`: replace coverage merge (`cat + tail`) with single file copy
- [ ] 3.7 Update `.github/workflows/ci_local.yml`: remove truthbeam from `verify-codegen` deps loop; remove `oapi-codegen` install step (only needed for truthbeam generated client)
- [ ] 3.8 Update `.github/workflows/ci_local.yml`: update weaver-semantic-check step description (remove "OCSF")
- [ ] 3.9 Remove truthbeam from `.github/dependabot.yml` gomod directories
- [ ] 3.10 Update `Makefile` MODULES: `./proofwatch ./truthbeam` → `./proofwatch`
- [ ] 3.11 Remove `weaver-codegen` truthbeam target (`weaver registry generate ... truthbeam/internal/applier`)

## 4. Remove Compass Service and Cert Infrastructure

- [ ] 4.1 Remove `compass` service from `compose.yaml`
- [ ] 4.2 Remove `depends_on: compass` from `collector` service in `compose.yaml`
- [ ] 4.3 Remove truthbeam cert volume mount from `collector` service in `compose.yaml`
- [ ] 4.4 Delete `hack/self-signed-cert/` directory (only `openssl.cnf` is tracked; certs are gitignored)
- [ ] 4.5 Remove `generate-self-signed-cert` Makefile target and `CERT_DIR`/`OPENSSL_CNF` variables

## 5. Rename: complybeacon → complytime-collector-components

### 5.1 Go module and imports
- [ ] 5.1.1 Update `proofwatch/go.mod` module path
- [ ] 5.1.2 Update `proofwatch/proofwatch.go` import of `internal/metrics` and `ScopeName` constant
- [ ] 5.1.3 Update `proofwatch/cmd/validate-logs/main.go` import and scope name string

### 5.2 Collector distro
- [ ] 5.2.1 `git mv beacon-distro collector-distro`
- [ ] 5.2.2 Update `collector-distro/manifest.yaml`: dist name `otelcol-complytime`, module path, description, output_path `./collector`
- [ ] 5.2.3 Update `collector-distro/Containerfile.collector`: labels, binary paths (`/otelcol-complytime`, `/etc/otelcol-complytime/config.yaml`), image source URL
- [ ] 5.2.4 Update `compose.yaml` build context: `./collector-distro`

### 5.3 Semantic convention model
- [ ] 5.3.1 Update `model/entities.yaml`: `entity.beacon.evidence` → `entity.complytime.evidence`, `beacon.evidence` → `complytime.evidence`
- [ ] 5.3.2 Update `model/registry_manifest.yaml`: name `beacon` → `complytime`, description, schema_base_url

### 5.4 CI workflows
- [ ] 5.4.1 Update `.github/workflows/ci_publish_ghcr.yml`: all job names (`build-beacon-distro` → `build-collector-distro`, etc.), image name, component_name, paths, needs references, summary table
- [ ] 5.4.2 Update `.github/workflows/ci_publish_quay.yml`: job names, image names (GHCR + Quay), crane tag command
- [ ] 5.4.3 Update `.github/dependabot.yml`: remove stale "used in complybeacon" comment

### 5.5 Documentation
- [ ] 5.5.1 Rewrite `README.md`: update title, architecture overview (2 components, not 4), quick start
- [ ] 5.5.2 Rewrite `docs/DESIGN.md`: remove truthbeam/compass sections, update architecture diagram, update code examples
- [ ] 5.5.3 Rewrite `docs/DEVELOPMENT.md`: remove compass/truthbeam setup, update project structure, update clone URL, update Go workspace modules
- [ ] 5.5.4 Update `proofwatch/README.md`: update import paths, remove truthbeam/compass note, update dev guide link
- [ ] 5.5.5 Update `docs/publish_image/publish_image.md`: update cosign verify commands (image names), remove compass image note
- [ ] 5.5.6 Update `docs/integration/Sync_Evidence2Hyperproof.md`: replace `Complybeacon` references with `ComplyTime Collector`

## 6. Add ClickHouse Exporter

- [ ] 6.1 Add `clickhouseexporter` to `collector-distro/manifest.yaml` at same version as other contrib components (v0.144.0)

## 7. Update Makefile

- [ ] 7.1 Update `weaver-semantic-check` target: remove format argument, update echo message
- [ ] 7.2 Update `weaver-semantic-check-verbose` target: same changes

## 8. Update Attribute Model

- [ ] 8.1 Remove `compliance.enrichment.status` from `model/attributes.yaml`
- [ ] 8.2 Downgrade or annotate `compliance.control.category`, `compliance.frameworks`, `compliance.requirements`, `compliance.risk.level` as not populated in default mode
- [ ] 8.3 Regenerate attribute docs with `make weaver-docsgen`
- [ ] 8.4 Regenerate Go code with `make weaver-codegen`

## 9. Verification

- [ ] 9.1 `go build -mod=readonly ./...` passes for proofwatch
- [ ] 9.2 `go test -mod=readonly -count=1 ./...` passes for proofwatch (or sync vendor with `go mod vendor`)
- [ ] 9.3 `go run -mod=readonly ./cmd/validate-logs` outputs correct attributes including `compliance.status`
- [ ] 9.4 `make weaver-semantic-check` passes with Gemara-only mode
- [ ] 9.5 Collector distro builds with `collector-distro/Containerfile.collector`
- [ ] 9.6 `podman-compose up` starts without compass dependency
- [ ] 9.7 No remaining references to `ocsf`, `compass`, `truthbeam`, `beacon`, or `complybeacon` in source files (excluding openspec/ and .git/)

## 10. Upstream (complyctl -- tracked separately)

- [ ] 10.1 Update `complyctl/internal/output/evaluator.go` `GemaraLog()` to populate `EvaluationLog.Target` with scanned repository `Resource` data
- [ ] 10.2 Consider per-target `EvaluationLog` generation (one log per target instead of merging)
