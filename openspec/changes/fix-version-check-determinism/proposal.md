# Proposal

## Why

`task version:check` currently creates a workspace when `go.work` is absent and then treats that generated file as its Go-version source. Because `go work init` records the runner's installed Go patch version, unrelated dependency updates can fail CI version validation after a runner update.

The version check must remain read-only and derive expectations from tracked repository declarations so results are reproducible across developer and CI environments.

## What Changes

- Remove the workspace-bootstrap dependency from `version:check`.
- Define a tracked Go-version declaration as the version check's source of truth.
- Preserve workspace creation for build and dependency operations that require it.
- Update version-validation documentation to identify the tracked source and read-only behavior.
- Add regression coverage for absent or stale untracked `go.work` files and differing installed Go patch versions.

## Capabilities

### New Capabilities
- `version-validation`: Deterministic, read-only validation of Go version alignment across tracked repository consumers.

### Modified Capabilities

None.

## Impact

- `.taskfiles/version.yml` and `.taskfiles/scripts/version-check.sh`
- Version-check test coverage and developer documentation
- Go module declarations and `beacon-distro/Containerfile.collector` remain validation targets
- CI jobs that invoke `task version:check`
