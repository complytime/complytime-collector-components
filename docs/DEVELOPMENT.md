# ComplyBeacon Development Guide

This guide provides comprehensive instructions for setting up, building, and testing the ComplyBeacon project.
It complements the [DESIGN.md](./DESIGN.md) document by focusing on the practical aspects of development.

<!-- TOC -->
* [ComplyBeacon Development Guide](#complybeacon-development-guide)
  * [Prerequisites](#prerequisites)
    * [Required Software](#required-software)
  * [Development Environment Setup](#development-environment-setup)
    * [1. Clone the Repository](#1-clone-the-repository)
    * [2. Install Task (if needed)](#2-install-task-if-needed)
    * [3. Initialize Go Workspace](#3-initialize-go-workspace)
    * [4. Install Dependencies](#4-install-dependencies)
    * [5. Verify Installation](#5-verify-installation)
  * [Project Structure](#project-structure)
  * [Testing](#testing)
    * [Running Tests](#running-tests)
    * [Integration Testing](#integration-testing)
  * [Component Development](#component-development)
    * [1. ProofWatch Development](#1-proofwatch-development)
    * [2. Beacon Distro Development](#2-beacon-distro-development)
  * [Debugging and Troubleshooting](#debugging-and-troubleshooting)
    * [Debugging Tools](#debugging-tools)
  * [Code Generation](#code-generation)
    * [1. OpenTelemetry Semantic Conventions](#1-opentelemetry-semantic-conventions)
    * [2. Manual Code Generation](#2-manual-code-generation)
  * [Deployment and Demo](#deployment-and-demo)
    * [Local Development Demo](#local-development-demo)
  * [Additional Resources](#additional-resources)
<!-- TOC -->

## Prerequisites

### Required Software

- **Go 1.26+**: The project uses Go 1.26.4
- **Podman**: For containerized development and deployment (Docker is not supported)
- **Task**: For build automation ([installation guide](https://taskfile.dev/installation/))
- **Git**: For version control
- **openssl**: Cryptography toolkit

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/complytime/complybeacon.git
cd complybeacon
```

### 2. Install Task (if needed)

The project uses [Task](https://taskfile.dev) for build automation. Install it if you don't have it:

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Or using Go
go install github.com/go-task/task/v3/cmd/task@latest

# Verify installation
task --version
```

### 3. Initialize Go Workspace

The project uses Go workspaces to manage multiple modules:

```bash
task workspace
```

This creates a `go.work` file that includes all project modules:
- `.` (root module — tool dependencies)
- `./proofwatch`
- `./tests/integration`

### 4. Install Dependencies

Dependencies are managed per module. Install them for all modules:

```bash
task deps
```

This automatically:
- Syncs the Go workspace
- Runs `go mod tidy` for each module
- Verifies and downloads dependencies

### 5. Verify Installation

```bash
# Run tests to verify everything works
task test

# Run all quality gates (lint + test)
task check
```

## Dependency Management

### Regular Dependency Updates

Update all Go packages to their latest versions:

```bash
task dev:deps:update
```

This command automatically:
- Updates all non-OTel Go packages to the latest versions
- **Excludes OTel packages** (they are updated manually - see below)
- Runs `go mod tidy` for all modules
- Syncs the Go workspace
- Verifies all modules

### OpenTelemetry Collector Versions

**Why OTel has two version series:**

OpenTelemetry Collector publishes two version series in each release:
- **Stable API** (v1.x) — backward compatible core interfaces (`component`, `pdata`, `consumer`)
- **Experimental API** (v0.x) — may break between releases, helpers & new features (`componenttest`, `processorhelper`, `config/*`)

Each release publishes **both versions together**:
- v0.151.0 release (April 2025): `v1.57.0` (stable) + `v0.151.0` (experimental)
- v0.155.0 release (June 2025): `v1.61.0` (stable) + `v0.155.0` (experimental)

**How we handle it:**

- **OTel versions are constrained by contrib package availability** — we pin to versions where all contrib packages exist
- `task dev:deps:update` **excludes OTel packages** — they are updated manually after verifying contrib compatibility
- `task version:sync` propagates pinned OTel versions to all modules, Containerfiles, and CI configs

**Current version:** v1.61.0 (stable) + v0.155.0 (experimental)

**Why the constraint?** Contrib packages (used in `beacon-distro/manifest.yaml`) release 1-2 versions behind the main collector packages. Blindly upgrading to the latest OTel version causes build failures when contrib packages don't exist yet.

### How to Upgrade OpenTelemetry Collector

**Important:** OTel packages are **not** upgraded automatically by `task dev:deps:update`. Follow this process:

**Step 1: Check latest contrib release**

Visit the [contrib releases page](https://github.com/open-telemetry/opentelemetry-collector-contrib/releases) and note the latest version (e.g., `v0.155.0`).

**Step 2: Identify required OTel version**

Check a contrib package's `go.mod` to see what OTel version it requires:

```bash
go mod download -json github.com/open-telemetry/opentelemetry-collector-contrib/connector/signaltometricsconnector@v0.155.0 | \
  jq -r '.GoMod' | xargs cat | grep 'go.opentelemetry.io/collector/component'
```

This will show something like `go.opentelemetry.io/collector/component v1.61.0` — that's your target stable version.

**Step 3: Update proofwatch (if needed)**

```bash
cd proofwatch
# ProofWatch uses stable OTel packages only
go get go.opentelemetry.io/otel/log@v0.20.0
go mod tidy
cd ..
```

**Step 4: Propagate across the project**

```bash
task version:sync
```

This syncs the versions to all workspace modules, Containerfiles, and CI configs.

**Step 5: Verify**

```bash
task test
task integration:test
```

### Troubleshooting Version Conflicts

**Error: `unknown revision v0.XXX.0`**

The OTel version doesn't exist yet. Check the [releases page](https://github.com/open-telemetry/opentelemetry-collector/releases) for the latest available version.

**Error: dependency conflicts after `go get -u`**

You've mixed stable/experimental versions from different releases. Reset to the pinned versions:

```bash
cd proofwatch
# Downgrade to current pinned versions
go get go.opentelemetry.io/otel/log@v1.57.0
go mod tidy
cd ..
task version:sync
```

## Project Structure

```
complybeacon/
├── compose.yaml                # Container orchestration configuration
├── Taskfile.yml                # Build automation
├── .taskfiles/                 # Task modules and helper scripts
├── docs/                       # Documentation
│   ├── DESIGN.md              # Architecture and design documentation
│   ├── DEVELOPMENT.md         # This file
│   └── attributes/            # Attribute documentation
├── model/                      # OpenTelemetry semantic conventions
│   ├── attributes.yaml        # Attribute definitions
│   └── entities.yaml          # Entity definitions
├── proofwatch/                 # ProofWatch instrumentation library
│   ├── attributes.go          # Attribute definitions
│   ├── evidence.go            # Evidence types
│   └── proofwatch.go          # Main library
├── beacon-distro/              # OpenTelemetry Collector distribution
│   ├── config.yaml            # Collector configuration
│   └── Containerfile.collector # Container definition
├── configs/                    # Deployment configs (collector, Loki)
│   ├── collector-auth.yaml    # Auth layer: OIDC-secured OTLP receivers
│   ├── collector-base.yaml    # Base layer: OCSF transform + Loki
│   ├── collector-storage.yaml # Storage layer: adds S3 export
│   ├── collector-storage-tls.yaml # Storage TLS layer: S3 with TLS
│   └── loki.yaml              # Loki configuration
├── certs/                      # TLS certificate generation
├── deploy/                     # Deployment infrastructure (Terraform)
├── tests/                      # Test infrastructure
│   └── integration/           # E2E Ginkgo tests and fixtures
└── bin/                        # Built binaries (created by task infra:deploy)
```

## Testing

### Running Tests

```bash
# Run all tests (includes version checks and coverage)
task test

# Run tests with race detection
task test-race

# Generate coverage reports
task dev:coverage-report

# Run tests for specific module
cd proofwatch && go test -v ./...
```

### Integration Testing

The project includes automated integration tests using [Ginkgo](https://onsi.github.io/ginkgo/) that validate the evidence pipeline at four deployment layers:

| Layer       | Profile       | What it tests                     |
|-------------|---------------|-----------------------------------|
| Base        | *(default)*   | OCSF transform + Loki export      |
| Storage     | `storage`     | S3 evidence export + partitioning |
| Storage-TLS | `storage-tls` | TLS-secured S3 export             |
| Auth        | `auth`        | OIDC-secured OTLP receivers       |

**Prerequisites:**
- Podman and podman-compose
- Go 1.26+ (Ginkgo CLI is managed via `tool` directive in root `go.mod`)

**Run all layers:**

```bash
task integration:test
```

**Run a single layer:**

```bash
task integration:test-profile PROFILE=base
task integration:test-profile PROFILE=storage
task integration:test-profile PROFILE=storage-tls
task integration:test-profile PROFILE=auth
```

Each run builds the collector image, starts the appropriate services, runs the matching Ginkgo test suite (filtered by label), and tears down. Certificates are generated automatically if missing. Test output is written to `.test-output/integration/`.

For details on test cases and fixtures, see [tests/integration/README.md](../tests/integration/README.md).

## Component Development

### 1. ProofWatch Development

ProofWatch is an instrumentation library for emitting compliance evidence.

**Key Files:**
- `proofwatch/proofwatch.go` - Main library interface
- `proofwatch/evidence.go` - Evidence type definition
- `proofwatch/attributes.go` - OpenTelemetry attributes

**Development Workflow:**
```bash
cd proofwatch

# Run tests
go test -v ./...

# Run linting (from root)
cd ..
task lint

# Format code
go fmt ./...
```

### 2. Beacon Distro Development

The Beacon distribution is a custom OpenTelemetry Collector.

**Key Files:**
- `beacon-distro/config.yaml` - Collector configuration
- `beacon-distro/Containerfile.collector` - Container definition
- `beacon-distro/manifest.yaml` - Collector builder configuration

**Development Workflow:**

**Local builds (quick iteration):**
```bash
# Build the collector image locally
podman build -f beacon-distro/Containerfile.collector -t complybeacon-collector beacon-distro/

# Or force rebuild without cache
podman build --no-cache -f beacon-distro/Containerfile.collector -t complybeacon-collector beacon-distro/

# Run locally for quick testing
podman run --rm complybeacon-collector --config /etc/otelcol-beacon/config.yaml

# Full stack deployment for integration testing
task infra:deploy
```

**CI builds (automated image publishing):**

When you modify Containerfiles or source code and open a PR, the CI automatically builds and publishes dev images (if you're an org member):

```bash
# 1. Make changes to beacon-distro or proofwatch
vim beacon-distro/Containerfile.collector

# 2. Commit and push to your branch
git add .
git commit -s -m "feat(beacon-distro): update base image to UBI10"
git push origin your-branch

# 3. Open a PR to main
# The workflow will automatically:
#   - Verify you're an org member
#   - Build the image
#   - Scan for vulnerabilities
#   - Sign the image
#   - Run integration tests
#   - Publish to ghcr.io/complytime/complybeacon-beacon-distro:dev-pr<number>

# 4. Verify your image was published
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123

# 5. Use the dev image in testing
podman pull ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
# Or in compose.yaml:
# image: ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
```

**When images are built:**
- ✅ Push to `main` branch (production, tagged `sha-<commit>`)
- ✅ PRs from org members (dev, tagged `dev-pr<number>` + `sha-<commit>`)
- ❌ PRs from external contributors (no image built for security)

See [docs/publish_image/publish_image.md](./publish_image/publish_image.md) for complete details on the image publishing pipeline.

## Debugging and Troubleshooting

### Debugging Tools

```bash
# View all container logs
podman-compose -f compose.yaml logs -f

# View specific service logs
podman-compose -f compose.yaml ps            # List running services
podman-compose -f compose.yaml logs -f collector

# Check container status
podman images | grep complybeacon            # List built images
podman inspect complybeacon-collector        # Inspect image details
```

## Verifying Published Images

When you open a PR or merge to `main`, the CI pipeline automatically builds and publishes container images to GitHub Container Registry (GHCR). Use `skopeo` to verify your images without pulling them.

### Install Skopeo

```bash
# macOS
brew install skopeo

# Fedora/RHEL/CentOS
dnf install skopeo

# Ubuntu/Debian
apt-get install skopeo
```

### Quick Checks

```bash
# List all available tags
skopeo list-tags docker://ghcr.io/complytime/complybeacon-beacon-distro

# Check if your PR image exists (replace 123 with your PR number)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123

# Check if a main branch image exists (replace abc123 with commit SHA)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:sha-abc123

# Get just the digest
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Digest}}"

# Get image creation timestamp
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 \
  --format "{{.Created}}"
```

### Authentication

If the repository is private, authenticate first:

```bash
# Create a GitHub Personal Access Token with 'read:packages' scope at:
# https://github.com/settings/tokens

# Then login
skopeo login ghcr.io
# Username: your-github-username
# Password: paste your token (ghp_...)

# Or use environment variable
echo $GITHUB_TOKEN | skopeo login ghcr.io -u your-github-username --password-stdin
```

### Image Tagging Strategy

| Build Type      | Trigger            | Tags                             | Notes                                 |
|-----------------|--------------------|----------------------------------|---------------------------------------|
| **Production**  | Merge to `main`    | `sha-<commit>`                   | Immutable, builds after CI passes     |
| **Dev**         | PR from org member | `dev-pr<number>`, `sha-<commit>` | `dev-pr` is mutable (updates on push) |
| **External PR** | PR from non-member | None                             | No images built (security)            |

### Common Scenarios

**Verify your PR image was published:**

```bash
# Find your PR number (visible in PR title, e.g., #123)
# Check the Actions tab for the "Publish Images to GHCR" workflow

# List all tags to confirm
skopeo list-tags docker://ghcr.io/complytime/complybeacon-beacon-distro | grep "dev-pr123"

# Inspect the image
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
```

**Use dev image in local testing:**

```bash
# Pull the image
podman pull ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123

# Or reference it directly in compose.yaml
# Edit compose.yaml:
# services:
#   collector:
#     image: ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
```

**Compare dev and production images:**

```bash
# Both should have the same digest if built from the same commit
DEV_DIGEST=$(skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 --format "{{.Digest}}")
SHA_DIGEST=$(skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:sha-abc123 --format "{{.Digest}}")

echo "Dev digest:  $DEV_DIGEST"
echo "SHA digest:  $SHA_DIGEST"

if [ "$DEV_DIGEST" = "$SHA_DIGEST" ]; then
  echo "✅ Images are identical"
else
  echo "❌ Images differ (expected if commits are different)"
fi
```

**Troubleshooting image builds:**

If your PR doesn't produce an image:

1. **Check org membership:** Only `complytime` org members' PRs build images
   ```bash
   # Verify you're listed as a member
   curl -s https://api.github.com/orgs/complytime/members | jq -r '.[].login' | grep your-username
   ```

2. **Check if files changed:** Image builds only trigger when:
   - Any `Containerfile*` changes
   - Source code in `beacon-distro/`, `proofwatch/` changes
   - The workflow file (`.github/workflows/ci_publish_ghcr.yml`) changes

3. **Check workflow run:** Go to **Actions** → **Publish Images to GHCR** and check for errors

4. **Check CI status:** The workflow waits for CI to complete on PRs to main

For complete details on the image publishing pipeline, see [docs/publish_image/publish_image.md](./publish_image/publish_image.md). For quick skopeo examples, see the [Container Image section in the README](../README.md#container-image).

---

## Code Generation

The project uses several code generation tools:

### 1. OpenTelemetry Semantic Conventions

Generate documentation and Go code from semantic convention models:

```bash
# Generate documentation
task codegen:weaver-docsgen

# Generate Go code
task codegen:weaver-codegen

# Validate models
task codegen:weaver-check

# Validate logs against semantic conventions
task codegen:weaver-semantic-check
```

### 2. Manual Code Generation

If you modify the semantic conventions:

```bash
# Update semantic conventions
vim model/attributes.yaml
vim model/entities.yaml

# Regenerate all code (weaver only)
task codegen:weaver-codegen
```

## Deployment and Demo

### Local Development Demo

The demo environment orchestrates multiple containers (Grafana, Loki, Beacon Collector).

1. **Generate self-signed certificate**

For TLS testing, first generate self-signed certificates:

```bash
task infra:generate-self-signed-cert
```

2. **Start the full stack:**
```bash
# Interactive mode (shows logs in terminal)
task deploy

# Or background/detached mode
podman-compose -f compose.yaml up -d
```

This automatically:
- Syncs OTel versions from proofwatch and manifest.yaml
- Builds the beacon collector image
- Starts all services (Grafana, Loki, Collector)

3. **Test the pipeline:**
```bash
curl -X POST http://localhost:8088/eventsource/receiver \
  -H "Content-Type: application/json" \
  -d @tests/integration/fixtures/evidence-fail.json
```

4. **View results:**
- Grafana: <http://localhost:3000>
- View logs: `podman-compose -f compose.yaml logs -f`

5. **Stop the stack:**
```bash
task infra:undeploy
```

---

## Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Go Documentation](https://golang.org/doc/)
- [Podman Documentation](https://docs.podman.io/)
- [Project Design Document](./DESIGN.md)
- [Attribute Documentation](./attributes/)
- [Containers Guide](https://github.com/complytime/community/blob/main/CONTAINERS_GUIDE.md)

For questions or support, please open an issue in the GitHub repository.
