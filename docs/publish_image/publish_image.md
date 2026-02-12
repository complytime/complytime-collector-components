# Image Publishing Process

This guide explains how to publish in GHCR and promote in Quay for container images by using the org-infra reusable workflows.

## Process Overview TL;DR

The publishing process values **security and automation** to provide predictable, low-cost image releases.

```
Main Branch Push  →  Build + Scan + Sign  →  GHCR
                                              ↓
Release Tag (v*)  →  Verify + Promote     →  Quay.io
```

---

## Main Branch Pipeline (Scan Source → Build → Scan Image → Sign)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              MAIN BRANCH PUSH                                   │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┐
│                          org-infra reusable workflows                           │
└ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘

  ╔═══════════════════════════════════════════════════════════════════════════╗
  ║  STAGE 1: BUILD & PUSH                                                    ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐  ║
  ║  │            reusable_publish_ghcr.yml                                │  ║
  ║  │  ┌───────────────────────────────────────────────────────────────┐  │  ║
  ║  │  │  • Checkout source                                            │  │  ║
  ║  │  │  • Login to GHCR                                              │  │  ║
  ║  │  │  • Build multi-platform image (Buildx)                        │  │  ║
  ║  │  │  • Push to ghcr.io/org/image:sha-<commit>                     │  │  ║
  ║  │  │  • Auto-generate SBOM + SLSA provenance (buildx attestations) │  │  ║
  ║  │  └───────────────────────────────────────────────────────────────┘  │  ║
  ║  │                                                                     │  ║
  ║  │  OUTPUTS: digest (sha256:...), image (ghcr.io/org/image)            │  ║
  ║  └─────────────────────────────────────────────────────────────────────┘  ║
  ╚═══════════════════════════════════════════════════════════════════════════╝
                                      │
                                      │ digest, image
                                      ▼
  ╔═══════════════════════════════════════════════════════════════════════════╗
  ║  STAGE 2: VULNERABILITY SCAN                                              ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐  ║
  ║  │            reusable_vuln_scan.yml                                   │  ║
  ║  │  ┌───────────────────────────────────────────────────────────────┐  │  ║
  ║  │  │  [OSV-Scanner]     Dependency CVE scan (lockfiles)            │  │  ║
  ║  │  │  [Trivy Source]    Secrets + misconfig scan                   │  │  ║
  ║  │  │  [Trivy Image]     Container OS/runtime vuln scan             │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  • Upload SARIF results to GitHub Security tab                │  │  ║
  ║  │  │  • Attach vuln attestation to image (cosign attest)           │  │  ║
  ║  │  └───────────────────────────────────────────────────────────────┘  │  ║
  ║  │                                                                     │  ║
  ║  │  OUTPUTS: trivy_image_scan_passed, vuln_attestation_attached        │  ║
  ║  └─────────────────────────────────────────────────────────────────────┘  ║
  ╚═══════════════════════════════════════════════════════════════════════════╝
                                      │
                                      │ digest, scan results
                                      ▼
  ╔═══════════════════════════════════════════════════════════════════════════╗
  ║  STAGE 3: SIGN & VERIFY                                                   ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐  ║
  ║  │            reusable_sign_and_verify.yml                             │  ║
  ║  │  ┌───────────────────────────────────────────────────────────────┐  │  ║
  ║  │  │  • Keyless signing via Sigstore/Fulcio (OIDC)                 │  │  ║
  ║  │  │  • cosign sign image@digest                                   │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  Verify all attestations:                                     │  │  ║
  ║  │  │  ✓ Signature (identity: github.com/org/*)                     │  │  ║
  ║  │  │  ✓ SLSA Provenance                                            │  │  ║
  ║  │  │  ✓ SBOM (SPDX)                                                │  │  ║
  ║  │  │  ✓ Vulnerability scan attestation                             │  │  ║
  ║  │  └───────────────────────────────────────────────────────────────┘  │  ║
  ║  └─────────────────────────────────────────────────────────────────────┘  ║
  ╚═══════════════════════════════════════════════════════════════════════════╝
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │  ✅ Image Ready in GHCR       │
                      │  ghcr.io/org/image:sha-abc123 │
                      │  + signature                  │
                      │  + SBOM                       │
                      │  + SLSA provenance            │
                      │  + vuln attestation           │
                      └───────────────────────────────┘
```

---

## Release Pipeline (Promote to Production)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                        RELEASE TAG PUSH (v1.2.3)                                │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
  ╔═══════════════════════════════════════════════════════════════════════════╗
  ║  STAGE 4: PROMOTE TO PRODUCTION REGISTRY                                  ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐  ║
  ║  │            reusable_publish_quay.yml                                │  ║
  ║  │  ┌───────────────────────────────────────────────────────────────┐  │  ║
  ║  │  │  1. Lookup source: ghcr.io/org/image:sha-<commit> → digest    │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  2. Pre-promotion verification:                               │  │  ║
  ║  │  │     ✓ Verify source signature                                 │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  3. cosign copy (preserves all signatures + attestations)     │  │  ║
  ║  │  │     ghcr.io/org/image@sha256:... → quay.io/org/image:v1.2.3   │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  4. Apply semver tags: v1.2.3, sha-<full>, and sha-<short>    │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  5. Post-promotion verification (destination registry)        │  │  ║
  ║  │  └───────────────────────────────────────────────────────────────┘  │  ║
  ║  │                                                                     │  ║
  ║  │  OUTPUTS: digest, dest_image_full                                   │  ║
  ║  └─────────────────────────────────────────────────────────────────────┘  ║
  ╚═══════════════════════════════════════════════════════════════════════════╝
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │  ✅ Production Image Ready    │
                      │  quay.io/org/image:v1.2.3    │
                      │  (all with preserved certs)  │
                      └───────────────────────────────┘
```

---

## Publishing Images

Images are automatically built and published when changes are pushed to the `main` branch.

### Manual Trigger

To manually trigger a build (e.g., for base image updates):

1. Go to **Actions** → **Publish Images to GHCR**
2. Click **Run workflow**
3. Optionally check **Force rebuild without cache**

## Promoting to Quay.io

Promotion copies signed images from GHCR to Quay.io for public distribution.

> **Key Point:** Promotion does **not rebuild** the image. It uses `cosign copy` to transfer the exact same bytes (identical `sha256` digest) from GHCR to Quay, preserving all signatures and attestations. This guarantees the image you tested in GHCR is identical to what's released on Quay.

### Creating a Release

```bash
# Ensure your changes are merged to main and images are built
git checkout main
git pull origin main

# Create a signed tag
git tag -s v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### Release Cadence

Releases are created as needed. Maintainers coordinate releases via issues or discussions.

## Setting Up The Repository

### 1. Create Caller Workflows

- Create a workflow for publishing to GHCR as [`ci_publish_ghcr.yml`](../.github/workflows/ci_publish_ghcr.yml) 
- Create a workflow for publishing to Quay as [`ci_publish_quay.yml`](../.github/workflows/ci_publish_quay.yml)
- Get the commit SHA from the successful run of ci_publish_ghcr.yml 
- Then tag the built commit and push trigger the Quay

### 2. Configure Secrets

Add these secrets in **Settings** → **Secrets and variables** → **Actions**:

| Secret | Required For | Description |
|--------|--------------|-------------|
| `QUAY_USERNAME` | Promotion | Quay.io robot account username |
| `QUAY_PASSWORD` | Promotion | Quay.io robot account token |

> **Note:** GHCR uses `GITHUB_TOKEN` automatically, no additional secrets needed.

### 3. Enable Branch Protection

In **Settings** → **Branches** → **main**:
- Require status checks to pass
- Require branches to be up to date

## Verifying Images

After publishing, verify images are properly signed:

```bash
# Verify GHCR image
cosign verify ghcr.io/complytime/complytime-compass \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# Verify Quay image
cosign verify quay.io/continuouscompliance/complytime-compass \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

## Quick Reference

| Task | Workflow | Trigger |
|------|----------|---------|
| Build & publish to GHCR | [`ci_publish_ghcr.yml`](../.github/workflows/ci_publish_ghcr.yml) | Push to `main` |
| Promote to Quay.io | [`ci_publish_quay.yml`](../.github/workflows/ci_publish_quay.yml)| Push tag `v*.*.*` |

## More Information
- [Sigstore Documentation](https://docs.sigstore.dev/) — Keyless signing details