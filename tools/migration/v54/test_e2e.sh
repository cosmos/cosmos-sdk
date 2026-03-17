#!/usr/bin/env bash
#
# End-to-end test for the v53 → v54 migration tool.
#
# Downloads the cosmos-sdk v0.53.6 simapp, runs the migration tool against it,
# builds the binary, initializes a chain, and attempts to start a node.
#
# Usage:
#   ./test_e2e.sh              # uses defaults
#   ./test_e2e.sh --keep       # keep work dir after run (for debugging)
#   ./test_e2e.sh --workdir /tmp/my-test   # use a specific work dir
#
# Output: a structured report at the end (and saved to a .log file) that can
# be copy-pasted to share with someone for debugging.

set -euo pipefail

# ── Configuration ────────────────────────────────────────────────────────────

# The commit on main that the migration tool targets (from module.go pseudo-version).
# This is used to clone the SDK source for local replace directives until v0.54.0 is tagged.
SDK_TARGET_COMMIT="2c527014f3ee"
COSMOS_SDK_TAG="${COSMOS_SDK_TAG:-main@${SDK_TARGET_COMMIT}}"
CHAIN_ID="migration-test-1"
MONIKER="test-node"
KEY_NAME="validator"
STAKE_DENOM="stake"
STAKE_AMOUNT="1000000000${STAKE_DENOM}"
NODE_STARTUP_TIMEOUT=30  # seconds to wait for node to produce blocks
BUILD_TAGS="app_v1"
MIN_GO_MAJOR=1
MIN_GO_MINOR=25

# ── Go version check ────────────────────────────────────────────────────────
# The v54 SDK and its dependency tree (CometBFT, OpenTelemetry, etc.) require
# Go >= 1.25.0. Check early so we fail fast with a clear message.

GO_VER_STR=$(go version 2>/dev/null | grep -oE 'go[0-9]+\.[0-9]+' | head -1 | sed 's/^go//')
GO_MAJ=$(echo "$GO_VER_STR" | cut -d. -f1)
GO_MIN=$(echo "$GO_VER_STR" | cut -d. -f2)

if [[ -z "$GO_MAJ" || -z "$GO_MIN" ]]; then
  echo "ERROR: Could not detect Go version. Is Go installed?" >&2
  exit 1
fi

if [[ "$GO_MAJ" -lt "$MIN_GO_MAJOR" ]] || \
   { [[ "$GO_MAJ" -eq "$MIN_GO_MAJOR" ]] && [[ "$GO_MIN" -lt "$MIN_GO_MINOR" ]]; }; then
  echo ""
  echo "═══════════════════════════════════════════════════════════"
  echo "  ERROR: Go >= ${MIN_GO_MAJOR}.${MIN_GO_MINOR}.0 is required"
  echo "═══════════════════════════════════════════════════════════"
  echo ""
  echo "  Installed: go${GO_VER_STR}"
  echo "  Required:  go${MIN_GO_MAJOR}.${MIN_GO_MINOR}.0+"
  echo ""
  echo "  The v54 cosmos-sdk and its dependencies (CometBFT v0.39+,"
  echo "  OpenTelemetry v1.42+, etc.) require Go 1.25 or later."
  echo ""
  echo "  To install Go 1.25:"
  echo "    go install golang.org/dl/go1.25.0@latest"
  echo "    go1.25.0 download"
  echo "    export PATH=\$(go1.25.0 env GOROOT)/bin:\$PATH"
  echo ""
  echo "  Or download from: https://go.dev/dl/"
  echo ""
  exit 1
fi

# ── Parse flags ──────────────────────────────────────────────────────────────

KEEP=false
LIVE=false
WORKDIR=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --keep)       KEEP=true; shift ;;
    --live)       LIVE=true; KEEP=true; shift ;;
    --workdir)    WORKDIR="$2"; shift 2 ;;
    *)            echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# ── Resolve paths ────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MIGRATION_TOOL_DIR="$SCRIPT_DIR"
CURATED_SIMAPP_FIXTURE="$(cd "$SCRIPT_DIR/../../../../simapp-v53/simapp" && pwd)"
COSMOS_SDK_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
# Allow callers to override the clone source, but default to the current repo so
# the harness works under `set -u` without extra environment.
COSMOS_SDK_REPO="${COSMOS_SDK_REPO:-$COSMOS_SDK_ROOT}"

if [[ -z "$WORKDIR" ]]; then
  WORKDIR="$(mktemp -d -t v54-migration-e2e.XXXXXX)"
fi
mkdir -p "$WORKDIR"

