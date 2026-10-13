# Tasks

## 1. Stable Logger Provider Migration

- [x] 1.1 Replace the deprecated `log/global` import and
  `global.GetLoggerProvider()` call in `proofwatch/proofwatch.go` with
  `otel.GetLoggerProvider()`, then verify the default-options case in
  `TestNewProofWatch` passes.
- [x] 1.2 Run the existing `TestNewProofWatch` default, custom-provider, and
  nil-provider subtests and verify all pass without changing option behavior.

## 2. Repository Quality Gates

- [x] 2.1 Run `task lint` and verify no SA1019 finding remains for
  `proofwatch/proofwatch.go`.
- [x] 2.2 Run `task test` and verify the complete ProofWatch test suite passes.

## Unleash Execution Checklist

- [x] Step 0: Startup Cleanup
- [x] Step 1: Branch Safety Gate
- [x] Step 2: Resumability Detection
- [x] Step 3: Clarify (Step 1) -- skipped (OpenSpec)
- [x] Step 4: Plan (Step 2) -- skipped (OpenSpec)
- [x] Step 5: Tasks (Step 3) -- skipped (OpenSpec)
- [x] Step 6: Spec Review (Step 4) -- iteration: 2/3 PASSED
- [x] Step 7: Implement (Step 5) -- all 2 phases complete, lint/test PASS
- [ ] Step 8: Code Review (Step 6)
- [ ] Step 9: Retrospective (Step 7)
- [ ] Step 10: Demo (Step 8)

<!-- spec-review: passed -->
