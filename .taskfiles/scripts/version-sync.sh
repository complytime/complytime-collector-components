#!/bin/bash
# Unified version sync: aligns Go and OTel versions across all modules,
# Containerfiles, CI workflows, and documentation.
#
# Source of truth:
#   Go version  — go.work `go` directive
#   OTel version — proofwatch/go.mod (stable v1.x series) + manifest.yaml (experimental v0.x)
#
# Strategy: Track stable (v1.x) versions. Go automatically pulls in
# the matching experimental (v0.x) versions as needed.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

GO_WORK="go.work"
PROOFWATCH_GOMOD="proofwatch/go.mod"
MANIFEST="beacon-distro/manifest.yaml"

# Auto-discover go.mod files and Containerfiles with Go base images
mapfile -t GO_MODS < <(find . -name go.mod -not -path '*/vendor/*' -print | sort || true)
mapfile -t CONTAINERFILES < <(grep -rl '^FROM golang:' . --include='Containerfile*' --include='Dockerfile*' 2>/dev/null | sort || true)

# Workspace modules (from go.work use block) — these carry OTel deps
mapfile -t WORKSPACE_MODULES < <(sed -n '/^use (/,/^)/{ s/^[[:space:]]*\.\///p }' "$GO_WORK" || true)

# ── Extract Go version from go.work ──────────────────────────────
GO_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GO_WORK" | head -1)
if [[ -z "$GO_VERSION" ]]; then
	echo "ERROR: Could not extract Go version from $GO_WORK"
	exit 1
fi
GO_MINOR="${GO_VERSION%.*}"
echo "=== Go version sync (source: $GO_WORK) ==="
echo "  Target: $GO_VERSION (minor: $GO_MINOR)"

# ── Sync go.mod files ────────────────────────────────────────────
for GOMOD in "${GO_MODS[@]}"; do
	perl -i -pe "s/^go \d+\.\d+(\.\d+)?$/go $GO_VERSION/" "$GOMOD"
	perl -i -ne 'print unless /^toolchain go/' "$GOMOD"
	echo "  go.mod: $GOMOD"
done

# ── Sync Containerfile Go image tags ─────────────────────────────
for CF in "${CONTAINERFILES[@]}"; do
	perl -i -pe "s{FROM golang:\d+\.\d+\.\d+}{FROM golang:$GO_VERSION}" "$CF"
	echo "  Containerfile: $CF"
done

# ── Sync CI workflow GO_VERSION ──────────────────────────────────
# Pin to full patch version (e.g., 1.26.4) to prevent setup-go from
# auto-upgrading to a newer patch and triggering version-check drift.
CI_WORKFLOWS=(
	".github/workflows/ci_local.yml"
	".github/workflows/ci_sonarcloud.yml"
)
for CI_WF in "${CI_WORKFLOWS[@]}"; do
	if [[ -f "$CI_WF" ]]; then
		perl -i -pe "s{^(\s*GO_VERSION:\s*)\S+}{\${1}$GO_VERSION}" "$CI_WF"
		echo "  CI workflow: $CI_WF (GO_VERSION: $GO_VERSION)"
	fi
done

# ── Sync documentation ──────────────────────────────────────────
DOCS=(
	"docs/DEVELOPMENT.md"
	"README.md"
	"AGENTS.md"
)
for DOC in "${DOCS[@]}"; do
	if [[ ! -f "$DOC" ]]; then
		continue
	fi
	perl -i -pe "s{Go \d+\.\d+\+}{Go ${GO_MINOR}+}g" "$DOC"
	perl -i -pe "s{Go \d+\.\d+\.\d+}{Go $GO_VERSION}g" "$DOC"
	perl -i -pe "s{golang:\d+\.\d+\.\d+}{golang:$GO_VERSION}g" "$DOC"
	echo "  Doc: $DOC"
done

echo ""

# ── Extract OTel versions ────────────────────────────────────────
echo "=== OTel version sync (sources: $PROOFWATCH_GOMOD + $MANIFEST) ==="

# Stable (v1.x) from proofwatch — the library module tracks the collector pdata series
OTEL_STABLE=$(sed -n '/^require (/,/^)/p' "$PROOFWATCH_GOMOD" |
	grep 'go.opentelemetry.io/collector' |
	grep -oE 'v1\.[0-9]+\.[0-9]+' |
	sort -V -u | tail -1)