SIMAPP_DIR="$WORKDIR/simapp"
SDK_MAIN_DIR="$WORKDIR/cosmos-sdk-main"
SIMD_BIN="$WORKDIR/simd"
NODE_HOME="$WORKDIR/node-home"
LOG_FILE="$WORKDIR/e2e-report.log"

# ── Result tracking (bash 3 compatible — no associative arrays) ─────────────
#
# We store results in a flat file: "step_name|PASS|" or "step_name|FAIL|error msg"
# and step order in a simple indexed array.

RESULTS_FILE="$WORKDIR/.step_results"
: > "$RESULTS_FILE"

step_num=0
CURRENT_STEP=""

record_result() {
  # $1 = step name, $2 = PASS|FAIL, $3 = error detail (optional)
  echo "${1}|${2}|${3:-}" >> "$RESULTS_FILE"
}

get_result() {
  # $1 = step name -> prints PASS, FAIL, or SKIP
  local r
  r=$(grep "^${1}|" "$RESULTS_FILE" 2>/dev/null | tail -1 | cut -d'|' -f2 || true)
  echo "${r:-SKIP}"
}

get_error() {
  # $1 = step name -> prints error detail
  local e
  e=$(grep "^${1}|" "$RESULTS_FILE" 2>/dev/null | tail -1 | cut -d'|' -f3- || true)
  echo "${e:-}"
}

# ── Helpers ──────────────────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

step() {
  step_num=$((step_num + 1))
  CURRENT_STEP="$1"
  echo ""
  printf "${CYAN}${BOLD}[%d] %s${NC}\n" "$step_num" "$1"
  echo "────────────────────────────────────────"
}

pass() {
  record_result "$CURRENT_STEP" "PASS"
  printf "    ${GREEN}✓ %s${NC}\n" "$1"
}

fail() {
  record_result "$CURRENT_STEP" "FAIL" "$1"
  printf "    ${RED}✗ %s${NC}\n" "$1"
}

warn() {
  printf "    ${YELLOW}⚠ %s${NC}\n" "$1"
}

info() {
  printf "    %s\n" "$1"
}

cleanup() {
  # Kill any background simd process
  if [[ -n "${SIMD_PID:-}" ]] && kill -0 "$SIMD_PID" 2>/dev/null; then
    kill "$SIMD_PID" 2>/dev/null || true
    wait "$SIMD_PID" 2>/dev/null || true
  fi
  if [[ "$KEEP" == false ]]; then
    rm -rf "$WORKDIR"
  else
    # Even in --keep mode, the SDK main clone is large; offer to skip it
    info "Note: SDK main clone at $SDK_MAIN_DIR is ~1GB. Delete manually if not needed."
  fi
}
trap cleanup EXIT

# ── Step names (define upfront for report iteration) ─────────────────────────

S_CLONE="Copy curated v53 simapp fixture"
S_SNAPSHOT="Snapshot pre-migration state"
S_CLONE_MAIN="Clone SDK main for local replace"
S_MIGRATE="Run v54 migration tool"
S_GOIMPORTS="Run goimports"
S_TIDY="Run go mod tidy"
S_VERIFY="Verify migration artifacts"
S_BUILD="Build simd binary"
S_INIT="Initialize chain"
S_NODE="Start node (${NODE_STARTUP_TIMEOUT}s timeout)"

ALL_STEPS="$S_CLONE
$S_SNAPSHOT
$S_MIGRATE
$S_CLONE_MAIN
$S_GOIMPORTS
$S_TIDY
$S_VERIFY
$S_BUILD
$S_INIT
$S_NODE"

# ── Header ───────────────────────────────────────────────────────────────────

echo ""
printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
printf "${BOLD}  v53 → v54 Migration Tool — End-to-End Test${NC}\n"
printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
echo ""
info "Source fixture:  $CURATED_SIMAPP_FIXTURE"
info "Migration tool:  $MIGRATION_TOOL_DIR"
info "SDK clone src:   $COSMOS_SDK_REPO"
info "Work directory:  $WORKDIR"
info "Build tags:      $BUILD_TAGS"
info "Timestamp:       $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
echo ""

# ── Step 1: Copy supported fixture ──────────────────────────────────────────

step "$S_CLONE"

if [[ -d "$CURATED_SIMAPP_FIXTURE" ]]; then
  cp -r "$CURATED_SIMAPP_FIXTURE" "$SIMAPP_DIR"
  # The curated fixture tracks main in-repo, but this e2e case is meant to
  # exercise a real v0.53.6 -> v0.54 migration.
  (cd "$SIMAPP_DIR" && go mod edit -require=github.com/cosmos/cosmos-sdk@v0.53.6)
  pass "Copied fixture simapp ($(find "$SIMAPP_DIR" -name '*.go' | wc -l | tr -d ' ') Go files)"
