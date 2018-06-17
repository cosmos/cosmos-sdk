# Cosmos Hub Spec

This directory contains specifications for the state transition machine of the
Cosmos Hub. 

The Cosmos Hub holds all of its state in a Merkle store. Updates to
the store may be made during transactions and at the beginning and end of every
block.

While the first implementation of the Cosmos Hub is built using the Cosmos-SDK,
these specifications aim to be independent of any implementation details. That
said, they provide a detailed resource for understanding the Cosmos-SDK.

- [Store](store) - The core Merkle store that holds the state.
- [Auth](auth) - The structure and authentication of accounts and transactions.
- [Bank](bank) - Sending tokens.
- [Governance](governance) - Proposals and voting.
- [Staking](staking) - Proof-of-stake bonding, delegation, etc.
- [Slashing](slashing) - Validator punishment mechanisms.
- [Provisioning](provisioning) - Fee distribution, and atom provision distribution 
- [IBC](ibc) - Inter-Blockchain Communication (IBC) protocol.
- [Other](other) - Other components of the Cosmos Hub, including the reserve 
  pool, All in Bits vesting, etc.

For details on the underlying blockchain and p2p protocols, see
the [Tendermint specification](https://github.com/tendermint/tendermint/tree/develop/docs/spec).

