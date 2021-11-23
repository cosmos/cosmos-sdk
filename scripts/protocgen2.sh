#!/usr/bin/env bash

set -eo pipefail

# this script does code generation using protoc-gen-go (eventually pulsar)
# instead of gogo for files that use the new google.golang.org/protobuf API

# NOTE: buf and protoc-gen-go mus be

cd api
buf generate .

