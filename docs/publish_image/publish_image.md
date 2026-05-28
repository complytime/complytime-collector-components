# Image Publishing Process

This guide explains how to publish in GHCR and promote in Quay for container images by using the org-infra reusable workflows.

## Process Overview TL;DR

The publishing process values **security and automation** to provide predictable, low-cost image releases.

```
Main Branch Push  →  Build + Scan + Sign  →  GHCR (sha-<commit>)
                                              ↓
Release Tag (v*)  →  Verify + Promote     →  Quay.io (v1.2.3)

PR from Org Member  →  Build + Scan + Sign  →  GHCR (dev-pr<number>)
```

**Tagging Strategy:**
- **Production builds** (main branch): `sha-<commit>` (immutable)
- **Dev builds** (org member PRs): `dev-pr<number>` + `sha-<commit>` (mutable + immutable)
- **External PRs**: No images built (security)

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

## PR Dev Build Pipeline (Org Members)

For pull requests from organization members, the same security pipeline runs to validate Containerfile changes before merge.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                       PULL REQUEST (from org member)                            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
  ╔═══════════════════════════════════════════════════════════════════════════╗
  ║  STAGE 0: ORG MEMBERSHIP CHECK                                            ║
  ║  ┌─────────────────────────────────────────────────────────────────────┐  ║
  ║  │  • Verify PR author is a complytime org member                      │  ║
  ║  │  • External contributors → skip image build (security)              │  ║
  ║  │  • Org members → proceed to build                                   │  ║
  ║  │  • Compute tag: dev-pr<number>                                      │  ║
  ║  └─────────────────────────────────────────────────────────────────────┘  ║
  ╚═══════════════════════════════════════════════════════════════════════════╝
                                      │
                                      ▼
                      (same pipeline as main branch)
                      scan-source → build → scan-image → sign → test
                                      │
                                      ▼
                      ┌───────────────────────────────┐
                      │  ✅ Dev Image Ready in GHCR   │
                      │  Primary: dev-pr123           │
                      │  Immutable: sha-abc123        │
                      │  + signature + SBOM           │
                      │  + provenance + vuln report   │
                      └───────────────────────────────┘
```

**Security Model:**
- Only **organization members** can trigger image builds on PRs
- External contributors' PRs **do not build images** (prevents resource abuse and supply chain attacks)
- Same security scanning and signing as production builds
- Images published to GHCR with `dev-pr<number>` tag (mutable) and `sha-<commit>` tag (immutable)

**Use Case:**
- Validate Containerfile changes before merge
- Test integration with dev images in local or staging environments
- Share dev images with team members for review

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
  ║  │  │     ghcr.io/org/image → quay.io/complytime/beacon-collector    │  │  ║
  ║  │  │                                                               │  │  ║
  ║  │  │  4. Apply tags: v1.2.3, latest, sha-<full>, sha-<short>      │  │  ║
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
                      │  quay.io/complytime/         │
                      │    beacon-collector:v1.2.3   │
                      │  (all with preserved certs)  │
                      └───────────────────────────────┘
```

---

## Publishing Images

Images are automatically built and published in two scenarios:

### 1. Production Builds (Main Branch)

When changes are merged to `main`, the workflow:
- Waits for CI to complete successfully
- Checks if Containerfiles changed
- If yes: builds, scans, signs, and publishes to GHCR with tag `sha-<commit>`
- Runs integration tests against the published image

**Trigger:** Push to `main` branch after CI passes

### 2. Dev Builds (Pull Requests)

When an **org member** opens or updates a PR that touches:
- Any `Containerfile*`
- Source code in `beacon-distro/`, `proofwatch/`, `truthbeam/`
- The workflow file itself

The workflow:
- Verifies PR author is a `complytime` org member
- If yes: builds, scans, signs, and publishes to GHCR with tags:
  - `dev-pr<number>` (mutable, updates on each push)
  - `sha-<commit>` (immutable)
- Runs integration tests against the published image

**Security:** External contributors' PRs **do not** trigger image builds.

**Example:**
```bash
# PR #123 opened by org member
# First push (commit abc123) creates:
#   - dev-pr123 → sha256:abc...
#   - sha-abc123 → sha256:abc...
#
# Second push (commit def456) updates:
#   - dev-pr123 → sha256:def...  (mutable, now points to new image)
#   - sha-def456 → sha256:def...  (new immutable tag)
#   - sha-abc123 → sha256:abc...  (old immutable tag still exists)
```

### Manual Trigger

To manually trigger a build (e.g., for base image updates or testing):

1. Go to **Actions** → **Publish Images to GHCR**
2. Click **Run workflow**
3. Select branch (usually `main`)
4. Optionally check **Force rebuild without cache**
5. Click **Run workflow**

The manual trigger uses the same security pipeline (scan, sign, test) as automatic builds.

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

