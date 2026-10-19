# Proposal

## Why

The OpenTelemetry Go API v1.47.0 deprecates the global logger-provider package
used by ProofWatch, causing `golangci-lint` to report SA1019. ProofWatch must use
the stable API before the dependency upgrade can pass the repository's quality
gates.

## What Changes

- Replace the deprecated `go.opentelemetry.io/otel/log/global` usage with the
  stable logger-provider accessor exposed by `go.opentelemetry.io/otel`.
- Align the directly used OpenTelemetry API and SDK modules on v1.47.0,
  consolidating the dependency updates from PRs #452 through #457.
- Preserve the existing `NewProofWatch` option behavior so an explicitly
  configured logger provider continues to override the global default.
- Run the existing default, custom, and nil logger-provider construction tests.
- Keep the change limited to the OpenTelemetry v1.47.0 compatibility fix.

## Capabilities

### New Capabilities

_(none - this change does not introduce new behavior)_

### Modified Capabilities

_(none - this is an implementation compatibility fix with no requirement changes)_

## Impact

- `proofwatch/proofwatch.go`: imports and default logger-provider selection.
- `proofwatch/go.mod` and `proofwatch/go.sum`: aligned OpenTelemetry v1.47.0
  API and SDK dependencies, including required indirect updates.
- Existing `proofwatch/proofwatch_test.go` construction tests remain the
  regression gate; no new behavior or test framework is required.
- OpenTelemetry Go API v1.47.0 compatibility and the `task lint` and `task test`
  quality gates.
- No public ProofWatch API or generated files are changed.
