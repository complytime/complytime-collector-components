#!/bin/bash
# Auto-sync all OTel Collector versions to the highest version found across all modules
set -euo pipefail

MODULES="./proofwatch ./truthbeam"
MANIFEST="beacon-distro/manifest.yaml"
CONTAINERFILE="beacon-distro/Containerfile.collector"

echo "Finding highest OTel Collector versions across all modules..."

# Find highest v0.x.x experimental version across all go.mod files
HIGHEST_EXPERIMENTAL=""
for MODULE in $MODULES; do
    GOMOD="$MODULE/go.mod"
    if [ ! -f "$GOMOD" ]; then
        continue
    fi

    VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
               grep -oE 'v0\.[0-9]+\.[0-9]+' | \
               sort -V -u || true)

    for VERSION in $VERSIONS; do
        if [ -z "$HIGHEST_EXPERIMENTAL" ] || [ "$(printf '%s\n' "$HIGHEST_EXPERIMENTAL" "$VERSION" | sort -V | tail -1)" = "$VERSION" ]; then
            HIGHEST_EXPERIMENTAL="$VERSION"
        fi
    done
done

# Find highest v1.x.x stable version across all go.mod files
HIGHEST_STABLE=""
for MODULE in $MODULES; do
    GOMOD="$MODULE/go.mod"
    if [ ! -f "$GOMOD" ]; then
        continue
    fi

    VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" | \
               grep -oE 'v1\.[0-9]+\.[0-9]+' | \
               sort -V -u || true)

    for VERSION in $VERSIONS; do
        if [ -z "$HIGHEST_STABLE" ] || [ "$(printf '%s\n' "$HIGHEST_STABLE" "$VERSION" | sort -V | tail -1)" = "$VERSION" ]; then
            HIGHEST_STABLE="$VERSION"
        fi
    done
done

if [ -z "$HIGHEST_EXPERIMENTAL" ]; then
    echo "ERROR: Could not find any experimental (v0.x) OTel Collector versions"
    exit 1
fi

if [ -z "$HIGHEST_STABLE" ]; then
    echo "ERROR: Could not find any stable (v1.x) OTel Collector versions"
    exit 1
fi

echo "  Highest experimental (v0.x): $HIGHEST_EXPERIMENTAL"
echo "  Highest stable (v1.x): $HIGHEST_STABLE"
echo ""

# Update each go.mod file
for MODULE in $MODULES; do
    GOMOD="$MODULE/go.mod"
    if [ ! -f "$GOMOD" ]; then
        continue
    fi

    echo "Updating $GOMOD..."

    # Update experimental packages to highest experimental version
    perl -i -pe "s{(go\.opentelemetry\.io/collector/[^/]+) v0\.[0-9]+\.[0-9]+}{\$1 $HIGHEST_EXPERIMENTAL}g" "$GOMOD"

    # Update stable packages to highest stable version
    perl -i -pe "s{(go\.opentelemetry\.io/collector/[^/]+) v1\.[0-9]+\.[0-9]+}{\$1 $HIGHEST_STABLE}g" "$GOMOD"

    # Run go mod tidy to ensure dependencies are resolved
    echo "  Running go mod tidy..."
    (cd "$MODULE" && GOWORK=off go mod tidy)

    echo "  ✓ $MODULE updated to experimental=$HIGHEST_EXPERIMENTAL stable=$HIGHEST_STABLE"
done

echo ""
echo "Updating manifest and Containerfile..."

# Update manifest.yaml
perl -i -pe "s{(go\.opentelemetry\.io/collector/(exporter|processor|receiver)/\w+) v[\d.]+}{\$1 $HIGHEST_EXPERIMENTAL}g" "$MANIFEST"
perl -i -pe "s{(github\.com/open-telemetry/opentelemetry-collector-contrib/(exporter|processor|receiver|connector|extension)/\w+) v[\d.]+}{\$1 $HIGHEST_EXPERIMENTAL}g" "$MANIFEST"
perl -i -pe "s{(go\.opentelemetry\.io/collector/confmap/provider/\w+) v[\d.]+}{\$1 $HIGHEST_STABLE}g" "$MANIFEST"

# Update Containerfile builder version
perl -i -pe "s{builder\@v[\d.]+}{builder\@$HIGHEST_EXPERIMENTAL}g" "$CONTAINERFILE"
perl -i -pe "s{builder version \(v[\d.]+\)}{builder version ($HIGHEST_EXPERIMENTAL)}g" "$CONTAINERFILE"

echo "  ✓ manifest.yaml and Containerfile.collector updated"
echo ""
echo "✓ All OTel Collector versions synced to experimental=$HIGHEST_EXPERIMENTAL stable=$HIGHEST_STABLE"
