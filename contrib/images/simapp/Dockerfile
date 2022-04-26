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
FROM faddat/archlinux

RUN pacman -Syyu --noconfirm leveldb rocksdb go base-devel git python

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/cosmos-sdk

# Add source files
COPY . .

# install simapp, remove packages
RUN make build-linux

# Copy over binaries from the build-env
RUN mv /go/src/github.com/cosmos/cosmos-sdk/build/simd /usr/bin/simd

EXPOSE 26656 26657 1317 9090

# Run simd by default, omit entrypoint to ease using container with simcli
CMD ["simd"]