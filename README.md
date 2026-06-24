# ComplyBeacon

----

> 🤖 LLM WARNING 🤖
>
> This project was written with LLM (AI) assistance.
>
> 🤖 LLM WARNING 🤖

----

**ComplyBeacon** is an open-source observability toolkit that collects, normalizes, and exports compliance evidence by extending the OpenTelemetry (OTel) standard.

It bridges the gap between raw policy scanner output and modern logging pipelines, providing a unified, normalized, and auditable data stream for security and compliance analysis.

---

> **WARNING:** All components are under initial development and are **not** ready for production use.

---

## Architecture

ComplyBeacon is composed of two components that form an evidence collection and normalization pipeline:

### 1. ProofWatch

An instrumentation library that accepts raw compliance evidence from policy scanners (OPA, Gatekeeper, Conforma, Sentinel), normalizes it into OCSF/Gemara format, and emits it as an OpenTelemetry log stream with accompanying metrics.

### 2. Beacon

A custom OpenTelemetry Collector distribution that hosts the pipeline. It receives log records via OTLP, webhook, or file input, batches and transforms them into standardized OCSF format, then exports to configured backends (Loki, S3).

### Data Flow

```mermaid
graph TB
    OPA["OPA"] --> NORM
    GK["Gatekeeper"] --> NORM
    CF["Conforma"] --> NORM
    SEN["Sentinel"] --> NORM

    subgraph pw ["ProofWatch - Instrumentation Library"]
        NORM["Normalize Evidence<br/>OCSF / Gemara"] --> EMIT["Emit OTLP<br/>Logs + Metrics"]
    end

    EMIT -- "OTLP" --> OTLPRX
    EMIT -- "OTLP" --> WHRX

    subgraph beacon ["Beacon - OTel Collector Distribution"]
        OTLPRX["OTLP Receiver"] --> BATCH
        WHRX["Webhook Receiver"] --> BATCH
        FLRX["FileLog Receiver"] --> BATCH

        BATCH["Batch"] --> XFORM["Transform<br/>OCSF parse"]

        XFORM --> S2M["SignalToMetrics"]
        XFORM --> DBG["Debug"]
        XFORM --> OTLPEX["OTLP/HTTP Exporter"]
        XFORM --> S3EX["S3 Exporter"]
        S2M --> FILEEX["File Exporter"]
    end

    OTLPEX --> LOKI
    S3EX --> S3STORE

    subgraph outputs ["Downstream Systems"]
        LOKI[("Loki")] --> GRAF["Grafana"]
        S3STORE[("S3 Archive")] --> HP["Hyperproof"]
    end

    classDef sourceNode fill:#6366f1,stroke:#4338ca,color:#fff
    classDef pwNode fill:#8b5cf6,stroke:#6d28d9,color:#fff
    classDef beaconNode fill:#0ea5e9,stroke:#0284c7,color:#fff
    classDef procNode fill:#d97706,stroke:#b45309,color:#fff
    classDef outputNode fill:#64748b,stroke:#334155,color:#fff
    classDef catNode fill:#fbbf24,stroke:#b45309,color:#333

    class OPA,GK,CF,SEN sourceNode
    class NORM,EMIT pwNode
    class OTLPRX,WHRX,FLRX,BATCH,XFORM,S2M beaconNode
    class DBG,OTLPEX,S3EX,FILEEX procNode
    class LOKI,S3STORE,GRAF,HP outputNode

    style pw fill:#f5f3ff,stroke:#8b5cf6,stroke-width:2px
    style beacon fill:#f0f9ff,stroke:#0ea5e9,stroke-width:2px
    style outputs fill:#f1f5f9,stroke:#64748b,stroke-width:2px
```

## Prerequisites

