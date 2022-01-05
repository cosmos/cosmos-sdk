#!/usr/bin/env bash

# this script is for generating protobuf files for the new google.golang.org/protobuf API

set -eo pipefail

echo "Generating API module"
(cd proto; buf generate --template buf.gen.pulsar.yaml)

(cd orm/internal; buf generate .)
