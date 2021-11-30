#!/usr/bin/env bash

# this script is for generating protobuf files for the new google.golang.org/protobuf API

set -eo pipefail

(cd api; buf generate .)

(cd orm/internal; buf generate .)
