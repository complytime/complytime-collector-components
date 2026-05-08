#!/bin/bash
# Check that Containerfile Go version satisfies all module requirements
set -euo pipefail

CONTAINERFILE="beacon-distro/Containerfile.collector"
MODULES="./proofwatch ./truthbeam"

# Extract Go version from Containerfile
CF_VERSION=$(sed -n 's/^FROM golang:\([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p' "$CONTAINERFILE" | head -1)
if [ -z "$CF_VERSION" ]; then
    echo "ERROR: Could not extract Go version from $CONTAINERFILE"
    exit 1
fi

CF_MAJOR=$(echo "$CF_VERSION" | cut -d. -f1)
CF_MINOR=$(echo "$CF_VERSION" | cut -d. -f2)
echo "Containerfile Go version: $CF_VERSION (major=$CF_MAJOR minor=$CF_MINOR)"

# Check each module's go.mod
FAILED=0
for m in $MODULES; do
    MOD_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\).*/\1/p' "$m/go.mod" | head -1)
    if [ -z "$MOD_VERSION" ]; then
        echo "WARNING: Could not extract go directive from $m/go.mod"
        continue
    fi

    MOD_MAJOR=$(echo "$MOD_VERSION" | cut -d. -f1)
    MOD_MINOR=$(echo "$MOD_VERSION" | cut -d. -f2)

    if [ "$CF_MAJOR" -lt "$MOD_MAJOR" ] || \
       { [ "$CF_MAJOR" -eq "$MOD_MAJOR" ] && [ "$CF_MINOR" -lt "$MOD_MINOR" ]; }; then
        echo "FAIL: $m/go.mod requires go $MOD_VERSION but $CONTAINERFILE uses $CF_VERSION"
        FAILED=1
    else
        echo "OK: $m/go.mod requires go $MOD_VERSION <= $CF_VERSION"
    fi
done

if [ "$FAILED" -ne 0 ]; then
    echo ""
    echo "ERROR: Containerfile Go version is behind module requirements."
    echo "Update the FROM golang:X.Y.Z line in $CONTAINERFILE."
    exit 1
fi

echo "--- Go version check passed ---"
