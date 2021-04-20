#!/bin/bash -eu

export FUZZ_ROOT="github.com/cosmos/cosmos-sdk"

compile_go_fuzzer "$FUZZ_ROOT"/fuzz/crypto/hd/deriveprivatekeyforpath Fuzz fuzz_crypto_hd_deriveprivatekeyforpath fuzz

compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/parsecoin Fuzz fuzz_types_parsecoin fuzz
compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/parsedeccoin Fuzz fuzz_types_parsedeccoin fuzz
compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/parsetimebytes Fuzz fuzz_types_parsetimebytes fuzz
compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/verifyaddressformat Fuzz fuzz_types_verifyaddressformat fuzz
compile_go_fuzzer "$FUZZ_ROOT"/fuzz/types/dec/setstring Fuzz fuzz_types_dec_setstring fuzz

compile_go_fuzzer "$FUZZ_ROOT"/fuzz/unknownproto Fuzz fuzz_unknownproto fuzz

compile_go_fuzzer "$FUZZ_ROOT"/fuzz/x/bank/types/addressfrombalancesstore Fuzz fuzz_x_bank_types_addressfrombalancesstore fuzz
