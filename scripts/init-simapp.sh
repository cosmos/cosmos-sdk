#!/usr/bin/env bash

SIMD_BIN=${SIMD_BIN:=$(which simd 2>/dev/null)}

if [ -z "$SIMD_BIN" ]; then echo "SIMD_BIN is not set. Make sure to run make install before"; exit 1; fi
echo "using $SIMD_BIN"
if [ -d "$($SIMD_BIN config home)" ]; then rm -r $($SIMD_BIN config home); fi
$SIMD_BIN config set client chain-id demo
$SIMD_BIN config set client keyring-backend test
$SIMD_BIN config set client keyring-default-keyname alice
$SIMD_BIN config set app api.enable true
$SIMD_BIN keys add alice --indiscreet
$SIMD_BIN keys add bob --indiscreet
$SIMD_BIN init test --chain-id demo
$SIMD_BIN genesis add-genesis-account alice 5000000000stake --keyring-backend test
$SIMD_BIN genesis add-genesis-account bob 5000000000stake --keyring-backend test
$SIMD_BIN genesis gentx alice 1000000stake --chain-id demo
$SIMD_BIN genesis collect-gentxs
