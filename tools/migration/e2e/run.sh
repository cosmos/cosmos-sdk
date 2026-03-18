#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# E2E Migration Test Runner
#
# Clones chain repos defined in chains.yaml, runs the v54 migration tool,
# and checks that the result compiles.
#
# Usage:
#   ./run.sh                     # run all non-skipped chains
#   ./run.sh simapp-v53          # run a specific chain by ID
#   ./run.sh simapp-v53 dydx-v4  # run multiple chains
#
# Requirements:
#   - go (1.23+)
#   - git
#   - python3 + pyyaml (for parsing chains.yaml)
#   - The migration tool must be built first (see below)
#
# Environment:
#   E2E_WORKDIR     Override the temp directory for cloned repos (default: /tmp/migration-e2e)
#   E2E_KEEP        If set to "1", don't delete workdir after run (useful for debugging)
#   E2E_SKIP_BUILD  If set to "1", skip `go build` even for chains that expect it
#   MIGRATE_BIN     Path to pre-built migration binary (default: builds from ../v54/)
# ──────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MIGRATION_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
SPEC_DIR="$MIGRATION_DIR/migration-spec/v50-to-v54"
CHAINS_FILE="$SCRIPT_DIR/chains.yaml"

WORKDIR="${E2E_WORKDIR:-/tmp/migration-e2e}"
KEEP="${E2E_KEEP:-0}"
SKIP_BUILD="${E2E_SKIP_BUILD:-0}"
MIGRATE_BIN="${MIGRATE_BIN:-}"

# ── colours ───────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'  # no colour

log()  { echo -e "${CYAN}[e2e]${NC} $*"; }
pass() { echo -e "${GREEN}  ✓${NC} $*"; }
fail() { echo -e "${RED}  ✗${NC} $*"; }
warn() { echo -e "${YELLOW}  ⚠${NC} $*"; }

# ── pre-flight checks ────────────────────────────────────────────────────────
check_deps() {
    local missing=0
    for cmd in go git python3; do
        if ! command -v "$cmd" &>/dev/null; then
            fail "Required command not found: $cmd"
            missing=1
        fi
    done
    python3 -c "import yaml" 2>/dev/null || {
        fail "Python pyyaml not installed (pip install pyyaml)"
        missing=1
    }
    if [[ $missing -eq 1 ]]; then
        exit 1
    fi
}

# ── build the migration tool ─────────────────────────────────────────────────
build_migrate_bin() {
    if [[ -n "$MIGRATE_BIN" && -x "$MIGRATE_BIN" ]]; then
        log "Using pre-built migration binary: $MIGRATE_BIN"
        return
    fi
    log "Building migration tool from $MIGRATION_DIR/v54/ ..."
    MIGRATE_BIN="$WORKDIR/migrate-v54"
    (cd "$MIGRATION_DIR/v54" && go build -o "$MIGRATE_BIN" .) || {
        fail "Failed to build migration tool"
        exit 1
    }
    pass "Migration tool built: $MIGRATE_BIN"
}

# ── parse chains.yaml ────────────────────────────────────────────────────────
# Outputs one JSON object per chain to stdout, filtered by requested IDs.
parse_chains() {
    local filter_ids=("$@")
    python3 - "$CHAINS_FILE" "${filter_ids[@]}" <<'PYEOF'
import yaml, json, sys

chains_file = sys.argv[1]
filter_ids = sys.argv[2:] if len(sys.argv) > 2 else []

with open(chains_file) as f:
    data = yaml.safe_load(f)

for chain in data.get("chains", []):
    if filter_ids and chain["id"] not in filter_ids:
        continue
    if chain.get("skip", False) and not filter_ids:
        # Skip unless explicitly requested
        continue
    print(json.dumps(chain))
PYEOF
}

# ── clone a chain ─────────────────────────────────────────────────────────────
clone_chain() {
    local repo="$1" ref="$2" dest="$3"
    if [[ -d "$dest/.git" ]]; then
        log "  Reusing existing clone at $dest"
        (cd "$dest" && git checkout -f "$ref" 2>/dev/null) || \
        (cd "$dest" && git fetch origin "$ref" && git checkout -f FETCH_HEAD)
        return
    fi
    git clone --depth 1 --branch "$ref" "$repo" "$dest" 2>&1 | tail -1
}

