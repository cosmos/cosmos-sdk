# Specifications

This directory contains specifications for the modules of the Cosmos SDK as well as Interchain Standards (ICS) and other specifications.

SDK applications hold this state in a Merkle store. Updates to
the store may be made during transactions and at the beginning and end of every
block.

### SDK specifications:

- [Store](./store) - The core Merkle store that holds the state.
- [Bech32](./other/bech32.md) - Address format for Cosmos SDK applications.

### Modules specifications:

- [Auth](./auth) - The structure and authentication of accounts and transactions.
- [Bank](./bank) - Sending tokens.
- [Governance](./governance) - Proposals and voting.
- [Staking](./staking) - Proof-of-stake bonding, delegation, etc.
- [Slashing](./slashing) - Validator punishment mechanisms.
- [Distribution](./distribution) - Fee distribution, and staking token provision distribution .
- [Inflation](./inflation) - Staking token provision creation
- [IBC](./ibc) - Inter-Blockchain Communication (IBC) protocol.

### Interchain standards

- [ICS30](./ics/ics-030-signed-messages.md) - Signed messages standard.
-

For details on the underlying blockchain and p2p protocols, see
the [Tendermint specification](https://github.com/tendermint/tendermint/tree/develop/docs/spec).
