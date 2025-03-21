#!/usr/bin/env bash

set -e -o pipefail

LINT_TAGS="e2e,ledger,test_ledger_mock,system_test,sims"
export LINT_TAGS

REPO_ROOT="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
export REPO_ROOT

# Define deprecated modules (directories) to ignore
DEPRECATED_MODULES=(
  "${REPO_ROOT}/x/crisis"
  "${REPO_ROOT}/x/params"
)

lint_module() {
  local root="$1"
  shift
  local module_dir
  module_dir="$(dirname "$root")"

  # Check if the module directory is in the deprecated list
  for dep in "${DEPRECATED_MODULES[@]}"; do
    if [[ "$module_dir" == "$dep" ]]; then
      echo "Skipping deprecated module $(basename "$dep")"
      return 0
    fi
  done

  cd "$module_dir" &&
    echo "linting $(grep '^module' go.mod) [$(date -Iseconds -u)]" &&
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=${LINT_TAGS}
}
export -f lint_module

# if LINT_DIFF env is set, only lint the files in the current commit; otherwise, lint all files
if [[ -z "${LINT_DIFF:-}" ]]; then
  # Use find and prune to exclude deprecated directories
  find "${REPO_ROOT}" \
    \( -path "${REPO_ROOT}/x/crisis" -o -path "${REPO_ROOT}/x/params" \) -prune -o \
    -type f -name go.mod -print0 |
    xargs -0 -I{} bash -c 'lint_module "$@"' _ {} "$@"
else
  if [[ -z $GIT_DIFF ]]; then
    GIT_DIFF=$(git diff --name-only --diff-filter=d | grep \.go$ | grep -v \.pb\.go$) || true
  fi

  if [[ -z "$GIT_DIFF" ]]; then
    echo "no files to lint"
    exit 0
  fi

  for f in $(dirname $(echo "$GIT_DIFF" | tr -d "'") | uniq); do
    # Skip if the directory is one of the deprecated modules
    if [[ "$f" == "${REPO_ROOT}/x/crisis" || "$f" == "${REPO_ROOT}/x/params" ]]; then
      echo "Skipping deprecated module $(basename "$f")"
      continue
    fi
    echo "linting $f [$(date -Iseconds -u)]" &&
    cd "$f" &&
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=${LINT_TAGS} &&
    cd "$REPO_ROOT"
  done
fi
