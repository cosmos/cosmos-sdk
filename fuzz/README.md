# Fuzzing

## Running

The fuzz tests are in standard [Go format](https://go.dev/doc/fuzz/).
To run a fuzz test, use the `-fuzz` flag to `go test`. For example:

```shell
go test -fuzz FuzzCryptoHDNewParamsFromPath ./tests
```
