# Tasks

## 1. Deterministic Version Inputs

- [x] 1.1 Remove the `:dev:workspace` dependency from `version:check` and verify `task version:check` does not create or modify `go.work`.
- [x] 1.2 Update `.taskfiles/scripts/version-check.sh` to read the expected Go version from the tracked root `go.mod` and discover checked module inputs without `go.work`; verify aligned tracked declarations pass and a deliberately drifted tracked declaration is reported.
- [x] 1.3 Preserve the existing OTel validation behavior after removing workspace-derived module inventory; verify `task version:check` passes on the current repository.

## 2. Regression Coverage and Documentation

- [x] 2.1 Add regression coverage that runs version validation without `go.work` and with an untracked conflicting `go.work`; verify identical outcomes and no generated workspace file.
- [x] 2.2 Add regression coverage proving the tracked root declaration, not the installed Go patch version, determines the expected version; verify the test passes under controlled inputs.
- [x] 2.3 Update developer-facing version-validation documentation to state the tracked source of truth and read-only behavior; verify all documented commands and paths match the implementation.

## 3. Integration Verification

- [x] 3.1 Run `task version:check` from a clean working tree and verify it succeeds without workspace mutation.
- [x] 3.2 Run the added regression coverage and `task check`; verify both complete successfully.

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
