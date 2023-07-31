#!/usr/bin/env bash

set -e -o pipefail

REPO_ROOT="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
export REPO_ROOT

lint_module() {
  local root="$1"
  shift
  cd "$(dirname "$root")" &&
    echo "linting $(grep "^module" go.mod) [$(date -Iseconds -u)]" &&
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@"
}
export -f lint_module

# if LINT_DIFF env is set, only lint the files in the current commit otherwise lint all files
if [[ -z "${LINT_DIFF:-}" ]]; then
  find "${REPO_ROOT}" -type f -name go.mod -print0 |
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
    echo "linting $f [$(date -Iseconds -u)]" &&
    cd $f &&
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" &&
    cd $REPO_ROOT
  done
fi