- **Go 1.26+**
- **Task** ([taskfile.dev](https://taskfile.dev/installation/))
- **Podman**
- **Git**

```bash
# Install Task
brew install go-task/tap/go-task  # macOS
# or
go install github.com/go-task/task/v3/cmd/task@latest
```

## Quick Start

### 1. Build and Deploy

The demo stack runs Beacon, Loki, and Grafana via compose:

```bash
# Build the collector image (resolves deps, syncs versions, builds container)
task build

# Start all services
task infra:deploy

# Stop everything
task infra:undeploy
```

### 2. Test the Pipeline

```bash
# Send sample compliance evidence
curl -X POST http://localhost:8088/eventsource/receiver \
  -H "Content-Type: application/json" \
  -d @tests/integration/fixtures/evidence-fail.json

# View enriched logs in Grafana at http://localhost:3000
```

### 3. Evidence Storage (S3)

The demo stack includes [rustfs](https://github.com/rustfs/rustfs), an S3-compatible object store, so evidence archival works out-of-the-box with no configuration.

After `task infra:deploy`, the following are available:

| Service        | URL                     | Credentials                   |
|----------------|-------------------------|-------------------------------|
| rustfs S3 API  | <http://localhost:9000> | `rustfsadmin` / `rustfsadmin` |
| rustfs Console | <http://localhost:9001> | `rustfsadmin` / `rustfsadmin` |

Evidence files are written to the `complybeacon-evidence` bucket and can be browsed in the rustfs console.

#### Using Real AWS S3 Instead

To export evidence to a real AWS S3 bucket, override the defaults by exporting environment variables before starting the stack:

| Variable                | Default (local dev)     | Description                  |
|-------------------------|-------------------------|------------------------------|
| `AWS_REGION`            | `us-east-1`             | AWS region of your S3 bucket |
| `S3_BUCKETNAME`         | `complybeacon-evidence` | Target S3 bucket name        |
| `S3_OBJ_DIR`            | `evidence`              | Folder prefix for objects    |
| `AWS_ACCESS_KEY_ID`     | `rustfsadmin`           | AWS access key               |
| `AWS_SECRET_ACCESS_KEY` | `rustfsadmin`           | Corresponding secret key     |
| `S3_ENDPOINT`           | `http://rustfs:9000`    | S3 endpoint (unset for AWS)  |

Production deployments should use a dedicated collector configuration with `disable_ssl: false` and `s3_force_path_style: false`. See [Sync_Evidence2Hyperproof](docs/integration/Sync_Evidence2Hyperproof.md) for the full AWS integration guide.

### 4. Authentication

Beacon bundles an OTel Collector auth extension (`oidcauthextension`) but does **not** enable it by default — the shipped configs prioritize development simplicity. A `bearertokenauthextension` is also available in the extension registry for outbound authentication.

| Extension           | Direction | Purpose                                               |
|---------------------|-----------|-------------------------------------------------------|
| `oidcauthextension` | Inbound   | Validates OIDC (JWT) tokens on incoming OTLP requests |

#### Configuring OIDC on OTLP Receivers

To require OIDC authentication on incoming OTLP gRPC/HTTP requests, add the extension and reference it from the receiver:

```yaml
extensions:
  oidc:
    providers:
      - issuer_url: https://your-idp.example.com/realms/your-realm
        audience: your-client-id
        username_claim: email

receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
        auth:
          authenticator: oidc
      http:
        endpoint: 0.0.0.0:4318
        auth:
          authenticator: oidc

service:
  extensions: [oidc]
  pipelines:
    logs/analysis_pipeline:
      receivers: [otlp]
      # ... processors and exporters
```

The extension discovers the provider's signing keys via the standard `/.well-known/openid-configuration` endpoint. Requests without a valid JWT are rejected at the receiver.

> **Note:** The webhook receiver does not support OTel auth extensions. Only OTLP gRPC and HTTP receivers can be protected with OIDC.

## Development

### Task Commands

```bash
task                       # List all available tasks
task build                 # Build the collector container image
task test                  # Run tests with coverage
task test-race             # Run tests with race detection
task lint                  # Run golangci-lint
task check                 # Run all quality gates (lint + test)
task test:integration      # Run end-to-end integration tests (Ginkgo)
task deps                  # Tidy, verify, and download dependencies
task workspace             # Set up Go workspace
task clean                 # Remove build artifacts and test output
```

### Code Generation

```bash
task codegen:weaver-codegen            # Regenerate attribute constants from model/
task codegen:weaver-docsgen            # Regenerate attribute docs from model/
task codegen:weaver-check              # Validate semantic convention model
task codegen:weaver-semantic-check     # Validate logs against semantic conventions
```

### Version Management

The project prevents version drift between the Go modules and the beacon-distro collector distribution. Version checks run automatically with `task test` and block CI if versions diverge.

```bash
task version:check-otel-versions       # Check alignment
task version:sync-otel-versions        # Sync beacon-distro with proofwatch
task version:check-go-version          # Verify Go version compatibility
task version:check-go-mod-consistency  # Check OTel dependency consistency
```

### Container Image

The beacon collector image is built from `beacon-distro/Containerfile.collector`.

**Local builds:**

```bash
# Build with default tag (complybeacon/collector:latest)
task build

# Build with custom image name and tag
task build IMAGE=ghcr.io/complytime/complybeacon TAG=v1.0.0

# Run standalone (without compose)
podman run --rm \
  -v ./configs/collector-base.yaml:/etc/otel-collector.yaml:Z \
  -p 4317:4317 -p 8088:8088 \
  complybeacon/collector:latest \
  --config=/etc/otel-collector.yaml
```

**CI-built images:**

Images are automatically published to GHCR:
- **Main branch:** Merged changes → `sha-<commit>` (immutable)
- **PRs (org members):** Open/update PR → `dev-pr<number>` (mutable) + `sha-<commit>`
- **External PRs:** No images built (security)

**Verify published images** (using [skopeo](https://github.com/containers/skopeo)):

```bash
# Install skopeo
brew install skopeo  # macOS
dnf install skopeo   # Fedora/RHEL/CentOS

# Authenticate (required for private repos)
# Create a token at https://github.com/settings/tokens with 'read:packages' scope
skopeo login ghcr.io
# Username: your-github-username
# Password: paste your token (ghp_...)

# List all tags
skopeo list-tags docker://ghcr.io/complytime/complybeacon-beacon-distro

# Inspect your PR image (replace 123 with your PR number)
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123

# Get just the digest
skopeo inspect docker://ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123 --format "{{.Digest}}"

# Pull and use (also requires authentication for private images)
podman login ghcr.io  # If not already logged in
podman pull ghcr.io/complytime/complybeacon-beacon-distro:dev-pr123
```

**Troubleshooting:**
- **"authentication required" error:** Login with `skopeo login ghcr.io` (token from [github.com/settings/tokens](https://github.com/settings/tokens))
- **PR didn't build an image:** Check you're a `complytime` org member, changed relevant files, and view "Publish Images to GHCR" in Actions tab
- **More details:** See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md#verifying-published-images)

See [docs/publish_image/publish_image.md](./docs/publish_image/publish_image.md) for the complete publishing pipeline.

**Grafana Dashboard:** To configure Loki as default datasource, see [deploy/terraform/README.md](./deploy/terraform/README.md).

## Additional Resources

- **Design & Architecture:** [docs/DESIGN.md](./docs/DESIGN.md)
- **Development Guide:** [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)
- **Image Publishing:** [docs/publish_image/publish_image.md](./docs/publish_image/publish_image.md)
- **OpenTelemetry:** [opentelemetry.io](https://opentelemetry.io/docs/)
