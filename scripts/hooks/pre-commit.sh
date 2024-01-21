#!/bin/bash

function lint_and_add_modified_go_files() {
  local go_files="$(git diff --name-only --diff-filter=d | grep \.go$ | grep -v \.pb\.go$)"
  for f in $go_files; do
    local dir_name="$(dirname $f)"
    golangci-lint run "${dir_name}" --fix --out-format=tab --issues-exit-code=0
    echo "adding ${f} to git index"
    git add $f
  done
}

lint_and_add_modified_go_files