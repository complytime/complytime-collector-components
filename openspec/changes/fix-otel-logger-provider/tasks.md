# Tasks

## 1. Stable Logger Provider Migration

- [ ] 1.1 Replace the deprecated `log/global` import and
  `global.GetLoggerProvider()` call in `proofwatch/proofwatch.go` with
  `otel.GetLoggerProvider()`, then verify the default-options case in
  `TestNewProofWatch` passes.
- [ ] 1.2 Confirm the custom and nil logger-provider cases still exercise the
  intended option behavior, adding focused assertions only if needed, then
  verify `TestNewProofWatch` passes in full.

## 2. Repository Quality Gates

- [ ] 2.1 Run `task lint` and verify no SA1019 finding remains for
  `proofwatch/proofwatch.go`.
- [ ] 2.2 Run `task test` and verify the complete ProofWatch test suite passes.
