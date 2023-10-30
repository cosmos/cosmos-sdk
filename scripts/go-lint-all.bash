#!/usr/bin/env bash

set -e -o pipefail

REPO_ROOT="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.." &> /dev/null && pwd )"
export REPO_ROOT

lint_module() {
  local root="$1"
  shift
  cd "$(dirname "$root")" &&
  echo "linting $(grep "^module" go.mod) [$(date -Iseconds -u)]" &&
  if [[ -z "${NIX:-}" ]]; then 
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=e2e,ledger,test_ledger_mock
  else
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=rocksdb,e2e,ledger,test_ledger_mock
  fi
  # always lint simapp with app_v1 build tag, otherwise it never gets linted
  if [[ "$(grep "^module" go.mod)" == "module cosmossdk.io/simapp" ]]; then
    golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=app_v1
  fi
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
    if [[ (-z "${NIX:-}" && $f != store) || $f == "tools/"* ]]; then 
      golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=e2e,ledger,test_ledger_mock
    else
      golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=rocksdb,e2e,ledger,test_ledger_mock
    fi

    if [[ $f == simapp || $f == simapp/simd/cmd ]]; then
      golangci-lint run ./... -c "${REPO_ROOT}/.golangci.yml" "$@" --build-tags=app_v1
    fi
    
    cd $REPO_ROOT
  done
fi
