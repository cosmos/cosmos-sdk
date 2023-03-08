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
# TODO: demo connecting rest-server (or is this in server now?)

# bullseye already comes with build dependencies, so we don't need anything extra to install
FROM --platform=$BUILDPLATFORM golang:1.20-bullseye AS build-env

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/cosmos-sdk

# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./
COPY collections/go.mod collections/go.sum ./collections/
COPY store/go.mod store/go.sum ./store/
COPY log/go.mod log/go.sum ./log/

RUN go mod download

# Add source files
COPY . .

# Dockerfile Cross-Compilation Guide
# https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide
ARG TARGETOS TARGETARCH

# install simapp, remove packages
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build


# Final image, without build artifacts. `/base` already contains openssl, glibc and all required libs to start an app
FROM gcr.io/distroless/base

EXPOSE 26656 26657 1317 9090
# Run simd by default, omit entrypoint to ease using container with simcli
CMD ["simd"]
STOPSIGNAL SIGTERM
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/cosmos/cosmos-sdk/build/simd /usr/bin/simd