# Experimental (v0.x) from manifest — the distro pins component versions explicitly
OTEL_EXPERIMENTAL=$(grep -E 'go\.opentelemetry\.io/collector/(exporter|processor|receiver)' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v0\.[0-9]+\.[0-9]+' |
	sort -V -u | tail -1)

if [[ -z "$OTEL_STABLE" ]]; then
	echo "ERROR: Could not extract stable (v1.x) OTel version from $PROOFWATCH_GOMOD"
	exit 1
fi
if [[ -z "$OTEL_EXPERIMENTAL" ]]; then
	echo "ERROR: Could not extract experimental (v0.x) OTel version from $MANIFEST"
	exit 1
fi

echo "  Stable (tracking): $OTEL_STABLE"
echo "  Experimental (derived): $OTEL_EXPERIMENTAL"

# ── Sync OTel versions in workspace module go.mod files ──────────
for MODULE in "${WORKSPACE_MODULES[@]}"; do
	GOMOD="$MODULE/go.mod"
	if [[ ! -f "$GOMOD" ]]; then
		continue
	fi
	perl -i -pe "s{(go\.opentelemetry\.io/collector/\S+)\s+v0\.\d+\.\d+}{\$1 $OTEL_EXPERIMENTAL}g" "$GOMOD"
	perl -i -pe "s{(go\.opentelemetry\.io/collector/\S+)\s+v1\.\d+\.\d+}{\$1 $OTEL_STABLE}g" "$GOMOD"
	echo "  go.mod: $GOMOD"
done

# ── Sync manifest.yaml component versions ───────────────────────
# The manifest uses experimental (v0.x) for components and stable (v1.x) for confmap providers.
if [[ -f "$MANIFEST" ]]; then
	perl -i -pe "s{(gomod: go\.opentelemetry\.io/collector/\S+)\s+v0\.\d+\.\d+}{\$1 $OTEL_EXPERIMENTAL}g" "$MANIFEST"
	perl -i -pe "s{(gomod: go\.opentelemetry\.io/collector/\S+)\s+v1\.\d+\.\d+}{\$1 $OTEL_STABLE}g" "$MANIFEST"
	perl -i -pe "s{(gomod: github\.com/open-telemetry/opentelemetry-collector-contrib/\S+)\s+v0\.\d+\.\d+}{\$1 $OTEL_EXPERIMENTAL}g" "$MANIFEST"
	echo "  Manifest: $MANIFEST (experimental=$OTEL_EXPERIMENTAL, stable=$OTEL_STABLE)"
fi

# ── Sync Containerfile builder version ───────────────────────────
# Builder uses the experimental version from the manifest components.
COLLECTOR_CF="beacon-distro/Containerfile.collector"
if [[ -f "$COLLECTOR_CF" ]]; then
	# Builder version matches the experimental version from manifest
	BUILDER_VERSION="$OTEL_EXPERIMENTAL"
	if [[ -n "$BUILDER_VERSION" ]]; then
		perl -i -pe "s{builder\@v[\d.]+}{builder\@$BUILDER_VERSION}g" "$COLLECTOR_CF"
		echo "  Builder: $COLLECTOR_CF (using manifest experimental version: $BUILDER_VERSION)"
	fi
fi

echo ""
echo "=== Running go mod tidy ==="
for MODULE in "${WORKSPACE_MODULES[@]}"; do
	if [[ ! -f "$MODULE/go.mod" ]]; then
		continue
	fi
	echo "  Tidying $MODULE..."
	(cd "$MODULE" && GOTOOLCHAIN=auto go mod tidy) || {
		echo "Tidy failed for $MODULE"
		exit 1
	}
done

echo ""
echo "=== Version sync complete ==="
echo "  Go: $GO_VERSION | OTel: stable=$OTEL_STABLE experimental=$OTEL_EXPERIMENTAL"
echo ""
echo "NOTE: Containerfile @sha256: digests are NOT updated automatically."
echo "      To update digests, pull the new image and replace the hash:"
echo "      podman pull golang:$GO_VERSION && podman inspect golang:$GO_VERSION --format '{{.Digest}}'"
