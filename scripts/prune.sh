#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD

SIMD="$ROOT/build/simdv2"


$SIMD store prune --keep-recent 3
