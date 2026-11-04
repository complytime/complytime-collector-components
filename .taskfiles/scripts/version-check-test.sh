#!/bin/bash
# Regression coverage for version validation's workspace-state independence.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
WORKSPACE_FILE="$ROOT_DIR/go.work"
TEMP_DIR="$(mktemp -d)"
FAKE_BIN_DIR="$TEMP_DIR/bin"
BACKUP_WORKSPACE="$TEMP_DIR/original-go.work"
MISSING_WORKSPACE_OUTPUT="$TEMP_DIR/missing-workspace.out"
CONFLICTING_WORKSPACE_OUTPUT="$TEMP_DIR/conflicting-workspace.out"
CONTROLLED_RUNTIME_OUTPUT="$TEMP_DIR/controlled-runtime.out"
LS_FILES_FAILURE_OUTPUT="$TEMP_DIR/ls-files-failure.out"
GREP_FAILURE_OUTPUT="$TEMP_DIR/grep-failure.out"
GREP_NO_MATCH_OUTPUT="$TEMP_DIR/grep-no-match.out"
CONFLICTING_WORKSPACE="$TEMP_DIR/conflicting-go.work"
ROOT_GO_VERSION=$(sed -n 's/^go \([0-9]*\.[0-9]*\.[0-9]*\)/\1/p' "$ROOT_DIR/go.mod" | head -1)
FAKE_GO_VERSION="9.99.99"
SYSTEM_GIT=$(command -v git)
ORIGINAL_WORKSPACE_BACKED_UP=false
TEST_OWNS_WORKSPACE=false

cleanup() {
	if [[ "$TEST_OWNS_WORKSPACE" == true ]]; then
		rm -f "$WORKSPACE_FILE"
	fi
	if [[ "$ORIGINAL_WORKSPACE_BACKED_UP" == true ]]; then
		mv "$BACKUP_WORKSPACE" "$WORKSPACE_FILE"
	fi
	rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

run_version_check() {
	task --silent version:check >"$1" 2>&1
}

run_with_controlled_go_version() {
	PATH="$FAKE_BIN_DIR:$PATH" task --silent version:check >"$1" 2>&1
}

run_with_git_failure() {
	PATH="$FAKE_BIN_DIR:$PATH" VERSION_CHECK_TEST_GIT_FAILURE="$2" \
		task --silent version:check >"$1" 2>&1
}

if [[ -f "$WORKSPACE_FILE" ]]; then
	if ! mv "$WORKSPACE_FILE" "$BACKUP_WORKSPACE"; then
		echo "could not back up the existing go.work" >&2
		exit 1
	fi
	ORIGINAL_WORKSPACE_BACKED_UP=true
fi

if ! run_version_check "$MISSING_WORKSPACE_OUTPUT"; then
	cat "$MISSING_WORKSPACE_OUTPUT"
	echo "version validation failed without go.work" >&2
	exit 1
fi

if [[ -e "$WORKSPACE_FILE" ]]; then
	echo "version validation created go.work" >&2
	exit 1
fi

cat >"$CONFLICTING_WORKSPACE" <<'EOF'
go 1.24.1

use .
EOF
TEST_OWNS_WORKSPACE=true
cp "$CONFLICTING_WORKSPACE" "$WORKSPACE_FILE"

if git ls-files --error-unmatch go.work >/dev/null 2>&1; then
	echo "test requires go.work to remain untracked" >&2
	exit 1
fi

if ! run_version_check "$CONFLICTING_WORKSPACE_OUTPUT"; then
	cat "$CONFLICTING_WORKSPACE_OUTPUT"
	echo "version validation failed with a conflicting go.work" >&2
	exit 1
fi

if ! cmp -s "$CONFLICTING_WORKSPACE" "$WORKSPACE_FILE"; then
	echo "version validation modified the untracked go.work" >&2
	exit 1
fi

if ! cmp -s "$MISSING_WORKSPACE_OUTPUT" "$CONFLICTING_WORKSPACE_OUTPUT"; then
	echo "version validation output differed by workspace state:" >&2
	diff -u "$MISSING_WORKSPACE_OUTPUT" "$CONFLICTING_WORKSPACE_OUTPUT" >&2 || true
	exit 1
fi

mkdir -p "$FAKE_BIN_DIR"
cat >"$FAKE_BIN_DIR/go" <<EOF
#!/bin/bash
if [[ "\${1:-}" == "version" ]]; then
	echo "go version go$FAKE_GO_VERSION controlled"
	exit 0
fi

echo "unexpected Go command: \$*" >&2
exit 1
EOF
chmod 755 "$FAKE_BIN_DIR/go"

if ! run_with_controlled_go_version "$CONTROLLED_RUNTIME_OUTPUT"; then
	cat "$CONTROLLED_RUNTIME_OUTPUT"
	echo "version validation used the controlled Go installation" >&2
	exit 1
fi

if ! grep -Fq "Source of truth (go.mod): $ROOT_GO_VERSION" "$CONTROLLED_RUNTIME_OUTPUT"; then
	cat "$CONTROLLED_RUNTIME_OUTPUT"
	echo "version validation did not use the root go.mod declaration" >&2
	exit 1
fi

if grep -Fq "$FAKE_GO_VERSION" "$CONTROLLED_RUNTIME_OUTPUT"; then
	cat "$CONTROLLED_RUNTIME_OUTPUT"
	echo "version validation used the controlled Go patch version" >&2
	exit 1
fi

cat >"$FAKE_BIN_DIR/git" <<EOF
#!/bin/bash
case "\${VERSION_CHECK_TEST_GIT_FAILURE:-}" in
	ls-files)
		if [[ "\${1:-}" == "ls-files" ]]; then
			echo "controlled git ls-files failure" >&2
			exit 2
		fi
		;;
	grep)
		if [[ "\${1:-}" == "grep" ]]; then
			echo "controlled git grep failure" >&2
			exit 2
		fi
		;;
	grep-no-match)
		if [[ "\${1:-}" == "grep" ]]; then
			exit 1
		fi
		;;
esac
exec "$SYSTEM_GIT" "\$@"
EOF
chmod 755 "$FAKE_BIN_DIR/git"

if run_with_git_failure "$LS_FILES_FAILURE_OUTPUT" ls-files; then
	cat "$LS_FILES_FAILURE_OUTPUT"
	echo "version validation ignored a git ls-files failure" >&2
	exit 1
fi

if ! grep -Fq "ERROR: Could not discover tracked Go module files" "$LS_FILES_FAILURE_OUTPUT"; then
	cat "$LS_FILES_FAILURE_OUTPUT"
	echo "version validation did not report the git ls-files failure" >&2
	exit 1
fi

if run_with_git_failure "$GREP_FAILURE_OUTPUT" grep; then
	cat "$GREP_FAILURE_OUTPUT"
	echo "version validation ignored a git grep failure" >&2
	exit 1
fi

if ! grep -Fq "ERROR: Could not discover tracked Go Containerfiles" "$GREP_FAILURE_OUTPUT"; then
	cat "$GREP_FAILURE_OUTPUT"
	echo "version validation did not report the git grep failure" >&2
	exit 1
fi

if ! run_with_git_failure "$GREP_NO_MATCH_OUTPUT" grep-no-match; then
	cat "$GREP_NO_MATCH_OUTPUT"
	echo "version validation rejected git grep's no-match status" >&2
	exit 1
fi

echo "Version validation is independent of go.work state."
echo "Version validation uses the tracked root Go declaration."
echo "Version validation fails when tracked-input discovery fails."
