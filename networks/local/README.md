# Local Cluster with Docker Compose

## Requirements

- [Install gaia](/docs/index.rst)
- [Install docker](https://docs.docker.com/engine/installation/)
- [Install docker-compose](https://docs.docker.com/compose/install/)

## Build

Build the `gaiad` binary and the `tendermint/gaiadnode` docker image.

Note the binary will be mounted into the container so it can be updated without
rebuilding the image.

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk

# Build the linux binary in ./build
make build-linux

# Build tendermint/gaiadnode image
make build-docker-gaiadnode
```


## Run a testnet

To start a 4 node testnet run:

```
make localnet-start
```

The nodes bind their RPC servers to ports 46657, 46660, 46662, and 46664 on the host.
This file creates a 4-node network using the gaiadnode image.
The nodes of the network expose their P2P and RPC endpoints to the host machine on ports 46656-46657, 46659-46660, 46661-46662, and 46663-46664 respectively.

To update the binary, just rebuild it and restart the nodes:

```
make build-linux
make localnet-stop
make localnet-start
```

## Configuration

The `make localnet-start` creates files for a 4-node testnet in `./build` by calling the `gaiad testnet` command.

The `./build` directory is mounted to the `/gaiad` mount point to attach the binary and config files to the container.

For instance, to create a single node testnet:

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk

# Clear the build folder
rm -rf ./build

# Build binary
make build-linux

# Create configuration
docker run -v `pwd`/build:/gaiad tendermint/gaiadnode testnet --o . --v 1

#Run the node
docker run -v `pwd`/build:/gaiad tendermint/gaiadnode

```

## Logging

Log is saved under the attached volume, in the `gaiad.log` file and written on the screen.

## Special binaries

If you have multiple binaries with different names, you can specify which one to run with the BINARY environment variable. The path of the binary is relative to the attached volume.