else
  fail "fixture not found: $CURATED_SIMAPP_FIXTURE"
  exit 1
fi

# ── Step 2: Snapshot pre-migration state ────────────────────────────────────

step "$S_SNAPSHOT"

PRE_GO_MOD_SDK_VERSION="$(grep -E 'github.com/cosmos/cosmos-sdk ' "$SIMAPP_DIR/go.mod" | awk '{print $2}' | head -1)"
PRE_FILE_COUNT=$(find "$SIMAPP_DIR" -name '*.go' | wc -l | tr -d ' ')
info "SDK version in go.mod: $PRE_GO_MOD_SDK_VERSION"
info "Go files: $PRE_FILE_COUNT"

if [[ -f "$SIMAPP_DIR/ante.go" ]]; then
  info "ante.go: present (expected to be removed)"
fi

pass "Snapshot captured"

# ── Step 3: Run migration tool ──────────────────────────────────────────────

step "$S_MIGRATE"

MIGRATE_LOG="$WORKDIR/migration.log"
if (cd "$MIGRATION_TOOL_DIR" && go run . "$SIMAPP_DIR") >"$MIGRATE_LOG" 2>&1; then
  WARNINGS=$(grep -c '"level":"warn"' "$MIGRATE_LOG" 2>/dev/null || echo "0")
  REMOVALS=$(grep -c 'Removing' "$MIGRATE_LOG" 2>/dev/null || echo "0")
  IMPORTS=$(grep -c 'updated import' "$MIGRATE_LOG" 2>/dev/null || echo "0")
  SURGERIES=$(grep -c 'surgery' "$MIGRATE_LOG" 2>/dev/null || echo "0")
  TEXT_REPLACEMENTS=$(grep -c 'text replacement' "$MIGRATE_LOG" 2>/dev/null || echo "0")
  info "Warnings: $WARNINGS | Removals: $REMOVALS | Import rewrites: $IMPORTS"
  info "Arg surgeries: $SURGERIES | Text replacements: $TEXT_REPLACEMENTS"
  pass "Migration tool completed"
else
  fail "Migration tool failed (exit code $?)"
  info "Last 30 lines of output:"
  tail -30 "$MIGRATE_LOG"
fi

# Delete go.sum so go mod tidy regenerates it cleanly for the new versions
# (the migration tool stripped local replaces and updated versions, so the
# old go.sum is invalid)
rm -f "$SIMAPP_DIR/go.sum"
info "Deleted go.sum (will be regenerated by go mod tidy)"

# ── Step 3b: Clone SDK main for local replace ───────────────────────────────
# The migration tool pins a pseudo-version for the SDK that can't be resolved
# from the Go module proxy until v0.54.0 is tagged. We clone the SDK source at
# the target commit and add a local replace directive so go mod tidy works.

step "$S_CLONE_MAIN"

