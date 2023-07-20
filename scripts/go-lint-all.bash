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

lint_files() {
  local go_files="$(git diff --name-only --diff-filter=d | grep \.go$ | grep -v \.pb\.go$)"
  if [[ -z "$go_files" && $GIT_DIFF ]]; then
    go_files="$(echo $GIT_DIFF | grep \.go$ | grep -v \.pb\.go$)"
  elif [[ -z "$go_files" ]]; then
    echo "no files to lint"
    exit 0
  fi

  for f in $go_files; do
    local dir_name="$(dirname $f)"
    echo "linting ${dir_name} [$(date -Iseconds -u)]"
    golangci-lint run "${dir_name}" -c "${REPO_ROOT}/.golangci.yml" "$@"
  done
}
export -f lint_files

# if LINT_DIFF env is set, only lint the files in the current commit otherwise lint all files
if [[ -z "${LINT_DIFF:-}" ]]; then
  find "${REPO_ROOT}" -type f -name go.mod -print0 |
    xargs -0 -I{} bash -c 'lint_module "$@"' _ {} "$@"
else
  lint_files "$@"
fi