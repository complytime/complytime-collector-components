#!/usr/bin/env bash
# Helper script to run commands for each Go module in the monorepo
# Usage: go-module-runner.sh <command>
#
# Module Discovery Strategy:
# - deps/deps-update: Auto-discovers ALL modules (app + tests) for dependency management
# - test/lint/codegen: Uses MODULES var (app modules only) - tests/integration has separate task
set -euo pipefail

MODULES="${MODULES:-./proofwatch}"
GAZE_COVERPROFILE="${GAZE_COVERPROFILE:-coverage.out}"
GAZE_NEW_FUNC_THRESHOLD="${GAZE_NEW_FUNC_THRESHOLD:-30}"

case "$1" in
deps)
	# Auto-discover all modules (any directory with go.mod, excluding vendor)
	ALL_MODULES=$(find . -type f -name 'go.mod' -not -path '*/vendor/*' -exec dirname {} \; | sort)

	for m in $ALL_MODULES; do
		echo "Processing deps for $m..."
		(cd "$m" && go mod tidy && go mod verify && go mod download) || {
			echo "Deps failed for module: $m"
			exit 1
		}
		echo "-------------------"
	done
	echo "--- Deps completed for all modules ---"
	;;

deps-update)
	echo "========================================================================================================="
	echo "Updating all dependencies across workspace (auto-discovering modules)..."
	echo "========================================================================================================="

	# Auto-discover all modules (any directory with go.mod, excluding vendor)
	ALL_MODULES=$(find . -type f -name 'go.mod' -not -path '*/vendor/*' -exec dirname {} \; | sort)

	for m in $ALL_MODULES; do
		echo "Updating dependencies for $m..."

		# Root module (.) only has tool dependencies, no packages - update tools only
		if [[ "$m" == "." ]]; then
			echo "  Root module - updating tool dependencies only..."
			# Extract tool directives and update them individually
			while read -r tool_path; do
				if [[ -n "$tool_path" ]]; then
					echo "  Updating tool: $tool_path"
					(cd "$m" && GOTOOLCHAIN=auto go get -u "$tool_path") || {
						echo "Tool update failed: $tool_path"
						exit 1
					}
				fi
			done < <(grep '^tool ' go.mod | awk '{print $2}' || true)
		else
			# Update non-OTel dependencies only
			# OTel versions are managed via version:sync to ensure alignment with contrib package availability
			echo "  Updating non-OTel dependencies (OTel packages managed separately)..."

			# Get list of direct dependencies (non-indirect, non-OTel)
			DEPS_TO_UPDATE=$(cd "$m" && go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all 2>/dev/null |
				grep -v '^$' |
				grep -v 'go.opentelemetry.io/collector' |
				grep -v "^$(go list -m)$" || true)

			if [[ -n "$DEPS_TO_UPDATE" ]]; then
				echo "$DEPS_TO_UPDATE" | while read -r dep; do
					[[ -z "$dep" ]] && continue
					(cd "$m" && GOTOOLCHAIN=auto go get -u "$dep") || echo "  Warning: Failed to update $dep (continuing)"
				done
			else
				echo "  No non-OTel dependencies to update"
			fi
		fi

		echo "Tidying $m..."
		(cd "$m" && GOTOOLCHAIN=auto go mod tidy) || {
			echo "Tidy failed for module: $m"
			exit 1
		}
		echo "-------------------"
	done

	echo "Syncing workspace..."
	go work sync || {
		echo "Workspace sync failed"
		exit 1
	}

	echo "Verifying all modules..."
	for m in $ALL_MODULES; do
		(cd "$m" && go mod verify) || {
			echo "Verify failed for module: $m"
			exit 1
		}
	done

	echo ""
	echo "--- All dependencies updated successfully! ---"
	echo ""
	echo "Running version:sync to propagate versions across the project..."
	bash "$(dirname "$0")/version-sync.sh" || {
		echo "Version sync failed"
		exit 1
	}
	;;

test)
	for m in $MODULES; do
		echo "========================================================================================================="
		echo "Running tests for $m..."
		echo "========================================================================================================="
		(cd "$m" && GOWORK=off go test -v -coverprofile=coverage.out -covermode=atomic ./...) || {
			echo "Tests failed for module: $m"
			exit 1
		}
		echo "Coverage summary for $m:"
		(cd "$m" && GOWORK=off go tool cover -func=coverage.out | tail -n1) || true
		echo "-------------------"
	done
	echo "--- All tests passed! ---"
	;;

test-race)
	for m in $MODULES; do
		echo "Running tests with race detection for $m..."
		(cd "$m" && GOWORK=off go test -v -race ./...) || {
			echo "Tests failed for module: $m"
			exit 1
		}
	done
	echo "--- All tests passed with race detection! ---"
	;;

coverage-report)
	for m in $MODULES; do
		echo "Generating coverage report for $m..."
		(cd "$m" && GOWORK=off go tool cover -html=coverage.out -o coverage.html)
		echo "Coverage summary for $m:"
		(cd "$m" && GOWORK=off go tool cover -func=coverage.out | tail -n1) || true
		echo "-------------------"
	done
	echo "--- Coverage reports generated! ---"
	;;

