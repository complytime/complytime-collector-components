#!/bin/bash
# Pre-release validation: checks that the project is ready to tag a release.
# Usage: release-prep.sh <version>   (e.g. release-prep.sh 0.1.0)
# Exit 0 = ready to tag, exit 1 = issues found.
set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
	echo "Usage: release-prep.sh <version>  (without 'v' prefix, e.g. 0.1.0)"
	exit 1
fi

# Strip leading 'v' if the caller accidentally included it
VERSION="${VERSION#v}"

# Validate semver format (major.minor.patch, no pre-release/build metadata)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "ERROR: '$VERSION' is not a valid semver (expected x.y.z)"
	exit 1
fi

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

FAILED=0

echo "=== Release prep check for v${VERSION} ==="
echo ""

# ── 1. Check that the tag does not already exist ────────────────
echo "--- Tag collision ---"
if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
	echo "  FAIL: Tag v${VERSION} already exists"
	FAILED=1
else
	echo "  OK: Tag v${VERSION} does not exist yet"
fi

# ── 2. Check proofwatch.Version() ───────────────────────────────
echo "--- proofwatch.Version() ---"
PW_FILE="proofwatch/proofwatch.go"
if [[ -f "$PW_FILE" ]]; then
	PW_VERSION=$(grep -oP 'return\s+"\K[0-9]+\.[0-9]+\.[0-9]+' "$PW_FILE" || true)
	if [[ "$PW_VERSION" == "$VERSION" ]]; then
		echo "  OK: proofwatch.Version() returns \"${VERSION}\""
	else
		echo "  FAIL: proofwatch.Version() returns \"${PW_VERSION}\" (expected \"${VERSION}\")"
		FAILED=1
	fi
else
	echo "  WARN: $PW_FILE not found, skipping"
fi

# ── 3. Check Containerfile IMAGE_VERSION ARG ────────────────────
echo "--- Containerfile IMAGE_VERSION ---"
CF="beacon-distro/Containerfile.collector"
if [[ -f "$CF" ]]; then
	CF_VERSION=$(grep -oP '^ARG IMAGE_VERSION=\K.*' "$CF" || true)
	if [[ "$CF_VERSION" == "$VERSION" ]]; then
		echo "  OK: ${CF} ARG IMAGE_VERSION=${VERSION}"
	elif [[ -z "$CF_VERSION" ]]; then
		echo "  FAIL: ${CF} is missing ARG IMAGE_VERSION"
		echo "  ACTION: Add 'ARG IMAGE_VERSION=${VERSION}' before the LABEL block in ${CF}"
		FAILED=1
	else
		# ARG exists but with wrong version — update in place
		sed -i "s/^ARG IMAGE_VERSION=.*/ARG IMAGE_VERSION=${VERSION}/" "$CF"
		echo "  FIXED: Updated ARG IMAGE_VERSION to ${VERSION} in ${CF}"
	fi
else
	echo "  WARN: $CF not found, skipping"
fi

# ── 4. Check CHANGELOG.md ───────────────────────────────────────
echo "--- CHANGELOG.md ---"
CL="CHANGELOG.md"
TODAY=$(date +%Y-%m-%d)
if [[ ! -f "$CL" ]]; then
	echo "  CHANGELOG.md not found, creating with template..."
	cat >"$CL" <<-TEMPLATE
		# Changelog

		All notable changes to this project will be documented in this file.

		The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
		and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

		## [${VERSION}] - ${TODAY}

		### Added

		- TODO: describe new features

		### Changed

		- TODO: describe changes to existing functionality

		### Fixed

		- TODO: describe bug fixes
	TEMPLATE
	# Dedent the heredoc (written with leading tabs for readability)
	sed -i 's/^\t//' "$CL"
	echo "  CREATED: $CL with template for ${VERSION}"
	echo "  ACTION: Fill in the changelog entries, then re-run this check"
	FAILED=1
