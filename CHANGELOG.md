# Changelog

All notable changes to this project will be documented in this file.

> The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] — 2026-06-17

### Fixed

- **beacon-distro**: Restored OIDC and bearer token auth extensions to the collector manifest. These extensions were unintentionally dropped during the UBI10 migration in v0.1.0, causing the collector to reject authentication extension configs at startup.

## [0.1.0] — 2026-06-08

### Security

- Bump Go from 1.26.3 to 1.26.4 across all modules, CI workflows, and container build stages. Fixes CVE-2026-42507 (`net/textproto` error message injection) and CVE-2026-27145 (`crypto/x509` inefficient hostname parsing) in the compiled collector binary.
- **beacon-distro**: Switch runtime base from `ubi-micro` (no package manager) to `ubi-minimal` so `microdnf` can apply OS security patches at build time. Resolves Trivy HIGH/CRITICAL scan failures from unpatched packages in the distroless base image.

### Added

- **proofwatch** library for collecting compliance evidence and emitting it as OpenTelemetry logs. Supports TLS-secured collector connections. Exposes OTel metrics counters that track evidence volume per control.
- **truthbeam** OTel Collector processor that enriches evidence logs with compliance metadata from an external enrichment service. Response caching and TLS enabled by default.
- **beacon-distro** pre-built OTel Collector distribution (UBI10 Minimal container) bundling proofwatch, truthbeam, and the AWS S3 exporter. Container images are cosign-signed with SBOM attestation and published to `ghcr.io` and `quay.io`.
- Compliance attribute model defined as OTel Weaver semantic conventions, with generated Go constants and Markdown attribute reference docs. Attributes match the Gemara v1 evidence schema and OCSF activity structure.
- AWS S3 evidence export with automatic partitioning by `policy.rule.id`. Evidence for each compliance rule lands in its own prefix for straightforward audit retrieval.
- Local development stack (`task infra:deploy`) running the full evidence pipeline via podman-compose: Loki for log aggregation, the beacon collector, RustFS for S3-compatible object storage, and Grafana with a pre-provisioned evidence dashboard. Works out of the box on Linux (including SELinux/RHEL) with no cloud credentials required.