| Secret          | Required For | Description                    |
|-----------------|--------------|--------------------------------|
| `QUAY_USERNAME` | Promotion    | Quay.io robot account username |
| `QUAY_PASSWORD` | Promotion    | Quay.io robot account token    |

> **Note:** GHCR uses `GITHUB_TOKEN` automatically, no additional secrets needed.

### 3. Enable Branch Protection

In **Settings** → **Branches** → **main**:
- Require status checks to pass
- Require branches to be up to date

## Verifying Images

### Using Skopeo (Recommended for Quick Checks)

[Skopeo](https://github.com/containers/skopeo) is a command-line tool for inspecting and copying container images **without pulling them**. This is faster and more efficient than pulling images with `podman` or `docker` when you just want to verify publication.

**Install Skopeo:**

```bash
# macOS
brew install skopeo

# Fedora/RHEL/CentOS
dnf install skopeo

# Ubuntu/Debian
apt-get install skopeo

# Or use podman alias (skopeo is bundled with podman)
alias skopeo='podman run --rm quay.io/skopeo/stable'
```

**List all tags:**

```bash
# List all tags for beacon-distro
skopeo list-tags docker://ghcr.io/complytime/complybeacon-beacon-distro
```

**Inspect a specific image:**

```bash
# Inspect production image (main branch)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:sha-0da6ac5

# Inspect dev image (PR build)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123

# Get just the digest
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Digest}}"

# Get creation timestamp
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Created}}"
```

**Authentication for private images:**

If the repository is private, authenticate first:

```bash
# Login with GitHub Personal Access Token (requires 'read:packages' scope)
skopeo login ghcr.io
# Username: your-github-username
# Password: ghp_yourPersonalAccessToken

# Or use environment variable
echo $GITHUB_TOKEN | skopeo login ghcr.io -u your-github-username --password-stdin
```

**Copy image between registries:**

```bash
# Copy dev image to local testing registry
skopeo copy \
  docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  docker://localhost:5000/beacon-distro:test
```

### Verifying Signatures (Cosign)

After confirming the image exists with `skopeo`, verify cryptographic signatures:

```bash
# Verify GHCR image (beacon-distro)
cosign verify ghcr.io/complytime/complybeacon-beacon-distro:sha-0da6ac5 \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# Verify dev image
cosign verify ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com

# Verify Quay image (production release)
cosign verify quay.io/complytime/beacon-collector:v1.2.3 \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

### Common Verification Scenarios

**Scenario 1: Check if my PR's image was published**

```bash
# Find your PR number (e.g., PR #123)
# Then check for the dev tag
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
```

**Scenario 2: Verify main branch image exists for a commit**

```bash
# Get the short commit SHA (e.g., 0da6ac5)
git rev-parse --short HEAD

# Inspect the image
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:sha-0da6ac5
```

**Scenario 3: Check when an image was last updated**

```bash
# Dev images are mutable and update on each PR push
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Created}}"
```

**Scenario 4: Verify dev and production images are identical**

```bash
# Compare digests (should match if built from same commit)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Digest}}"
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:sha-abc123 \
  --format "{{.Digest}}"
```

**Troubleshooting:**

| Error                                        | Cause                                    | Solution                                     |
|----------------------------------------------|------------------------------------------|----------------------------------------------|
| `manifest unknown`                           | Image doesn't exist                      | Check tag name, verify workflow succeeded    |
| `unauthorized`                               | Authentication required                  | Run `skopeo login ghcr.io` with GitHub token |
| `requested access to the resource is denied` | Not an org member or missing permissions | Verify org membership, check token scopes    |
| `connection refused`                         | Network/registry issue                   | Check network, retry after a moment          |



## Quick Reference

| Task                                 | Workflow                                                          | Trigger                   | Tags                             |
|--------------------------------------|-------------------------------------------------------------------|---------------------------|----------------------------------|
| Build & publish to GHCR (production) | [`ci_publish_ghcr.yml`](../.github/workflows/ci_publish_ghcr.yml) | Push to `main` (after CI) | `sha-<commit>`                   |
| Build & publish to GHCR (dev)        | [`ci_publish_ghcr.yml`](../.github/workflows/ci_publish_ghcr.yml) | PR from org member        | `dev-pr<number>`, `sha-<commit>` |
| Promote to Quay.io                   | [`ci_publish_quay.yml`](../.github/workflows/ci_publish_quay.yml) | Push tag `v*.*.*`         | `v1.2.3`, `latest`, `sha-<commit>` |

**Verification Commands:**

```bash
# List all tags
skopeo list-tags docker://ghcr.io/complytime/complybeacon-beacon-distro

# Inspect specific image
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:TAG

# Verify signature
cosign verify ghcr.io/complytime/complybeacon-beacon-distro:TAG \
  --certificate-identity-regexp='https://github.com/complytime/.*' \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com
```

## More Information

- [README Container Image Section](../../README.md#container-image) — Quick skopeo examples
- [Sigstore Documentation](https://docs.sigstore.dev/) — Keyless signing details