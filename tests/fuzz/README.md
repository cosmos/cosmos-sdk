# Fuzzing

This directory contains fuzz tests for the Cosmos SDK to identify potential bugs, crashes, and security vulnerabilities through automated input generation.

## What is Fuzzing?

Fuzzing is a testing technique that automatically generates random inputs to find edge cases, crashes, and unexpected behavior in code. The Go fuzzing framework uses coverage-guided fuzzing to efficiently explore code paths.

## Available Fuzz Tests

The following fuzz tests are available:

- **Crypto HD**: `FuzzCryptoHDNewParamsFromPath`, `FuzzCryptoHDDerivePrivateKeyForPath`
- **Types**: `FuzzTypesParseTimeBytes`, `FuzzTypesParseDecCoin`, `FuzzTypesParseCoin`, `FuzzTypesVerifyAddressFormat`, `FuzzTypesDecSetString`
- **Tendermint**: `FuzzTendermintAminoDecodeTime`
- **Crypto Types**: `FuzzCryptoTypesCompactBitArrayMarshalUnmarshal`
- **Unknown Proto**: `FuzzUnknownProto`

## Running Fuzz Tests

The fuzz tests use the standard [Go fuzzing format](https://go.dev/doc/fuzz/). To run a specific fuzz test, use the `-fuzz` flag with `go test`:

```shell
# Run a specific fuzz test
go test -fuzz FuzzCryptoHDNewParamsFromPath ./tests/fuzz/tests

# Run all fuzz tests
go test -fuzz . ./tests/fuzz/tests

# Run with time limit (e.g., 30 seconds)
go test -fuzz FuzzTypesParseTimeBytes -fuzztime 30s ./tests/fuzz/tests
```

## Interpreting Results

- **Crashes**: If a fuzz test crashes, it will save the input that caused the crash in the `testdata` directory
- **Coverage**: Fuzzing automatically tracks code coverage and focuses on unexplored paths
- **Timeouts**: Long-running fuzz tests can be stopped with `Ctrl+C`

## Requirements

- Go 1.18+ (required for built-in fuzzing support)
- The `gofuzz` build tag is also supported for compatibility

## Contributing

When adding new fuzz tests:
1. Follow the naming convention: `Fuzz<Package><Function>`
2. Include proper error handling and input validation
3. Add the test to this README
4. Ensure the test can run for extended periods without crashes
