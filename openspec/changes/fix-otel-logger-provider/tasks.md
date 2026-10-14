# Tasks

## 1. Stable Logger Provider Migration and Dependency Alignment

- [x] 1.1 Replace the deprecated `log/global` import and
  `global.GetLoggerProvider()` call in `proofwatch/proofwatch.go` with
  `otel.GetLoggerProvider()`, then verify the default-options case in
  `TestNewProofWatch` passes.
- [x] 1.2 Run the existing `TestNewProofWatch` default, custom-provider, and
  nil-provider subtests and verify all pass without changing option behavior.
- [x] 1.3 Align the direct OpenTelemetry API and SDK modules on v1.47.0,
  run `go mod tidy`, confirm the direct module versions and required
  `otel/metric/x` v0.69.0 and `x/sys` v0.48.0 indirect versions, then run
  `go mod verify`.

## 2. Repository Quality Gates

- [x] 2.1 Run `task lint` and verify no SA1019 finding remains for
  `proofwatch/proofwatch.go`.
- [x] 2.2 Run `task test` and verify the complete ProofWatch test suite passes.
- [x] 2.3 Run `task version:check` and verify repository version alignment.

## Unleash Execution Checklist

- [x] Step 0: Startup Cleanup
- [x] Step 1: Branch Safety Gate
- [x] Step 2: Resumability Detection
- [x] Step 3: Clarify (Step 1) -- skipped (OpenSpec)
- [x] Step 4: Plan (Step 2) -- skipped (OpenSpec)
- [x] Step 5: Tasks (Step 3) -- skipped (OpenSpec)
- [ ] Step 6: Spec Review (Step 4) -- iteration: 1/3 advisories addressed
- [x] Step 7: Implement (Step 5) -- full v1.47.0 alignment complete
- [ ] Step 8: Code Review (Step 6) -- required for expanded scope
- [ ] Step 9: Retrospective (Step 7) -- rerun after expanded-scope review
- [ ] Step 10: Demo (Step 8)
