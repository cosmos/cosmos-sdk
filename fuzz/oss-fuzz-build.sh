#!/bin/bash -eu

export FUZZ_ROOT="github.com/cosmos/cosmos-sdk"

compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/setstring Fuzz fuzz_types_dec_setstring fuzz