CLONE_MAIN_LOG="$WORKDIR/clone-main.log"
if git clone --filter=blob:none --single-branch --branch main "$COSMOS_SDK_REPO" "$SDK_MAIN_DIR" 2>"$CLONE_MAIN_LOG"; then
  # Check out the exact commit the migration targets
  if (cd "$SDK_MAIN_DIR" && git checkout "$SDK_TARGET_COMMIT") >>"$CLONE_MAIN_LOG" 2>&1; then
    info "Checked out SDK at commit $SDK_TARGET_COMMIT"
  else
    warn "Could not checkout $SDK_TARGET_COMMIT, using HEAD of main"
  fi

  # Add local replace directives for the core SDK AND all cosmossdk.io/* submodules.
  # The v54 SDK main branch migrated from cosmossdk.io/log to cosmossdk.io/log/v2,
  # which means ALL submodules (store, core, x/tx, x/bank, etc.) must come from the
  # same SDK main checkout so their interfaces are consistent. Without this, you get
  # type mismatches like: store/types.Context expects log.Logger but SDK has log/v2.Logger.
  #
  # This mirrors what the real simapp on main does — it has replace directives for
  # every submodule pointing to relative paths within the monorepo.
  #
  # We also cap each submodule's go directive to match our installed Go version, since
  # SDK main may declare a newer patch version (e.g., go 1.25.8 vs our 1.25.0).
  INSTALLED_GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1 | sed 's/^go//')

  # Cap the root SDK go directive first
  SDK_GO_VERSION=$(grep '^go ' "$SDK_MAIN_DIR/go.mod" | awk '{print $2}')
  if [[ -n "$SDK_GO_VERSION" && -n "$INSTALLED_GO_VERSION" && "$SDK_GO_VERSION" != "$INSTALLED_GO_VERSION" ]]; then
    info "Capping cloned SDK go directive from $SDK_GO_VERSION to $INSTALLED_GO_VERSION"
    (cd "$SDK_MAIN_DIR" && go mod edit -go="$INSTALLED_GO_VERSION") 2>/dev/null || true
    (cd "$SDK_MAIN_DIR" && go mod edit -toolchain=none) 2>/dev/null || true
  fi

  # Add replace for the root SDK module
  (cd "$SIMAPP_DIR" && go mod edit -replace "github.com/cosmos/cosmos-sdk=$SDK_MAIN_DIR") 2>>"$CLONE_MAIN_LOG"
  info "Added replace: github.com/cosmos/cosmos-sdk => $SDK_MAIN_DIR"

  # Discover all submodules in the SDK clone and add local replaces for each.
  # This finds every go.mod under $SDK_MAIN_DIR (excluding the root and simapp itself),
  # extracts the module path, and adds a replace directive + go directive capping.
  REPLACE_COUNT=0
  while IFS= read -r submod_gomod; do
    submod_dir=$(dirname "$submod_gomod")
    # Skip the root go.mod (already handled above)
    [[ "$submod_dir" == "$SDK_MAIN_DIR" ]] && continue
    # Skip simapp directories to avoid circular replaces
    [[ "$submod_dir" == *"/simapp"* ]] && continue
    # Skip tests, testutil, internal directories that aren't real modules
    [[ "$submod_dir" == *"/tests/"* ]] && continue

    # Extract the module path from go.mod
    mod_path=$(grep '^module ' "$submod_gomod" 2>/dev/null | awk '{print $2}')
    [[ -z "$mod_path" ]] && continue

    # Cap go directive if it differs from installed version
    sub_go=$(grep '^go ' "$submod_gomod" 2>/dev/null | awk '{print $2}')
    if [[ -n "$sub_go" && -n "$INSTALLED_GO_VERSION" && "$sub_go" != "$INSTALLED_GO_VERSION" ]]; then
      (cd "$submod_dir" && go mod edit -go="$INSTALLED_GO_VERSION") 2>/dev/null || true
      (cd "$submod_dir" && go mod edit -toolchain=none) 2>/dev/null || true
    fi

    # Add replace directive
    (cd "$SIMAPP_DIR" && go mod edit -replace "$mod_path=$submod_dir") 2>>"$CLONE_MAIN_LOG"
    REPLACE_COUNT=$((REPLACE_COUNT + 1))
  done < <(find "$SDK_MAIN_DIR" -name "go.mod" -type f 2>/dev/null | sort)

  info "Added $REPLACE_COUNT local replace directives for SDK submodules"

  pass "Cloned SDK main and added local replace directives"
else
  fail "Failed to clone SDK main — see $WORKDIR/clone-main.log"
  cat "$CLONE_MAIN_LOG" >&2
fi

# ── Step 4: Run goimports ───────────────────────────────────────────────────

step "$S_GOIMPORTS"

if ! command -v goimports &>/dev/null; then
  warn "goimports not found, installing..."
  go install golang.org/x/tools/cmd/goimports@latest 2>"$WORKDIR/goimports-install.log" || true
fi

GOIMPORTS_LOG="$WORKDIR/goimports.log"
if goimports -w "$SIMAPP_DIR" >"$GOIMPORTS_LOG" 2>&1; then
  pass "goimports succeeded"
else
  fail "goimports failed"
  cat "$GOIMPORTS_LOG"
fi

# ── Step 5: Run go mod tidy ────────────────────────────────────────────────

step "$S_TIDY"

# Ensure we have the installed Go version (may already be set from clone step)
if [[ -z "${INSTALLED_GO_VERSION:-}" ]]; then
  INSTALLED_GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+(\.[0-9]+)?' | head -1 | sed 's/^go//')
fi

# Pre-set simapp's go directive to our installed version BEFORE tidy.
# This prevents tidy from bumping it to a version we can't build with
# (e.g., cometbft v0.39.0-beta.3 requires go >= 1.25.0).
(cd "$SIMAPP_DIR" && go mod edit -go="$INSTALLED_GO_VERSION") 2>/dev/null || true
# Remove any toolchain directive that might conflict
(cd "$SIMAPP_DIR" && go mod edit -toolchain=none) 2>/dev/null || true

