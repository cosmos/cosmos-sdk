# Mintnet - minting your own crypto-cash

This directory is an example of extending basecoin with a plugin architecture to allow minting your own money. This directory is designed to be stand-alone and can be copied to your own repo as a starting point.  Just make sure to change the import path in `./cmd/mintnet`.

First, make sure everything is working on your system, by running `make all` in this directory, this will update all dependencies, run the test quite, and install the `mintnet` binary.  After that, you can run all commands with mintnet.

## Setting Initial State

**TODO** document SetOption (over abci-cli) and the initial genesis block.  Also using plugin specific options.

## Testing with a CLI

**TODO** Once we authorized some keys to mint cash, let's do it.  And send those shiny new bills to our friends.

## Attaching a GUI

**TODO** showcase matt's ui and examples of how to extend it?
