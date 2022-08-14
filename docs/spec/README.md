---
parent:
  order: false
---

# Specifications

This directory contains specifications for the modules of the Cosmos SDK as well as Interchain Standards (ICS) and other specifications.

SDK applications hold this state in a Merkle store. Updates to
the store may be made during transactions and at the beginning and end of every
block.

## SDK specifications

- [Store](./store) - The core Merkle store that holds the state.
- [Bech32](./addresses/bech32.md) - Address format for Cosmos SDK applications.

## Modules specifications

Go the [module directory](../../x/README.md)

## Tendermint

For details on the underlying blockchain and p2p protocols, see
the [Tendermint specification](https://github.com/tendermint/spec/tree/master/spec).
