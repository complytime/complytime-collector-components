#!/bin/bash
# Sync manifest.yaml versions from truthbeam/go.mod (the idiomatic way)
# This keeps manifest.yaml in sync with Go module versions
set -euo pipefail

TRUTHBEAM_GOMOD="truthbeam/go.mod"
MANIFEST="beacon-distro/manifest.yaml"

echo "Syncing OTel Collector versions from truthbeam to manifest..."

# Extract OTel Collector version from truthbeam (using processorhelper as reference)
OTEL_VERSION=$(grep 'go.opentelemetry.io/collector/processor/processorhelper' "$TRUTHBEAM_GOMOD" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -z "$OTEL_VERSION" ]; then
    echo "ERROR: Could not extract OTel Collector version from $TRUTHBEAM_GOMOD"
    exit 1
fi

# Extract provider version (different version scheme)
PROVIDER_VERSION=$(grep 'go.opentelemetry.io/collector/component v' "$TRUTHBEAM_GOMOD" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -z "$PROVIDER_VERSION" ]; then
    echo "ERROR: Could not extract provider version from $TRUTHBEAM_GOMOD"
    exit 1
fi

echo "  OTel Collector: $OTEL_VERSION"
echo "  Providers: $PROVIDER_VERSION"

# Update versions using perl (more reliable than sed for complex patterns)
perl -i -pe "s{(go\.opentelemetry\.io/collector/(exporter|processor|receiver)/\w+) v[\d.]+}{\$1 $OTEL_VERSION}g" "$MANIFEST"
perl -i -pe "s{(github\.com/open-telemetry/opentelemetry-collector-contrib/(exporter|processor|receiver|connector|extension)/\w+) v[\d.]+}{\$1 $OTEL_VERSION}g" "$MANIFEST"
perl -i -pe "s{(go\.opentelemetry\.io/collector/confmap/provider/\w+) v[\d.]+}{\$1 $PROVIDER_VERSION}g" "$MANIFEST"

# Update builder version in Containerfile
perl -i -pe "s{builder\@v[\d.]+}{builder\@$OTEL_VERSION}g" beacon-distro/Containerfile.collector
perl -i -pe "s{builder version \(v[\d.]+\)}{builder version ($OTEL_VERSION)}g" beacon-distro/Containerfile.collector

echo "✓ Synced all versions to OTel $OTEL_VERSION / Providers $PROVIDER_VERSION"
