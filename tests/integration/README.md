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
task test:integration

# Run a single layer
task test:integration PROFILE=base
task test:integration PROFILE=storage
task test:integration PROFILE=enrichment
task test:integration PROFILE=compliance  # pulls OCI policy bundle via complyctl
```

### IDE / Manual Debugging

Start the stack without running tests, then run Ginkgo directly:

```bash
# Start the stack for a specific layer
task test:integration:up PROFILE=base

# Run tests from repo root
go tool ginkgo run -vv --label-filter="base" ./tests/integration/

# Tear down when done
task test:integration:down
```

Test output is written to `.test-output/integration/`.

## Deployment Layers

| Layer       | Compose Profile | Collector Config                     | Services                              |
|-------------|-----------------|--------------------------------------|---------------------------------------|
| Base        | *(none)*        | `configs/collector-base.yaml`        | collector, Loki                       |
| Storage     | `storage`       | `configs/collector-storage.yaml`     | collector, Loki, RustFS               |
| Storage TLS | `storage-tls`   | `configs/collector-storage-tls.yaml` | collector-tls, Loki, RustFS (TLS)     |
| Enrichment  | `enrichment`    | `configs/collector-enrichment.yaml`  | collector, Loki, RustFS, mock Compass |
| Compliance  | `compliance`    | `configs/collector-enrichment.yaml`  | collector, Loki, RustFS, mock Compass |

The **Compliance** layer reuses the enrichment stack and additionally requires OCI policy bundles pulled from Quay.io via `task test:compliance:pull` (runs automatically as a dependency). Tests skip gracefully if the bundle is unavailable.

## Test Suites

| File                              | Label        | Test Cases                                                                           |
|-----------------------------------|--------------|--------------------------------------------------------------------------------------|
| `base_test.go`                    | `base`       | Healthcheck, OCSF transform to Loki, success evidence, malformed evidence resilience |
| `storage_test.go`                 | `storage`    | S3 export, S3 partitioning by policy ID                                              |
| `storage_tls_test.go`             | `storage-tls`| TLS S3 export, TLS S3 partitioning (via `rc` client)                                 |
| `enrichment_test.go`              | `enrichment` | Enrichment applied, unknown policy graceful handling                                 |
| `compliance/compliance_test.go`   | `compliance` | Upstream policy parsing, per-control Loki + S3 pipeline, enrichment validation       |

## Mock Compass

The `mock-compass/` directory contains a lightweight Go HTTP server that simulates the Compass enrichment API. It:

1. Loads `fixtures/compass-responses.json` at startup
2. Serves `POST /v1/enrich` — looks up `policyRuleId` in fixtures, returns matching response or `Unmapped`
3. Serves `GET /healthz` — returns 200

### Adding a New Policy Response

Add an entry to `fixtures/compass-responses.json` keyed by the policy rule ID:

```json
{
  "my_new_policy": {
    "compliance": {
      "control": {
        "id": "MY-CTRL-01",
        "catalogId": "MY-CATALOG",
        "category": "My Category"
      },
      "enrichmentStatus": "Success",
      "frameworks": {
        "frameworks": ["My Framework v1"],
        "requirements": ["MY-CTRL-01"]
      }
    }
  }
}
```

Then create a matching evidence fixture in `fixtures/` with `policy.uid` set to `my_new_policy`.

## Adding a New Test Case

1. Create evidence fixture(s) in `fixtures/` following the OCSF format from existing fixtures
2. If the test needs a new Compass response, add it to `compass-responses.json`
3. Add the test spec to the appropriate layer file (`base_test.go`, `storage_test.go`, `storage_tls_test.go`, `enrichment_test.go`, or `compliance/compliance_test.go`)
4. Use the `Label()` decorator matching the layer so `--label-filter` selects it correctly
5. Follow the pattern: `postEvidence()` → `Eventually` poll via `queryLoki()`/`listS3Objects()` → verify pipeline health
