#!/bin/bash
# Read-only version validation: checks that Go and OTel versions are
# consistent across all modules, Containerfiles, and manifest.yaml.
# Exit 0 = all aligned, exit 1 = drift detected.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

FAILED=0

GO_WORK="go.work"
MANIFEST="beacon-distro/manifest.yaml"

# Auto-discover go.mod files and Containerfiles with Go base images
mapfile -t GO_MODS < <(find . -name go.mod -not -path '*/vendor/*' -print | sort || true)
mapfile -t CONTAINERFILES < <(grep -rl '^FROM golang:' . --include='Containerfile*' --include='Dockerfile*' 2>/dev/null | sort || true)

# Workspace modules (from go.work use block) — these carry OTel deps
mapfile -t WORKSPACE_MODULES < <(sed -n '/^use (/,/^)/{ s/^[[:space:]]*\.\///p }' "$GO_WORK" || true)

# ── Go version alignment ────────────────────────────────────────
echo "=== Go version check ==="

GO_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GO_WORK" | head -1)
if [[ -z "$GO_VERSION" ]]; then
	echo "ERROR: Could not extract Go version from $GO_WORK"
	exit 1
fi
echo "  Source of truth ($GO_WORK): $GO_VERSION"

for GOMOD in "${GO_MODS[@]}"; do
	MOD_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$GOMOD" | head -1)
	if [[ -z "$MOD_VERSION" ]]; then
		MOD_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\)/\1/p' "$GOMOD" | head -1)
	fi
	if [[ "$MOD_VERSION" != "$GO_VERSION" ]]; then
		echo "  FAIL: $GOMOD has go $MOD_VERSION (expected $GO_VERSION)"
		FAILED=1
	else
		echo "  OK: $GOMOD"
	fi
done

for CF in "${CONTAINERFILES[@]}"; do
	CF_VERSION=$(sed -n 's/^FROM golang:\([0-9]*\.[0-9]*\.[0-9]*\).*/\1/p' "$CF" | head -1)
	if [[ -z "$CF_VERSION" ]]; then
		echo "  WARNING: Could not extract Go version from $CF"
		continue
	fi
	if [[ "$CF_VERSION" != "$GO_VERSION" ]]; then
		echo "  FAIL: $CF uses golang:$CF_VERSION (expected $GO_VERSION)"
		FAILED=1
	else
		echo "  OK: $CF"
	fi
done

for GOMOD in "${GO_MODS[@]}"; do
	if grep -q '^toolchain go' "$GOMOD"; then
		echo "  FAIL: $GOMOD has stale toolchain directive"
		FAILED=1
	fi
done

# Check CI workflows use go-version-file (preferred) or a matching GO_VERSION pin
CI_WORKFLOWS=(
	".github/workflows/ci_local.yml"
	".github/workflows/ci_sonarcloud.yml"
)
for CI_WF in "${CI_WORKFLOWS[@]}"; do
	if [[ ! -f "$CI_WF" ]]; then
		continue
	fi
	# Accept go-version-file as the preferred approach — the version is
	# read from go.mod at CI time, so no hardcoded pin to drift.
	if grep -qE '^\s*go-version-file:' "$CI_WF"; then
		echo "  OK: $CI_WF (uses go-version-file)"
		continue
	fi
	CI_GO=$(grep -E '^\s*GO_VERSION:' "$CI_WF" | head -1 | sed 's/.*GO_VERSION:\s*//' | tr -d ' ')
	if [[ "$CI_GO" != "$GO_VERSION" ]]; then
		echo "  FAIL: $CI_WF has GO_VERSION: $CI_GO (expected $GO_VERSION)"
		FAILED=1
	else
		echo "  OK: $CI_WF"
	fi
done

echo ""

# ── OTel version consistency ────────────────────────────────────
echo "=== OTel version check ==="

# Both version series are derived from the manifest — the single source of truth
# for the collector stack. proofwatch is a separate Go module that only uses pdata
# and manages its own OTel versions independently.
# See: https://github.com/complytime/complytime-collector-components/issues/430

# Experimental (v0.x) from manifest components
OTEL_EXPERIMENTAL=$(grep -E 'go\.opentelemetry\.io/collector/(exporter|processor|receiver)' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v0\.[0-9]+\.[0-9]+' |
	sort -V -u | tail -1)

# Stable (v1.x) from manifest providers (confmap providers use the stable series)
OTEL_STABLE=$(grep -E 'go\.opentelemetry\.io/collector/confmap/provider' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v1\.[0-9]+\.[0-9]+' |
	sort -V -u | tail -1)

if [[ -z "$OTEL_EXPERIMENTAL" ]]; then
	echo "ERROR: Could not extract experimental (v0.x) OTel version from $MANIFEST"
	exit 1
fi
if [[ -z "$OTEL_STABLE" ]]; then
	echo "ERROR: Could not extract stable (v1.x) OTel version from $MANIFEST"
	exit 1
fi

echo "  Source of truth ($MANIFEST): experimental=$OTEL_EXPERIMENTAL stable=$OTEL_STABLE"

# Modules that are part of the collector stack — proofwatch is excluded because
# it only depends on pdata and evolves its OTel version independently.
COLLECTOR_MODULES=()
for MODULE in "${WORKSPACE_MODULES[@]}"; do
	if [[ "$MODULE" == "proofwatch" ]]; then
		echo "  SKIP: $MODULE/go.mod (independent pdata consumer, not part of collector stack)"
		continue
	fi
	COLLECTOR_MODULES+=("$MODULE")
