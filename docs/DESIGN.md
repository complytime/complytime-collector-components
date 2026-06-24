# ComplyBeacon Design Documentation

## Key Features

- **OpenTelemetry Native**: Built on the OpenTelemetry standard for seamless integration with existing observability pipelines.
- **Composability**: Components are designed as a toolkit; they are not required to be used together, and users can compose their own pipelines.
- **Compliance-as-Code**: Leverages the `gemara` model for a robust, auditable, and automated approach to risk assessment.

## Architecture Overview

### Design Principles

* **Modularity:** The system is composed of small, focused, and interchangeable services.

* **Standardization:** The architecture is built on OpenTelemetry to ensure broad compatibility and interoperability.

* **Operational Experience:** The toolkit is built for easy deployment, configuration, and maintenance using familiar cloud-native practices and protocols.


### Data Flow

The ComplyBeacon architecture processes compliance evidence through a collect, normalize, and export pipeline. The primary data flow begins with a source that generates OpenTelemetry-compliant logs.

1. **Log Ingestion**: A source generates compliance evidence and sends it as a structured log record to the `Beacon` collector, typically using `ProofWatch` to handle the emission. This can also be done by an OpenTelemetry collector agent.
2. **OCSF Transform**: The log record is received by the `Beacon` collector and transformed into OCSF (Open Cybersecurity Schema Framework) format for standardized compliance event representation.
3. **Export**: The normalized log record is exported from the `Beacon` collector to a final destination (e.g., Loki, S3, a SIEM, or data lake) for analysis and correlation.

```
┌──────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                              │
│                                                                                              │
│                                          ┌─────────────────────────┐                         │
│                                          │                         │                         │
│                                          │ Beacon Collector Distro │                         │
│   ┌────────────────────┐   ┌──────────┐ │                         │                         │
│   │                    │   │          │ ├─────────────────────────┤                         │
│   │                    ├───┤ProofWatch├─┼────┐                    │                         │
│   │                    │   │          │ │    │                    │                         │
│   │    Policy Log      │   └──────────┘ │   ┌┴─────────────────┐  │                         │
│   │    Source App      │                │   │                  │  │                         │
│   │                    │                │   │      OTLP        │  │                         │
│   │                    │                │   │    Receiver      │  │                         │
│   │                    │  ┌─────────────┼───┤                  │  │                         │
│   └────────────────────┘  │             │   └────────┬─────────┘  │                         │
│                           │             │            │            │                         │
│                           │             │   ┌────────┴─────────┐  │                         │
│                           │             │   │                  │  │                         │
│                           │             │   │   Transform /    │  │                         │
│   ┌───────────────────────┴───┐         │   │   OCSF Parse     │  │                         │
│   │                           │         │   │                  │  │                         │
│   │                           │         │   └────────┬─────────┘  │                         │
│   │      OpenTelemetry        │         │            │            │                         │
│   │      Collector Agent      │         │   ┌────────┴─────────┐  │                         │
│   │                           │         │   │    Exporter      │  │                         │
│   │                           │         │   │   (e.g. Loki,    │  │                         │
│   │                           │         │   │   Splunk, AWSS3) │  │                         │
│   │                           │         │   └──────────────────┘  │                         │
│   │                           │         └─────────────────────────┘                         │
│   └───────────────────────────┘                                                             │
│                                                                                              │
└──────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Deployment Patterns

ComplyBeacon is designed to be a flexible toolkit. Its components can be used in different combinations to fit a variety of operational needs.

* **Basic Pipeline**: The most common use case where `ProofWatch` emits events to the `Beacon` collector, which transforms them to OCSF format and exports to Loki for log aggregation and querying.
* **Storage Pipeline**: Extends the basic pipeline with S3 export for long-term evidence archival, partitioned by policy rule ID for audit retrieval.
* **Auth Pipeline**: Adds OIDC authentication via Dex to the OTLP receivers, requiring valid JWT tokens for evidence ingestion.

## Component Analysis

### 1. ProofWatch

**Purpose**: An instrumentation library for collecting and emitting compliance evidence as OpenTelemetry log streams. It provides a standardized interface for tracking policy evaluation events and compliance evidence in real-time.

**Key Responsibilities**:
* Converts compliance evidence data into standardized OpenTelemetry log records.
* Emits log records to the OpenTelemetry Collector using the OTLP (OpenTelemetry Protocol).
* Provides metrics and tracing for evidence collection and processing.

`proofwatch` attributes defined [here](./attributes)

_Example code snippet_
```go
import (
    "context"
    "log"
    
    "go.opentelemetry.io/otel/log"
    "github.com/complytime/complybeacon/proofwatch"
)

// Create a new ProofWatch instance
pw, err := proofwatch.NewProofWatch()
if err != nil {
    log.Fatal(err)
}

// Create evidence (example with GemaraEvidence)
evidence := proofwatch.GemaraEvidence{
    // ... populate evidence fields
}

// Log evidence with default severity
err = pw.Log(ctx, evidence)
if err != nil {
    return fmt.Errorf("error logging evidence: %w", err)
}

// Or log with specific severity
err = pw.LogWithSeverity(ctx, evidence, olog.SeverityWarn)
```

### 2. Beacon Collector Distro

**Purpose**: A minimal OpenTelemetry Collector distribution that acts as the runtime environment for the `complybeacon` evidence pipeline.

**Key Responsibilities**:
* Receiving log records from sources like `proofwatch`
* Transforming log records into OCSF (Open Cybersecurity Schema Framework) format for standardized compliance event representation.
* Exporting the processed logs to configured backends (Loki, S3, SIEM).

The full evidence pipeline is validated by automated integration tests in `tests/integration/`. See `docs/DEVELOPMENT.md` for how to run them.