# ── run verification checks from specs ────────────────────────────────────────
run_verify() {
    local dir="$1"
    local errors=0

    log "  Running spec verification checks..."
    for spec_file in "$SPEC_DIR"/*.yaml; do
        python3 - "$spec_file" "$dir" <<'PYEOF' || errors=$((errors + 1))
import yaml, sys, os, subprocess

spec_file = sys.argv[1]
search_dir = sys.argv[2]

with open(spec_file) as f:
    spec = yaml.safe_load(f)

spec_id = spec.get("id", os.path.basename(spec_file))
verification = spec.get("verification", {})
failed = False

def grep_dir(pattern, directory):
    """Search for pattern in .go files, return list of matching files."""
    try:
        result = subprocess.run(
            ["grep", "-rl", "--include=*.go", pattern, directory],
            capture_output=True, text=True, timeout=30
        )
        return [f for f in result.stdout.strip().split("\n") if f]
    except Exception:
        return []

# must_not_import: these import prefixes must not appear in any .go file
for imp in verification.get("must_not_import", []):
    matches = grep_dir(f'"{imp}', search_dir)
    if matches:
        print(f"  FAIL [{spec_id}] must_not_import '{imp}' found in: {', '.join(matches[:3])}")
        failed = True

# must_not_contain: these patterns must not appear
for entry in verification.get("must_not_contain", []):
    pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
    file_match = entry.get("file_match", "") if isinstance(entry, dict) else ""
    if not pattern:
        continue
    matches = grep_dir(pattern, search_dir)
    if file_match:
        matches = [m for m in matches if m.endswith(file_match)]
    if matches:
        print(f"  FAIL [{spec_id}] must_not_contain '{pattern}' found in: {', '.join(matches[:3])}")
        failed = True

# must_contain: these patterns must appear somewhere
for entry in verification.get("must_contain", []):
    pattern = entry if isinstance(entry, str) else entry.get("pattern", "")
    if not pattern:
        continue
    matches = grep_dir(pattern, search_dir)
    if not matches:
        print(f"  FAIL [{spec_id}] must_contain '{pattern}' not found anywhere")
        failed = True

if failed:
    sys.exit(1)
else:
    print(f"  PASS [{spec_id}]")
PYEOF
    done

    return $errors
}

# ── run one chain ─────────────────────────────────────────────────────────────
run_chain() {
    local chain_json="$1"
    local id repo ref app_dir migrate_dir expect_build expect_fatal
    id=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['id'])")
    repo=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['repo'])")
    ref=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d['ref'])")
    app_dir=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('app_dir','.'))")
    migrate_dir=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('migrate_dir','.'))")
    expect_build=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(str(d.get('expected',{}).get('build',True)).lower())")
    expect_fatal=$(echo "$chain_json" | python3 -c "import json,sys; d=json.load(sys.stdin); print(d.get('expected',{}).get('fatal','') or '')")

    local chain_dir="$WORKDIR/$id"
    local result=0

    echo ""
    log "━━━ $id ($ref) ━━━"

    # Clone
    log "  Cloning $repo @ $ref ..."
    clone_chain "$repo" "$ref" "$chain_dir" || {
        fail "Clone failed for $id"
        return 1
    }
    pass "Cloned"

    local target_dir="$chain_dir"
    [[ "$migrate_dir" != "." ]] && target_dir="$chain_dir/$migrate_dir"

    # Run migration
    log "  Running migration on $target_dir ..."
    local migrate_output
    migrate_output=$("$MIGRATE_BIN" --dir "$target_dir" 2>&1) || true
    local migrate_exit=$?

    # Check for expected fatal halt
    if [[ -n "$expect_fatal" ]]; then
        if echo "$migrate_output" | grep -qi "fatal\|halted\|abort"; then
            pass "Migration halted as expected (fatal: $expect_fatal)"
            return 0
        else
            fail "Expected fatal halt for $expect_fatal but migration did not halt"
            echo "    Output: $(echo "$migrate_output" | head -5)"
            return 1
        fi
    fi

    # Migration should have succeeded
    if [[ $migrate_exit -ne 0 ]]; then
        # Check if it's a warning-only exit
        if echo "$migrate_output" | grep -qi "warning"; then
            warn "Migration completed with warnings (exit $migrate_exit)"
        else
            fail "Migration failed (exit $migrate_exit)"
            echo "    Output: $(echo "$migrate_output" | tail -10)"
            return 1
        fi
    else
        pass "Migration completed"
    fi

    # Print any warnings
    echo "$migrate_output" | grep -i "warning\|moved to contrib\|deprecated" | while read -r line; do
        warn "$line"
    done

    # Run go mod tidy
    log "  Running go mod tidy ..."
    local tidy_dir="$chain_dir"
    [[ "$app_dir" != "." ]] && tidy_dir="$chain_dir/$app_dir"
    (cd "$tidy_dir" && go mod tidy 2>&1 | tail -3) || {
        warn "go mod tidy had issues (may be expected for forked chains)"
    }

    # Verification checks
    run_verify "$target_dir" || {
        warn "Some verification checks failed"
        result=1
    }

    # Build check
    if [[ "$expect_build" == "true" && "$SKIP_BUILD" != "1" ]]; then
        log "  Running go build ./... in $tidy_dir ..."
        if (cd "$tidy_dir" && go build ./... 2>&1); then
            pass "Build succeeded"
        else
            fail "Build failed"
            result=1
        fi
    elif [[ "$expect_build" == "false" ]]; then
        warn "Build skipped (expected: false)"
    elif [[ "$SKIP_BUILD" == "1" ]]; then
        warn "Build skipped (E2E_SKIP_BUILD=1)"
    fi

    return $result
}

# ── main ──────────────────────────────────────────────────────────────────────
main() {
    check_deps

    mkdir -p "$WORKDIR"
    log "Work directory: $WORKDIR"

    build_migrate_bin

    local total=0 passed=0 failed=0 skipped=0
    local failed_ids=()

    while IFS= read -r chain_json; do
        [[ -z "$chain_json" ]] && continue
        total=$((total + 1))
        local chain_id
        chain_id=$(echo "$chain_json" | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")

        if run_chain "$chain_json"; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
            failed_ids+=("$chain_id")
        fi
    done < <(parse_chains "$@")

    # Summary
    echo ""
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "Results: $passed/$total passed, $failed failed"
    if [[ ${#failed_ids[@]} -gt 0 ]]; then
        fail "Failed chains: ${failed_ids[*]}"
    fi
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    # Cleanup
    if [[ "$KEEP" != "1" && $failed -eq 0 ]]; then
        log "Cleaning up $WORKDIR"
        rm -rf "$WORKDIR"
    elif [[ "$KEEP" != "1" && $failed -gt 0 ]]; then
        log "Keeping $WORKDIR for debugging (failed chains present)"
    fi

    [[ $failed -eq 0 ]]
}

main "$@"
