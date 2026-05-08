#!/bin/bash
# Check that beacon-distro/manifest.yaml OTel versions align with truthbeam/go.mod
set -euo pipefail

MANIFEST="beacon-distro/manifest.yaml"
TRUTHBEAM_GOMOD="truthbeam/go.mod"

echo "Checking OTel Collector version consistency..."

# Extract collector version from truthbeam go.mod (using processorhelper as reference)
TRUTHBEAM_VERSION=$(grep 'go.opentelemetry.io/collector/processor/processorhelper' "$TRUTHBEAM_GOMOD" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | head -1)
if [ -z "$TRUTHBEAM_VERSION" ]; then
    echo "ERROR: Could not extract OTel Collector version from $TRUTHBEAM_GOMOD"
    exit 1
fi

echo "truthbeam requires: $TRUTHBEAM_VERSION"

# Check manifest.yaml components
FAILED=0
while IFS= read -r line; do
    COMPONENT=$(echo "$line" | grep -oE 'go\.opentelemetry\.io/collector/[^[:space:]]+' || true)
    VERSION=$(echo "$line" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true)

    if [ -n "$COMPONENT" ] && [ -n "$VERSION" ]; then
        if [ "$VERSION" != "$TRUTHBEAM_VERSION" ]; then
            echo "MISMATCH: $COMPONENT is at $VERSION (expected $TRUTHBEAM_VERSION)"
            FAILED=1
        fi
    fi
done < <(grep -E 'go.opentelemetry.io/collector/(exporter|processor|receiver)' "$MANIFEST")

# Check builder version in Containerfile
BUILDER_VERSION=$(grep 'go.opentelemetry.io/collector/cmd/builder@' beacon-distro/Containerfile.collector | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
if [ "$BUILDER_VERSION" != "$TRUTHBEAM_VERSION" ]; then
    echo "MISMATCH: Builder is at $BUILDER_VERSION (expected $TRUTHBEAM_VERSION)"
    FAILED=1
fi

if [ "$FAILED" -ne 0 ]; then
    echo ""
    echo "ERROR: OTel Collector version mismatch detected!"
    echo "Update manifest.yaml and Containerfile.collector to use $TRUTHBEAM_VERSION"
    exit 1
fi

echo "✓ All OTel Collector versions aligned at $TRUTHBEAM_VERSION"