TIDY_LOG="$WORKDIR/tidy.log"
# Use -e flag to continue past errors (e.g., missing test-only packages
# like cosmossdk.io/api/cosmos/group/v1 that aren't needed for the build).
if (cd "$SIMAPP_DIR" && GOFLAGS="-mod=mod" go mod tidy -e) >"$TIDY_LOG" 2>&1; then
  POST_GO_MOD_SDK_VERSION=$(grep 'github.com/cosmos/cosmos-sdk ' "$SIMAPP_DIR/go.mod" | awk '{print $2}' || echo "unknown")
  info "SDK version after migration: $POST_GO_MOD_SDK_VERSION"

  # Force-cap go directive again — tidy may have bumped it despite our pre-set
  GO_MOD_VERSION=$(grep '^go ' "$SIMAPP_DIR/go.mod" | awk '{print $2}')
  if [[ -n "$GO_MOD_VERSION" && -n "$INSTALLED_GO_VERSION" ]]; then
    MOD_MAJOR=$(echo "$GO_MOD_VERSION" | cut -d. -f1)
    MOD_MINOR=$(echo "$GO_MOD_VERSION" | cut -d. -f2)
    INST_MAJOR=$(echo "$INSTALLED_GO_VERSION" | cut -d. -f1)
    INST_MINOR=$(echo "$INSTALLED_GO_VERSION" | cut -d. -f2)
    if [[ "$MOD_MAJOR" -gt "$INST_MAJOR" ]] || \
       { [[ "$MOD_MAJOR" -eq "$INST_MAJOR" ]] && [[ "$MOD_MINOR" -gt "$INST_MINOR" ]]; }; then
      warn "go.mod directive ($GO_MOD_VERSION) exceeds installed Go ($INSTALLED_GO_VERSION) — capping"
      (cd "$SIMAPP_DIR" && go mod edit -go="$INSTALLED_GO_VERSION") 2>/dev/null || true
    fi
  fi
  # Strip toolchain directive if tidy added one
  (cd "$SIMAPP_DIR" && go mod edit -toolchain=none) 2>/dev/null || true

  pass "go mod tidy succeeded"
else
  fail "go mod tidy failed"
  info "Last 40 lines:"
  tail -40 "$TIDY_LOG"
fi

# ── Step 6: Verify migration artifacts ──────────────────────────────────────

step "$S_VERIFY"

VERIFY_PASS=true

# ante.go should be removed
if [[ -f "$SIMAPP_DIR/ante.go" ]]; then
  fail "ante.go still exists (should have been removed)"
  VERIFY_PASS=false
fi

# Check that circuit/nft/group keepers are removed from app.go
if grep -q 'CircuitKeeper' "$SIMAPP_DIR/app.go" 2>/dev/null; then
  warn "CircuitKeeper still referenced in app.go"
  VERIFY_PASS=false
fi
if grep -q 'NFTKeeper' "$SIMAPP_DIR/app.go" 2>/dev/null; then
  warn "NFTKeeper still referenced in app.go"
  VERIFY_PASS=false
fi

# Check log/v2 import
if grep -q 'cosmossdk.io/log/v2' "$SIMAPP_DIR/app.go" 2>/dev/null; then
  info "log/v2 import: present ✓"
else
  warn "log/v2 import not found in app.go"
  VERIFY_PASS=false
fi

# Check go.mod for removed vanity modules
for mod in "cosmossdk.io/x/circuit" "cosmossdk.io/x/nft" "cosmossdk.io/x/evidence" "cosmossdk.io/x/upgrade"; do
  if grep -q "^[[:space:]]*$mod " "$SIMAPP_DIR/go.mod" 2>/dev/null; then
    warn "$mod still in go.mod (should be removed)"
    VERIFY_PASS=false
  fi
done

if [[ "$VERIFY_PASS" == true ]]; then
  pass "All migration artifacts look correct"
fi

# ── Step 7: Build ───────────────────────────────────────────────────────────

step "$S_BUILD"

BUILD_LOG="$WORKDIR/build.log"
BUILD_START=$(date +%s)
if (cd "$SIMAPP_DIR" && go build -tags "$BUILD_TAGS" -o "$SIMD_BIN" ./simd/) >"$BUILD_LOG" 2>&1; then
  BUILD_END=$(date +%s)
  BUILD_SECS=$((BUILD_END - BUILD_START))
  BINARY_SIZE=$(du -h "$SIMD_BIN" | cut -f1)
  info "Binary size: $BINARY_SIZE | Build time: ${BUILD_SECS}s"
  GO_VERSION=$("$SIMD_BIN" version 2>/dev/null || echo "unknown")
  info "simd version: $GO_VERSION"
  pass "Build succeeded"
else
  BUILD_END=$(date +%s)
  BUILD_SECS=$((BUILD_END - BUILD_START))
  fail "Build failed after ${BUILD_SECS}s"
  info ""
  info "=== BUILD ERRORS ==="
  cat "$BUILD_LOG"
  info "=== END BUILD ERRORS ==="
