# Simple usage with a mounted data directory:
# > docker build -t simapp .
#
# Server:
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.simapp:/root/.simapp simapp simd init test-chain
# TODO: need to set validator in genesis so start runs
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.simapp:/root/.simapp simapp simd start
#
# Client: (Note the simapp binary always looks at ~/.simapp we can bind to different local storage)
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.simappcli:/root/.simapp simapp simd keys add foo
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.simappcli:/root/.simapp simapp simd keys list
#
# This image is pushed to the GHCR as https://ghcr.io/cosmos/simapp

FROM golang:1.22-alpine AS build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev
RUN apk add --no-cache $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/cosmos-sdk

# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./
COPY math/go.mod math/go.sum ./math/
COPY api/go.mod api/go.sum ./api/
COPY core/go.mod core/go.sum ./core/
COPY collections/go.mod collections/go.sum ./collections/
COPY store/go.mod store/go.sum ./store/
COPY x/accounts/go.mod x/accounts/go.sum ./x/accounts/
COPY x/tx/go.mod x/tx/go.sum /x/tx/
COPY x/protocolpool/go.mod x/protocolpool/go.sum ./x/protocolpool/
COPY x/gov/go.mod x/gov/go.sum ./x/gov/
COPY x/distribution/go.mod x/distribution/go.sum ./x/distribution/
COPY x/slashing/go.mod x/slashing/go.sum ./x/slashing/
COPY x/staking/go.mod x/staking/go.sum ./x/staking/
COPY x/auth/go.mod x/auth/go.sum ./x/auth/
COPY x/authz/go.mod x/authz/go.sum ./x/authz/
COPY x/bank/go.mod x/bank/go.sum ./x/bank/
COPY x/mint/go.mod x/mint/go.sum ./x/mint/
COPY depinject/go.mod depinject/go.sum ./depinject/
RUN go mod download

# Add source files
COPY . .

# Dockerfile Cross-Compilation Guide
# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide
ARG TARGETOS TARGETARCH

# install simapp, remove packages
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build

# Use alpine:3 as a base image
FROM alpine:3

EXPOSE 26656 26657 1317 9090
# Run simd by default, omit entrypoint to ease using container with simcli
CMD ["simd"]
STOPSIGNAL SIGTERM
WORKDIR /root

# Install minimum necessary dependencies
RUN apk add --no-cache curl make bash jq sed

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/cosmos/cosmos-sdk/build/simd /usr/bin/simd