elif grep -qP "## \[?${VERSION}\]?" "$CL"; then
	# Entry exists — check for unfilled TODO placeholders
	if sed -n "/## \[${VERSION}\]/,/^## /p" "$CL" | grep -q 'TODO:'; then
		echo "  WARN: CHANGELOG.md has an entry for ${VERSION} but contains TODO placeholders"
		FAILED=1
	else
		echo "  OK: CHANGELOG.md has an entry for ${VERSION}"
	fi
else
	# File exists but no entry for this version — insert template after the header
	echo "  No entry for ${VERSION}, adding template..."
	TEMPLATE=$(
		cat <<-EOF

			## [${VERSION}] - ${TODAY}

			### Added

			- TODO: describe new features

			### Changed

			- TODO: describe changes to existing functionality

			### Fixed

			- TODO: describe bug fixes
		EOF
	)
	# Dedent the template (strip one leading tab per line)
	TEMPLATE="${TEMPLATE#$'\t'}"          # strip leading tab on first line
	TEMPLATE="${TEMPLATE//$'\n\t'/$'\n'}" # strip leading tab on remaining lines

	# Insert after the first header line (# Changelog or similar)
	# If there's an [Unreleased] section, insert before it; otherwise after line 1
	if grep -qn '## \[Unreleased\]' "$CL"; then
		# Insert after the [Unreleased] section header
		UNRELEASED_LINE=$(grep -n '## \[Unreleased\]' "$CL" | head -1 | cut -d: -f1)
		# Find the next ## heading after Unreleased (that's where we insert before)
		NEXT_HEADING=$(tail -n "+$((UNRELEASED_LINE + 1))" "$CL" | grep -n '^## ' | head -1 | cut -d: -f1)
		if [[ -n "$NEXT_HEADING" ]]; then
			INSERT_LINE=$((UNRELEASED_LINE + NEXT_HEADING - 1))
		else
			# No next heading — append after Unreleased content
			INSERT_LINE=$(wc -l <"$CL")
		fi
		head -n "$INSERT_LINE" "$CL" >"${CL}.tmp"
		echo "$TEMPLATE" >>"${CL}.tmp"
		tail -n "+$((INSERT_LINE + 1))" "$CL" >>"${CL}.tmp"
		mv "${CL}.tmp" "$CL"
	else
		# No Unreleased section — insert after the first line
		head -n 1 "$CL" >"${CL}.tmp"
		echo "$TEMPLATE" >>"${CL}.tmp"
		tail -n +2 "$CL" >>"${CL}.tmp"
		mv "${CL}.tmp" "$CL"
	fi
	echo "  ADDED: Template entry for ${VERSION} in $CL"
	echo "  ACTION: Fill in the changelog entries, then re-run this check"
	FAILED=1
fi

# ── 5. Run dependency version check ─────────────────────────────
echo "--- Dependency version alignment ---"
if bash "${ROOT_DIR}/.taskfiles/scripts/version-check.sh" >/dev/null 2>&1; then
	echo "  OK: Go and OTel versions are aligned"
else
	echo "  FAIL: Dependency version drift detected (run 'task version:check' for details)"
	FAILED=1
fi

# ── 6. Check that HEAD is on main ───────────────────────────────
echo "--- Branch check ---"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
if [[ "$CURRENT_BRANCH" == "main" ]]; then
	echo "  OK: On branch main"
else
	echo "  WARN: On branch '${CURRENT_BRANCH}' (releases should be tagged from main)"
fi

# ── Summary ─────────────────────────────────────────────────────
echo ""
if [[ "$FAILED" -eq 0 ]]; then
	echo "All checks passed. Ready to tag:"
	echo "  git tag -s v${VERSION} -m \"Release v${VERSION}\""
	echo "  git push origin v${VERSION}"
else
	echo "Some checks failed. Fix the issues above before tagging."
	exit 1
fi