fi

# ── Step 8: Init chain ─────────────────────────────────────────────────────

step "$S_INIT"

INIT_LOG="$WORKDIR/init.log"
if [[ -f "$SIMD_BIN" ]]; then
  INIT_OK=true

  # 8a. Init the chain
  if ! "$SIMD_BIN" init "$MONIKER" --chain-id "$CHAIN_ID" --home "$NODE_HOME" >>"$INIT_LOG" 2>&1; then
    fail "simd init failed"
    tail -20 "$INIT_LOG"
    INIT_OK=false
  fi

  # 8b. Create a validator key (test keyring — no password)
  if [[ "$INIT_OK" == true ]]; then
    if ! "$SIMD_BIN" keys add "$KEY_NAME" --keyring-backend test --home "$NODE_HOME" >>"$INIT_LOG" 2>&1; then
      fail "keys add failed"
      tail -10 "$INIT_LOG"
      INIT_OK=false
    fi
  fi

  # 8c. Add genesis account with tokens
  if [[ "$INIT_OK" == true ]]; then
    VALIDATOR_ADDR=$("$SIMD_BIN" keys show "$KEY_NAME" --keyring-backend test --home "$NODE_HOME" -a 2>/dev/null)
    if ! "$SIMD_BIN" genesis add-genesis-account "$VALIDATOR_ADDR" "2000000000${STAKE_DENOM}" \
        --keyring-backend test --home "$NODE_HOME" >>"$INIT_LOG" 2>&1; then
      fail "add-genesis-account failed"
      tail -10 "$INIT_LOG"
      INIT_OK=false
    fi
  fi

  # 8d. Create genesis transaction (validator self-delegation)
  if [[ "$INIT_OK" == true ]]; then
    if ! "$SIMD_BIN" genesis gentx "$KEY_NAME" "$STAKE_AMOUNT" \
        --chain-id "$CHAIN_ID" --keyring-backend test --home "$NODE_HOME" >>"$INIT_LOG" 2>&1; then
      fail "gentx failed"
      tail -10 "$INIT_LOG"
      INIT_OK=false
    fi
  fi

  # 8e. Collect genesis transactions
  if [[ "$INIT_OK" == true ]]; then
    if ! "$SIMD_BIN" genesis collect-gentxs --home "$NODE_HOME" >>"$INIT_LOG" 2>&1; then
      fail "collect-gentxs failed"
      tail -10 "$INIT_LOG"
      INIT_OK=false
    fi
  fi

  if [[ "$INIT_OK" == true ]]; then
    pass "Chain initialized with validator (chain-id=$CHAIN_ID)"
  fi
else
  fail "Skipped — no binary (build failed)"
fi

# ── Step 9: Start node ─────────────────────────────────────────────────────

step "$S_NODE"

# Disable errexit for the node test — this section handles all errors explicitly.
# set -e inside a polling loop with kill/wait/grep is fragile and causes silent exits.
set +e

NODE_LOG="$WORKDIR/node.log"
if [[ -f "$SIMD_BIN" && -d "$NODE_HOME" ]]; then
  # Start node in background to verify it produces blocks
  "$SIMD_BIN" start --home "$NODE_HOME" >"$NODE_LOG" 2>&1 &
  SIMD_PID=$!
  info "Started simd (PID $SIMD_PID)"

  # Wait for block production or failure
  NODE_OK=false
  for i in $(seq 1 "$NODE_STARTUP_TIMEOUT"); do
    sleep 1
    if ! kill -0 "$SIMD_PID" 2>/dev/null; then
      # Process died
      wait "$SIMD_PID" 2>/dev/null
      fail "Node exited prematurely"
      info ""
      info "=== LAST 50 LINES OF NODE LOG ==="
      tail -50 "$NODE_LOG"
      info "=== END NODE LOG ==="
      break
    fi
    # Check for block commits (various CometBFT log formats across versions)
    if grep -q 'committed state' "$NODE_LOG" 2>/dev/null \
       || grep -q 'finalized block' "$NODE_LOG" 2>/dev/null \
       || grep -q 'indexed block' "$NODE_LOG" 2>/dev/null \
       || grep -q 'received proposal' "$NODE_LOG" 2>/dev/null \
       || grep -q 'executed block' "$NODE_LOG" 2>/dev/null \
       || grep -q 'block{' "$NODE_LOG" 2>/dev/null; then
      BLOCK_HEIGHT=$(grep -oE 'height=[0-9]+' "$NODE_LOG" | tail -1 | cut -d= -f2)
      pass "Node producing blocks (reached height ${BLOCK_HEIGHT:-unknown})"
      NODE_OK=true
      break
    fi
  done

  if [[ "$NODE_OK" == false ]] && kill -0 "$SIMD_PID" 2>/dev/null; then
    fail "Timed out after ${NODE_STARTUP_TIMEOUT}s without seeing blocks"
    info ""
    info "=== LAST 50 LINES OF NODE LOG ==="
    tail -50 "$NODE_LOG"
    info "=== END NODE LOG ==="
  elif [[ "$NODE_OK" == false ]]; then
    fail "Node failed (see logs above)"
  fi

  # Shutdown the background node
  if kill -0 "$SIMD_PID" 2>/dev/null; then
    kill "$SIMD_PID" 2>/dev/null
    wait "$SIMD_PID" 2>/dev/null
    info "Node shut down cleanly"
  fi
