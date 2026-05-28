# Changelog

All notable changes to this project will be documented in this file.

> The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **proofwatch** library for collecting compliance evidence and emitting it as OpenTelemetry logs. Supports TLS-secured collector connections. Exposes OTel metrics counters that track evidence volume per control.
- **truthbeam** OTel Collector processor that enriches evidence logs with compliance metadata from an external enrichment service. Response caching and TLS enabled by default.
- **beacon-distro** pre-built OTel Collector distribution (UBI10 Micro container) bundling proofwatch, truthbeam, the AWS S3 exporter, OIDC and bearer token auth extensions, and OpenAPI request validation. Container images are cosign-signed with SBOM attestation and published to `ghcr.io` and `quay.io`.
- Compliance attribute model defined as OTel Weaver semantic conventions, with generated Go constants and Markdown attribute reference docs. Attributes match the Gemara v1 evidence schema and OCSF activity structure.
- AWS S3 evidence export with automatic partitioning by `policy.rule.id`. Evidence for each compliance rule lands in its own prefix for straightforward audit retrieval.
- Local development stack (`task infra:deploy`) running the full evidence pipeline via podman-compose: Loki for log aggregation, the beacon collector, RustFS for S3-compatible object storage, and Grafana with a pre-provisioned evidence dashboard. Works out of the box on Linux (including SELinux/RHEL) with no cloud credentials required.
