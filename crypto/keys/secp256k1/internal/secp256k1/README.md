# secp256k1

 This package is copied from https://github.com/ethereum/go-ethereum/tree/8fddf27a989e246659fd018ea9be37b2b4f55326/crypto/secp256k1

 Unlike the rest of go-ethereum it is [3-clause BSD](https://opensource.org/licenses/BSD-3-Clause) licensed so compatible with our Apache2.0 license. We opt to copy in here rather than depend on go-ethereum to avoid issues with vendoring of the GPL parts of that repository by downstream.

## Duplicate Symbols

If a project makes use of go-ethereum and Cosmos's cgo secp256k1, C linker will fail with duplicated symbols. To avoid
this `ldflags` must set to allow multiple definitions. This only works with Linux machines.

#### Gcc

 + `go build -tags libsecp256k1_sdk  -ldflags=all="-extldflags=-Wl,--allow-multiple-definition"`

#### Clang

 + `go build -tags libsecp256k1_sdk -ldflags=all="-extldflags=-zmuldefs"`