#!/usr/bin/env bash

set -o errexit
set -o nounset
set -x

ROOT=$PWD

SIMD="$ROOT/build/simdv2"

$SIMD store restore 4 3

