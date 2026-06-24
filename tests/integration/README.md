# Integration Tests

End-to-end tests for the ComplyBeacon evidence pipeline. Tests use [Ginkgo](https://onsi.github.io/ginkgo/) v2 with [Gomega](https://onsi.github.io/gomega/) matchers to drive the compose stack at multiple deployment layers and validate that evidence flows correctly.

## Prerequisites

- [Task](https://taskfile.dev/installation/) v3+
- [Podman](https://docs.podman.io/) and podman-compose (`pip install podman-compose`)
- Go 1.25+ (Ginkgo CLI is managed via `tool` directive in root `go.mod`)

### Installing Task

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin
```

Certificates are generated automatically if missing.

## Running Tests

```bash
# Run all layers sequentially
task integration:test

# Run a single layer
task integration:test-profile PROFILE=base
task integration:test-profile PROFILE=storage

# Top-level alias (runs all layers)
task test:integration
```

### IDE / Manual Debugging

Start the stack without running tests, then run Ginkgo directly:

```bash
# Start the stack for a specific layer
task integration:up PROFILE=base

# Run tests from repo root
go tool ginkgo run -vv --label-filter="base" ./tests/integration/

# Tear down when done
task integration:down
```

Test output is written to `.test-output/integration/`.

## Deployment Layers

| Layer       | Compose Profile | Collector Config                     | Services                                 |
|-------------|-----------------|--------------------------------------|------------------------------------------|
| Base        | *(none)*        | `configs/collector-base.yaml`        | collector, Loki                          |
| Storage     | `storage`       | `configs/collector-storage.yaml`     | collector, Loki, RustFS                  |
| Storage TLS | `storage-tls`   | `configs/collector-storage-tls.yaml` | collector-tls, Loki, RustFS (TLS)        |
| Auth        | `auth`          | `configs/collector-auth.yaml`        | collector-auth, Loki, RustFS, Dex (OIDC) |

## Test Suites

| File                  | Label         | Test Cases                                                                                                     |
|-----------------------|---------------|----------------------------------------------------------------------------------------------------------------|
| `base_test.go`        | `base`        | Healthcheck, OCSF transform to Loki, success evidence, malformed evidence resilience                           |
| `storage_test.go`     | `storage`     | S3 export, S3 partitioning by policy ID                                                                        |
| `storage_tls_test.go` | `storage-tls` | TLS S3 export, TLS S3 partitioning (via `rc` client)                                                           |
| `auth/auth_test.go`   | `auth`        | OIDC reject unauthenticated/invalid/expired/wrong-audience, accept valid token, webhook unauthenticated access |

## Adding a New Test Case

1. Create evidence fixture(s) in `fixtures/` following the OCSF format from existing fixtures
2. Add the test spec to the appropriate layer file (`base_test.go`, `storage_test.go`, `storage_tls_test.go`, or `auth/auth_test.go`)
3. Use the `Label()` decorator matching the layer so `--label-filter` selects it correctly
4. Follow the pattern: `postEvidence()` → `Eventually` poll via `queryLoki()`/`listS3Objects()` → verify pipeline health
