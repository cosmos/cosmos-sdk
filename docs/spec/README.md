# Specifications

This directory contains specifications for the modules of the Cosmos SDK as well as Interchain Standards (ICS) and other specifications.

SDK applications hold this state in a Merkle store. Updates to
the store may be made during transactions and at the beginning and end of every
block.

## SDK specifications

- [Store](./store) - The core Merkle store that holds the state.
- [Bech32](./addresses/bech32.md) - Address format for Cosmos SDK applications.

## Modules specifications

- [Auth](../../x/auth/spec) - The structure and authentication of accounts and transactions.
- [Bank](../../x/bank/spec) - Sending tokens.
- [Governance](../../x/governance/spec) - Proposals and voting.
- [Staking](../../x/staking/spec) - Proof-of-stake bonding, delegation, etc.
- [Slashing](../../x/slashing/spec) - Validator punishment mechanisms.
- [Distribution](../../x/distribution/spec) - Fee distribution, and staking token provision distribution.
- [Crisis](.../../x/crisis/spec) - Halting the blockchain under certain circumstances.
- [Mint](../../x/mint/spec) - Staking token provision creation.
- [Params](../../x/params/spec) - Globally available parameter store.
- [Supply](../../x/supply/spec) - Total supply of the chain.
- [NFT](https://github.com/cosmos/modules/tree/master/incubator/nft/docs/spec) - Non-fungible tokens.

For details on the underlying blockchain and p2p protocols, see
the [Tendermint specification](https://github.com/tendermint/tendermint/tree/master/docs/spec).
