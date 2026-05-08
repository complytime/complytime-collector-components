#!/bin/bash
# Check that all OTel Collector dependencies within each go.mod are at consistent versions
set -euo pipefail

MODULES="./proofwatch ./truthbeam"

echo "Checking internal go.mod version consistency..."

TOTAL_FAILED=0

for MODULE in $MODULES; do
    GOMOD="$MODULE/go.mod"

    if [ ! -f "$GOMOD" ]; then
        echo "WARNING: $GOMOD not found, skipping"
        continue
    fi

    echo ""
    echo "Checking $GOMOD..."

    # OTel Collector uses dual versioning:
    # - v1.x.x for stable/core APIs (component, consumer, pdata, processor, etc.)
    # - v0.x.x for experimental packages (componenttest, processorhelper, xprocessor, etc.)
    #
    # We check consistency within each scheme separately

    # Check v0.x.x experimental packages
    EXPERIMENTAL_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
                           grep -v 'go.opentelemetry.io/contrib' | \
                           grep -oE 'v0\.[0-9]+\.[0-9]+' | \
                           sort -u || true)

    if [ -n "$EXPERIMENTAL_VERSIONS" ]; then
        EXP_COUNT=$(echo "$EXPERIMENTAL_VERSIONS" | wc -l)
        if [ "$EXP_COUNT" -eq 1 ]; then
            echo "  ✓ All experimental OTel packages at $(echo "$EXPERIMENTAL_VERSIONS" | head -1)"
        else
            echo "  ERROR: Multiple experimental (v0.x) OTel Collector versions detected:"
            while IFS= read -r version; do
                echo "    - $version"
                grep -E "go\.opentelemetry\.io/collector/[^/]+.*$version" "$GOMOD" | sed 's/^/      /'
            done <<< "$EXPERIMENTAL_VERSIONS"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
        fi
    fi

    # Check v1.x.x stable packages
    STABLE_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
                     grep -v 'go.opentelemetry.io/contrib' | \
                     grep -oE 'v1\.[0-9]+\.[0-9]+' | \
                     sort -u || true)

    if [ -n "$STABLE_VERSIONS" ]; then
        STABLE_COUNT=$(echo "$STABLE_VERSIONS" | wc -l)
        if [ "$STABLE_COUNT" -eq 1 ]; then
            echo "  ✓ All stable OTel packages at $(echo "$STABLE_VERSIONS" | head -1)"
        else
            echo "  ERROR: Multiple stable (v1.x) OTel Collector versions detected:"
            while IFS= read -r version; do
                echo "    - $version"
                grep -E "go\.opentelemetry\.io/collector/[^/]+.*$version" "$GOMOD" | sed 's/^/      /'
            done <<< "$STABLE_VERSIONS"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
        fi
    fi

    # Check contrib separately (different versioning scheme is acceptable)
    CONTRIB_VERSIONS=$(grep -E 'github\.com/open-telemetry/opentelemetry-collector-contrib' "$GOMOD" | \
                       grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | \
                       sort -u || true)

    if [ -n "$CONTRIB_VERSIONS" ]; then
        CONTRIB_COUNT=$(echo "$CONTRIB_VERSIONS" | wc -l)
        if [ "$CONTRIB_COUNT" -eq 1 ]; then
            echo "  ✓ All OTel Collector contrib dependencies at $(echo "$CONTRIB_VERSIONS" | head -1)"
        else
            echo "  ERROR: Multiple OTel Collector contrib versions detected:"
            while IFS= read -r version; do
                echo "    - $version"
                grep -E "github\.com/open-telemetry/opentelemetry-collector-contrib.*$version" "$GOMOD" | sed 's/^/      /'
            done <<< "$CONTRIB_VERSIONS"
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
        fi
    fi
done

echo ""
if [ "$TOTAL_FAILED" -gt 0 ]; then
    echo "ERROR: Found version inconsistencies in $TOTAL_FAILED module(s)"
    echo "All OTel Collector dependencies within a go.mod should be at the same version."
    exit 1
fi

echo "✓ All go.mod files have consistent OTel Collector versions"
