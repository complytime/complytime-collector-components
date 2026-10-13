# Design

## Context

`NewProofWatch` initializes default meter, logger, and tracer providers before
applying caller options. Its logger default comes from the deprecated
`go.opentelemetry.io/otel/log/global` package, while the stable OpenTelemetry API
now exposes the same process-wide provider through `otel.GetLoggerProvider()`.
See `proposal.md` for the motivation and scope.

## Goals / Non-Goals

**Goals:**

- Select the default logger provider through the stable OpenTelemetry API.
- Preserve option ordering so `WithLoggerProvider` continues to override the
  process-wide default.
- Keep default, custom, and nil-provider construction paths covered by tests.

**Non-Goals:**

- Change the public ProofWatch API or logger configuration semantics.
- Alter meter-provider or tracer-provider selection.
- Add compatibility wrappers around the deprecated package.
- Include unrelated OpenTelemetry dependency changes.

## Decisions

### Use `otel.GetLoggerProvider()` directly

Initialize `config.LoggerProvider` with `otel.GetLoggerProvider()` and remove the
deprecated `log/global` import. This mirrors the existing meter and tracer
defaults and uses the supported API without introducing an adapter.

The alternative is to retain the deprecated package behind a local helper. That
would silence call-site churn but preserve the deprecated dependency and add an
unnecessary compatibility layer.

### Preserve default-before-options initialization

Continue constructing the default config before applying `OptionFunc` values.
This guarantees that `WithLoggerProvider(customProvider)` replaces the stable
global default exactly as it does today.

The alternative is to resolve missing providers after options are applied. That
would broaden the change and could alter nil-option behavior across all provider
types.

### Verify through existing construction paths

Use the existing `TestNewProofWatch` default, custom-provider, and nil-provider
cases as regression coverage. Add only the assertions needed to prove logger
construction and custom-provider selection if the current assertions do not
distinguish those paths.

## Risks / Trade-offs

- [Risk] Tests that only assert a non-nil logger may miss use of the wrong
  provider. -> Mitigation: assert provider interaction or identity where the
  OpenTelemetry interfaces permit it without coupling to internals.
- [Risk] Expanding provider tests could turn a one-line migration into a broad
  refactor. -> Mitigation: keep changes inside `proofwatch.go` and its focused
  tests.

## Migration Plan

Replace the deprecated import and accessor, run the focused ProofWatch tests,
then run `task lint` and `task test`. Rollback is the inverse source change; no
data, configuration, or deployment migration is required.