else
  fail "Skipped — no binary or init failed"
fi

set -e

# ── Report ──────────────────────────────────────────────────────────────────

echo ""
printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
printf "${BOLD}  REPORT${NC}\n"
printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
echo ""

# Generate the plain-text report (for sharing)
# Disable errexit inside the tee subshell — grep/awk returning no-match (exit 1)
# combined with pipefail would silently kill the subshell.
set +e
{
  echo "# v54 Migration E2E Test Report"
  echo ""
  echo "Date:       $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "Source:     $COSMOS_SDK_TAG"
  echo "Go:         $(go version 2>/dev/null || echo 'unknown')"
  echo "OS/Arch:    $(uname -s)/$(uname -m)"
  echo "Work dir:   $WORKDIR"
  echo ""
  echo "## Results"
  echo ""

  all_pass=true
  echo "$ALL_STEPS" | while IFS= read -r step_name; do
    result=$(get_result "$step_name")
    if [[ "$result" == "PASS" ]]; then
      echo "  ✓ PASS  $step_name"
    elif [[ "$result" == "FAIL" ]]; then
      echo "  ✗ FAIL  $step_name"
      err=$(get_error "$step_name")
      if [[ -n "$err" ]]; then
        echo "          → $err"
      fi
    else
      echo "  - SKIP  $step_name"
    fi
  done

  # Check overall pass/fail
  if grep -q '|FAIL|' "$RESULTS_FILE" 2>/dev/null; then
    echo ""
    echo "## Overall: SOME FAILURES ✗"
  else
    echo ""
    echo "## Overall: ALL PASS ✓"
  fi

  echo ""
  echo "## Logs"
  echo ""

  if [[ -f "$MIGRATE_LOG" ]]; then
    echo "### Migration tool output"
    echo '```'
    cat "$MIGRATE_LOG"
    echo '```'
    echo ""
  fi

  if [[ -f "${GOIMPORTS_LOG:-}" ]] && [[ -s "${GOIMPORTS_LOG:-}" ]]; then
    echo "### goimports output"
    echo '```'
    cat "$GOIMPORTS_LOG"
    echo '```'
    echo ""
  fi

  if [[ -f "${TIDY_LOG:-}" ]] && [[ -s "${TIDY_LOG:-}" ]]; then
    echo "### go mod tidy output"
    echo '```'
    cat "$TIDY_LOG"
    echo '```'
    echo ""
  fi

  if [[ -f "${BUILD_LOG:-}" ]] && [[ -s "${BUILD_LOG:-}" ]]; then
    echo "### Build output"
    echo '```'
    cat "$BUILD_LOG"
    echo '```'
    echo ""
  fi

  if [[ -f "${NODE_LOG:-}" ]] && [[ -s "${NODE_LOG:-}" ]]; then
    echo "### Node log (last 80 lines)"
    echo '```'
    tail -80 "$NODE_LOG"
    echo '```'
    echo ""
  fi

  # ── Post-migration diagnostics ──────────────────────────────────────────
  echo "## Post-Migration Diagnostics"
  echo ""

  # File inventory
  echo "### File inventory after migration"
  echo '```'
  if [[ -d "$SIMAPP_DIR" ]]; then
    echo "Go files remaining:"
    find "$SIMAPP_DIR" -name '*.go' | sort | while IFS= read -r f; do
      echo "  $(echo "$f" | sed "s|$SIMAPP_DIR/||")"
    done
    echo ""
    echo "Removed files (expected: ante.go, app_di.go, root_di.go):"
    for expected_removed in ante.go app_di.go "simd/cmd/root_di.go"; do
      if [[ ! -f "$SIMAPP_DIR/$expected_removed" ]]; then
        echo "  ✓ $expected_removed (removed)"
      else
        echo "  ✗ $expected_removed (STILL EXISTS)"
      fi
    done
  fi
  echo '```'
  echo ""

  # Import blocks for key files
  echo "### Import blocks (post-migration)"
  for key_file in app.go app_config.go app_test.go; do
    if [[ -f "$SIMAPP_DIR/$key_file" ]]; then
      echo "#### $key_file imports"
      echo '```go'
      # Extract the import block(s) — from "import (" to the matching ")"
      awk '/^import \(/{p=1} p{print} /^\)/{if(p) p=0}' "$SIMAPP_DIR/$key_file"
      echo '```'
      echo ""
    fi
  done

  # Remaining x/group references (the main failure pattern)
  echo "### Remaining x/group references"
  echo '```'
  GROUP_REFS=$(grep -rn 'x/group\|cosmos/group' "$SIMAPP_DIR" --include='*.go' 2>/dev/null || true)
  if [[ -n "$GROUP_REFS" ]]; then
    echo "$GROUP_REFS" | sed "s|$SIMAPP_DIR/||"
  else
    echo "(none — all x/group references removed)"
  fi
  echo '```'
  echo ""

  # Remaining circuit/nft references
  echo "### Remaining circuit/nft references"
  echo '```'
  CN_REFS=$(grep -rn 'x/circuit\|x/nft\|contrib/x/circuit\|contrib/x/nft\|circuitkeeper\|nftkeeper\|CircuitKeeper\|NFTKeeper' "$SIMAPP_DIR" --include='*.go' 2>/dev/null || true)
  if [[ -n "$CN_REFS" ]]; then
    echo "$CN_REFS" | sed "s|$SIMAPP_DIR/||"
  else
    echo "(none — all circuit/nft references removed)"
  fi
  echo '```'
  echo ""

  # go.mod contents
  echo "### go.mod (post-migration)"
  echo '```'
  if [[ -f "$SIMAPP_DIR/go.mod" ]]; then
    # Show the module line, go directive, require block (direct deps only), and replace block
    echo "--- module & go directive ---"
    head -5 "$SIMAPP_DIR/go.mod"
    echo ""
    echo "--- require (direct) ---"
    awk '/^require \(/{p=1;next} /^\)/{if(p){p=0;next}} p && !/\/\/ indirect/' "$SIMAPP_DIR/go.mod"
    echo ""
    echo "--- replace directives ---"
    grep '^replace ' "$SIMAPP_DIR/go.mod" 2>/dev/null || echo "(none in single-line form)"
    awk '/^replace \(/{p=1;next} /^\)/{if(p){p=0;next}} p' "$SIMAPP_DIR/go.mod" 2>/dev/null
  fi
  echo '```'
  echo ""

  echo "## Diff: go.mod changes"
  echo '```'
  echo "Before: ${PRE_GO_MOD_SDK_VERSION:-unknown}"
  echo "After:  ${POST_GO_MOD_SDK_VERSION:-unknown}"
  echo '```'

} | tee "$LOG_FILE"
set -e

