# Spec Delta

## Purpose

Defines deterministic, read-only validation of Go versions across the repository's tracked version consumers.

## ADDED Requirements

### Requirement: Version check uses a tracked Go-version source
The version-validation command SHALL derive its expected Go version from a tracked repository file and SHALL compare that value with each tracked Go module and supported Go container image.

#### Scenario: Aligned tracked declarations
- **WHEN** every tracked Go version consumer matches the tracked source of truth
- **THEN** the version-validation command succeeds

#### Scenario: Drifted tracked declaration
- **WHEN** a tracked Go module or supported Go container image differs from the tracked source of truth
- **THEN** the version-validation command fails and identifies the drifted file

### Requirement: Version check is independent of workspace state
The version-validation command SHALL not create, modify, or use an untracked workspace file when determining the expected Go version or the set of checked version consumers.

#### Scenario: No workspace file exists
- **WHEN** the repository has no workspace file
- **THEN** the version-validation command completes without creating one

#### Scenario: Untracked workspace file has a different version
- **WHEN** an untracked workspace file declares a Go version different from the tracked source of truth
- **THEN** the version-validation result is identical to the result without that workspace file

### Requirement: Version check is reproducible across Go installations
The version-validation command SHALL produce the same result for unchanged tracked inputs regardless of the installed Go patch version.

#### Scenario: Runner Go patch version differs
- **WHEN** the installed Go patch version differs from the tracked source of truth
- **THEN** the version-validation result is determined only by tracked declarations