done

for MODULE in "${COLLECTOR_MODULES[@]}"; do
	GOMOD="$MODULE/go.mod"
	if [[ ! -f "$GOMOD" ]]; then
		continue
	fi

	EXP_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" |
		grep -v 'go.opentelemetry.io/contrib' |
		grep -oE 'v0\.[0-9]+\.[0-9]+' |
		sort -u || true)

	if [[ -n "$EXP_VERSIONS" ]]; then
		EXP_COUNT=$(echo "$EXP_VERSIONS" | wc -l)
		if [[ "$EXP_COUNT" -gt 1 ]]; then
			echo "  WARNING: $GOMOD has mixed experimental OTel versions (transitive dependencies):"
			echo "${EXP_VERSIONS//$'\n'/$'\n'    }"
			# Not a failure - MVS can pull in multiple versions via transitive deps
		else
			FIRST_EXP=$(echo "$EXP_VERSIONS" | head -1 || true)
			echo "  OK: $GOMOD experimental at $FIRST_EXP"
		fi
	fi

	STABLE_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/[^/]+' "$GOMOD" |
		grep -v 'go.opentelemetry.io/contrib' |
		grep -oE 'v1\.[0-9]+\.[0-9]+' |
		sort -u || true)

	if [[ -n "$STABLE_VERSIONS" ]]; then
		STABLE_COUNT=$(echo "$STABLE_VERSIONS" | wc -l)
		if [[ "$STABLE_COUNT" -gt 1 ]]; then
			echo "  WARNING: $GOMOD has mixed stable OTel versions (transitive dependencies):"
			echo "${STABLE_VERSIONS//$'\n'/$'\n'    }"
			# Not a failure - MVS can pull in multiple versions via transitive deps
		else
			FIRST_STABLE=$(echo "$STABLE_VERSIONS" | head -1 || true)
			echo "  OK: $GOMOD stable at $FIRST_STABLE"
		fi
	fi
done

# ── Manifest version check ──────────────────────────────────────
# The manifest uses experimental (v0.x) for components/contrib and stable (v1.x) for
# confmap providers (which migrated to the stable series upstream).
# Local module placeholders (v0.0.0) are excluded from this check.

MANIFEST_FAIL=0

MANIFEST_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/(exporter|processor|receiver)' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

if [[ -n "$MANIFEST_VERSIONS" ]]; then
	for V in $MANIFEST_VERSIONS; do
		if [[ "$V" != "$OTEL_EXPERIMENTAL" ]]; then
			echo "  FAIL: $MANIFEST has component at $V (expected $OTEL_EXPERIMENTAL)"
			FAILED=1
			MANIFEST_FAIL=1
		fi
	done
fi

PROVIDER_VERSIONS=$(grep -E 'go\.opentelemetry\.io/collector/confmap/provider' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

if [[ -n "$PROVIDER_VERSIONS" ]]; then
	for V in $PROVIDER_VERSIONS; do
		if [[ "$V" != "$OTEL_STABLE" ]]; then
			echo "  FAIL: $MANIFEST has provider at $V (expected $OTEL_STABLE)"
			FAILED=1
			MANIFEST_FAIL=1
		fi
	done
fi

CONTRIB_VERSIONS=$(grep -E 'github\.com/open-telemetry/opentelemetry-collector-contrib' "$MANIFEST" |
	grep -v '^\s*#' |
	grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' | sort -u || true)

if [[ -n "$CONTRIB_VERSIONS" ]]; then
	for V in $CONTRIB_VERSIONS; do
		if [[ "$V" != "$OTEL_EXPERIMENTAL" ]]; then
			echo "  FAIL: $MANIFEST has contrib at $V (expected $OTEL_EXPERIMENTAL)"
			FAILED=1
			MANIFEST_FAIL=1
		fi
	done
fi

# Report success if all manifest versions match
if [[ "$MANIFEST_FAIL" -eq 0 && -n "$MANIFEST_VERSIONS$PROVIDER_VERSIONS$CONTRIB_VERSIONS" ]]; then
	echo "  OK: $MANIFEST components at $OTEL_EXPERIMENTAL, providers at $OTEL_STABLE"
fi

COLLECTOR_CF="beacon-distro/Containerfile.collector"
if [[ -f "$COLLECTOR_CF" ]]; then
	BUILDER_VERSION=$(grep 'go.opentelemetry.io/collector/cmd/builder@' "$COLLECTOR_CF" | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || true)
	# Builder should use the experimental version from manifest components.
	if [[ -n "$BUILDER_VERSION" && "$BUILDER_VERSION" != "$OTEL_EXPERIMENTAL" ]]; then
		echo "  FAIL: Builder at $BUILDER_VERSION (expected experimental version: $OTEL_EXPERIMENTAL)"
		FAILED=1
	elif [[ -n "$BUILDER_VERSION" ]]; then
		echo "  OK: Builder at $BUILDER_VERSION (experimental version)"
	fi
fi

# Summary already reported in sections above

echo ""

if [[ "$FAILED" -ne 0 ]]; then
	echo "FAILED: Version drift detected. Run 'task version:sync' to fix."
	exit 1
fi

echo "All version checks passed."