echo ""
printf "${BOLD}Report saved to: ${CYAN}file://%s${NC}\n" "$LOG_FILE"

if [[ "$KEEP" == true ]]; then
  printf "${BOLD}Work dir kept at: ${CYAN}%s${NC}\n" "$WORKDIR"
  echo "  Simapp:  $SIMAPP_DIR"
  echo "  Binary:  $SIMD_BIN"
  echo "  Logs:    $WORKDIR/*.log"
fi

echo ""

# Exit with failure if any step failed (unless --live, which continues)
if grep -q '|FAIL|' "$RESULTS_FILE" 2>/dev/null; then
  if [[ "$LIVE" == true ]]; then
    echo ""
    printf "${RED}${BOLD}Some steps failed — skipping live node.${NC}\n"
  fi
  exit 1
fi

# ── Live mode: restart node in foreground ─────────────────────────────────
if [[ "$LIVE" == true && -f "$SIMD_BIN" && -d "$NODE_HOME" ]]; then
  echo ""
  printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
  printf "${GREEN}${BOLD}  All tests passed — starting node in live mode${NC}\n"
  printf "${BOLD}═══════════════════════════════════════════════════════════${NC}\n"
  echo ""
  echo "  RPC:   http://localhost:26657"
  echo "  gRPC:  localhost:9090"
  echo "  Home:  $NODE_HOME"
  echo ""
  echo "  Press Ctrl+C to stop."
  echo ""
  # Disable the cleanup trap — the user will Ctrl+C the node directly.
  # The work dir is kept (--live implies --keep).
  trap - EXIT
  # Restart with existing home dir. CometBFT resumes from the last committed block.
  exec "$SIMD_BIN" start --home "$NODE_HOME"
fi

exit 0
