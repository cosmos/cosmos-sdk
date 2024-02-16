#!/bin/bash

set -euo pipefail

export FUZZ_ROOT="github.com/cosmos/cosmos-sdk"

build_go_fuzzer() {
	local function="$1"
	local fuzzer="$2"

	go run github.com/orijtech/otils/corpus2ossfuzz@latest -o "$OUT"/"$fuzzer"_seed_corpus.zip -corpus fuzz/tests/testdata/fuzz/"$function"
	compile_native_go_fuzzer "$FUZZ_ROOT"/fuzz/tests "$function" "$fuzzer"
}

(
	cd math && \
	go get github.com/AdamKorcz/go-118-fuzz-build/testing && \
	compile_native_go_fuzzer cosmossdk.io/math FuzzLegacyNewDecFromStr fuzz_math_legacy_new_dec_from_str
)

go get github.com/AdamKorcz/go-118-fuzz-build/testing

# TODO: fails to build with
# main.413864645.go:12:2: found packages query (collections_pagination.go) and query_test (fuzz_test.go_fuzz.go) in /src/cosmos-sdk/types/query
# because of the separate query_test package.
# compile_native_go_fuzzer "$FUZZ_ROOT"/types/query FuzzPagination fuzz_types_query_pagination
compile_native_go_fuzzer "$FUZZ_ROOT"/types FuzzCoinUnmarshalJSON fuzz_types_coin_unmarshal_json

build_go_fuzzer FuzzCryptoHDDerivePrivateKeyForPath fuzz_crypto_hd_deriveprivatekeyforpath
build_go_fuzzer FuzzCryptoHDNewParamsFromPath fuzz_crypto_hd_newparamsfrompath

build_go_fuzzer FuzzCryptoTypesCompactbitarrayMarshalUnmarshal fuzz_crypto_types_compactbitarray_marshalunmarshal

build_go_fuzzer FuzzTendermintAminoDecodeTime fuzz_tendermint_amino_decodetime

build_go_fuzzer FuzzTypesParseCoin fuzz_types_parsecoin
build_go_fuzzer FuzzTypesParseDecCoin fuzz_types_parsedeccoin
build_go_fuzzer FuzzTypesParseTimeBytes fuzz_types_parsetimebytes
build_go_fuzzer FuzzTypesDecSetString fuzz_types_dec_setstring

build_go_fuzzer FuzzUnknownProto fuzz_unknownproto
