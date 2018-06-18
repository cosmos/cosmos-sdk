# Cosmos Hub Spec

This directory contains specifications for the application level components of 
the Cosmos Hub.

NOTE: the specifications are not yet complete and very much a work in progress.

- [Store](store) - The core Merkle store that holds the state.
- [Auth](auth) - The structure and authnentication of accounts and transactions.
- [Bank](bank) - Sending tokens.
- [Governance](governance) - Proposals and voting.
- [IBC](ibc) - Inter-Blockchain Communication (IBC) protocol.
- [Staking](staking) - Proof-of-stake bonding, delegation, etc.
- [Slashing](slashing) - Validator punishment mechanisms.
- [Provisioning](provisioning) - Fee distribution, and atom provision distribution 
- [Other](other) - Other components of the Cosmos Hub, including the reserve 
  pool, All in Bits vesting, etc.

The [specification for Tendermint](https://github.com/tendermint/tendermint/tree/develop/docs/specification/new-spec),
i.e. the underlying blockchain, can be found elsewhere.
