#!/usr/bin/env bash
# use this script from the repository root to build all docker images in the repository in sequence, beginning with the development environment, also located in the repository root
docker build -t ghcr.io/cosmos/cosmos-sdk:latest -
docker build --dockerfile contrib/images/proto-builder -t ghcr.io/cosmos/cosmos-sdk/proto-builder:latest -
docker build --dockerfile contrib/images/simapp -t ghcr.io/cosmos/cosmos-sdk/simapp:latest -
docker build --dockerfile contrib/images/simd-dlv -t ghcr.io/cosmos/cosmos-sdk/simd-dlv:latest contrib/images/simd-dlv
docker build --dockerfile contrib/images/simd-env -t ghcr.io/cosmos/cosmos-sdk/simd-env:latest contrib/images/simd-env
