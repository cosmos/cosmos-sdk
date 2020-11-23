#!/bin/sh

set -e

echo "12345678" | simd tx bank send cosmos1wjmt63j4fv9nqda92nsrp2jp2vsukcke4va3pt "$1" 100stake --chain-id="testing" --node tcp://cosmos:26657 --yes