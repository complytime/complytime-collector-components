# Design

## Context

See `proposal.md` for motivation. The repository's tracked root `go.mod` already declares the Go version used by the other tracked module declarations and the collector build image. Today, `version:check` invokes workspace setup and the check script reads both its Go version and collector-module inventory from `go.work`, which may be generated and untracked.

## Goals / Non-Goals

**Goals:**

- Make `task version:check` read-only with respect to the working tree.
- Base Go-version alignment on tracked repository inputs.
- Preserve the existing version and OTel validation coverage.
- Add regression tests for workspace-file and runtime-version independence.

**Non-Goals:**

- Change workspace setup for builds, dependency operations, or other commands that require it.
- Change Renovate grouping, which remains covered by issue #464.
- Alter the independent OTel version-management policy.

## Decisions

### Use the root `go.mod` as the Go version source

The root `go.mod` is tracked and already represents the repository-wide toolchain declaration. The checker will derive the expected Go version from it rather than `go.work`.

Alternatives considered:

- Keep `go.work`: rejected because CI may generate it from the installed toolchain.
- Add a dedicated version file: rejected because it would add a second declaration that can drift from the root module.
- Use a nested module: rejected because no nested module represents the repository-wide declaration.

### Derive checked modules from tracked files

The checker will discover tracked `go.mod` files directly instead of parsing workspace membership. This preserves validation even when no workspace file exists and avoids treating local workspace state as configuration.

Alternatives considered:

- Maintain a hard-coded module list: rejected because new tracked modules could be omitted.
- Parse `go.work` only when present: rejected because an untracked file would still affect validation.

### Test the script's observable contract

Regression coverage will execute the version check with no `go.work` and with an untracked conflicting `go.work`, asserting identical results and no file creation. Tests will also demonstrate that the installed Go patch version does not supply the expected version.

Alternatives considered:

- Test only Task dependencies: rejected because script-level tests are needed to prove the source-of-truth behavior.

## Risks / Trade-offs

- [Tracked module discovery includes an unintended fixture] -> Exclude fixture and vendor paths or constrain discovery to repository modules used by the existing check.
- [The root module no longer represents the intended toolchain] -> Document the source-of-truth choice and make drift failures identify the source and mismatched file.
- [Tests rely on host Go behavior] -> Use fixture content and controlled command inputs rather than asserting the host Go version.

## Migration Plan

1. Update the Task dependency and checker inputs.
2. Add and run regression coverage.
3. Run `task version:check` to verify existing declarations remain aligned.
4. Roll back the Task and script changes together if the new check cannot validate the existing repository layout.