lint)
	# Auto-discover all modules (any directory with go.mod, excluding vendor)
	ALL_MODULES=$(find . -type f -name 'go.mod' -not -path '*/vendor/*' -exec dirname {} \; | sort)

	for m in $ALL_MODULES; do
		# Skip modules with no .go files (e.g., root module with only tool directives)
		if ! find "$m" -maxdepth 1 -name '*.go' -print -quit | grep -q .; then
			echo "Skipping $m (no .go files)..."
			continue
		fi

		echo "Running golangci-lint for $m..."

		# Determine config path relative to module directory
		if [[ "$m" == "." ]]; then
			CONFIG_PATH=".golangci.yml"
		else
			# Count directory depth to build relative path to root config
			DEPTH=$(echo "$m" | tr -cd '/' | wc -c || true)
			CONFIG_PATH=""
			for ((i = 0; i < DEPTH; i++)); do CONFIG_PATH="../${CONFIG_PATH}"; done
			CONFIG_PATH="${CONFIG_PATH}.golangci.yml"
		fi

		(cd "$m" && GOWORK=off golangci-lint run --config "$CONFIG_PATH" ./...) || {
			echo "Linting failed for module: $m"
			exit 1
		}
	done
	echo "--- All linting passed! ---"
	;;

crapload)
	for m in $MODULES; do
		echo "========================================================================================================="
		echo "CRAP analysis for $m..."
		echo "========================================================================================================="
		(cd "$m" && go tool gaze crap --format=text --coverprofile="$GAZE_COVERPROFILE" ./...)
	done
	;;

crapload-baseline)
	for m in $MODULES; do
		echo "Generating baseline for $m..."
		mkdir -p "$m/.gaze"
		MODULE_ROOT=$(cd "$m" && pwd)
		(cd "$m" && go tool gaze crap --format=json --coverprofile="$GAZE_COVERPROFILE" ./... 2>/dev/null |
			jq --arg root "$MODULE_ROOT/" '(.scores[],.summary.worst_crap[]?,.summary.worst_gaze_crap[]?) |= (.file |= ltrimstr($root))' >.gaze/baseline.json)
		echo "Baseline written to $m/.gaze/baseline.json"
	done
	;;

crapload-check)
	TOTAL_REGRESSIONS=0
	for m in $MODULES; do
		echo "========================================================================================================="
		echo "Checking CRAP regressions for $m..."
		echo "========================================================================================================="
		BASELINE="$m/.gaze/baseline.json"
		if [[ ! -f "$BASELINE" ]]; then
			echo "ERROR: Baseline file $BASELINE not found. Run 'task quality:crapload-baseline' first."
			exit 1
		fi
		MODULE_ROOT=$(cd "$m" && pwd)
		(cd "$m" && go tool gaze crap --format=json --coverprofile="$GAZE_COVERPROFILE" ./... 2>/dev/null |
			jq --arg root "$MODULE_ROOT/" '(.scores[],.summary.worst_crap[]?,.summary.worst_gaze_crap[]?) |= (.file |= ltrimstr($root))' >/tmp/crapload-current.json)
		echo "Comparing against baseline..."
		jq -r '.scores[] | "\(.file):\(.function)\t\(.crap)\t\(.gaze_crap // 0)"' "$BASELINE" | sort >/tmp/crapload-baseline.tsv
		REGRESSIONS=0
		while IFS=$'\t' read -r func crap gaze_crap; do
			baseline_line=$(grep -F "$func	" /tmp/crapload-baseline.tsv | head -1 || true)
			if [[ -z "$baseline_line" ]]; then
				if [[ "$(echo "$crap > $GAZE_NEW_FUNC_THRESHOLD" | bc -l || true)" = "1" ]]; then
					echo "NEW FUNCTION VIOLATION: $func CRAP=$crap (threshold=$GAZE_NEW_FUNC_THRESHOLD)"
					REGRESSIONS=$((REGRESSIONS + 1))
				fi
			else
				b_crap=$(echo "$baseline_line" | cut -f2)
				b_gaze=$(echo "$baseline_line" | cut -f3)
				if [[ "$(echo "$crap > $b_crap" | bc -l || true)" = "1" ]]; then
					echo "REGRESSION: $func CRAP $b_crap -> $crap"
					REGRESSIONS=$((REGRESSIONS + 1))
				fi
				if [[ "$(echo "$gaze_crap > $b_gaze" | bc -l || true)" = "1" ]]; then
					echo "REGRESSION: $func GazeCRAP $b_gaze -> $gaze_crap"
					REGRESSIONS=$((REGRESSIONS + 1))
				fi
			fi
		done < <(jq -r '.scores[] | "\(.file):\(.function)\t\(.crap)\t\(.gaze_crap // 0)"' /tmp/crapload-current.json | sort || true)
		TOTAL_REGRESSIONS=$((TOTAL_REGRESSIONS + REGRESSIONS))
		if [[ "$REGRESSIONS" -gt 0 ]]; then
			echo "$m: $REGRESSIONS regression(s) detected"
		else
			echo "$m: No regressions detected"
		fi
	done
	if [[ "$TOTAL_REGRESSIONS" -gt 0 ]]; then
		echo "FAIL: $TOTAL_REGRESSIONS total regression(s) detected"
		exit 1
	else
		echo "PASS: No regressions detected across all modules"
	fi
	;;

*)
	echo "Unknown command: $1"
	echo "Available commands: deps, deps-update, test, test-race, coverage-report, lint, crapload, crapload-baseline, crapload-check"
	exit 1
	;;
esac
