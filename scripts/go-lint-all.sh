#!/usr/bin/env bash

set -e

args=("$@")

export pwd=$(pwd)

for modfile in $(find . -name go.mod); do
 echo "linting $(dirname $modfile)"
 DIR=$(dirname $modfile)
 (cd $DIR; golangci-lint run ./... -c $pwd/.golangci.yml $args)
done
