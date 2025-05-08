---
sidebar_position: 0
---

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

## Essential Modules

Essential modules include functionality that _must_ be included in your Cosmos SDK blockchain.
These modules provide the core behaviors that are needed for users and operators such as balance tracking,
proof-of-stake capabilities and governance.

* [Auth](./auth/README.md) - Authentication of accounts and transactions for Cosmos SDK applications.
* [Bank](./bank/README.md) - Token transfer functionalities.
* [Circuit](./circuit/README.md) - Circuit breaker module for pausing messages.
* [Consensus](./consensus/README.md) - Consensus module for modifying CometBFT's ABCI consensus params.
* [Distribution](./distribution/README.md) - Fee distribution, and staking token provision distribution.
* [Evidence](./evidence/README.md) - Evidence handling for double signing, misbehaviour, etc.
* [Governance](./gov/README.md) - On-chain proposals and voting.
* [Genutil](./genutil/README.md) - Genesis utilities for the Cosmos SDK.
* [Mint](./mint/README.md) - Creation of new units of staking token.
* [Slashing](./slashing/README.md) - Validator punishment mechanisms.
* [Staking](./staking/README.md) - Proof-of-Stake layer for public blockchains.
* [Upgrade](./upgrade/README.md) - Software upgrades handling and coordination.

## Supplementary Modules

Supplementary modules are modules that are maintained in the Cosmos SDK but are not necessary for
the core functionality of your blockchain.  They can be thought of as ways to extend the 
capabilities of your blockchain or further specialize it.

* [Authz](./authz/README.md) - Authorization for accounts to perform actions on behalf of other accounts.
* [Epochs](./epochs/README.md) - Registration so SDK modules can have logic to be executed at the timed tickers.
* [Feegrant](./feegrant/README.md) - Grant fee allowances for executing transactions.
* [ProtocolPool](./protocolpool/README.md) - Extended management of community pool functionality.

## Deprecated Modules

The following modules are deprecated.  They will no longer be maintained and eventually will be removed
in an upcoming release of the Cosmos SDK per our [release process](https://github.com/cosmos/cosmos-sdk/blob/main/RELEASE_PROCESS.md).

* [Crisis](./crisis/README.md) - *Deprecated* halting the blockchain under certain circumstances (e.g. if an invariant is broken).
* [Params](./params/README.md) - *Deprecated* Globally available parameter store.
* [NFT](./nft/README.md) - *Deprecated* NFT module implemented based on [ADR43](https://docs.cosmos.network/main/architecture/adr-043-nft-module.html).  This module will be moved to the `cosmos-sdk-legacy` repo for use.
* [Group](./group/README.md) - *Deprecated* Allows for the creation and management of on-chain multisig accounts.  This module will be moved to the `cosmos-sdk-legacy` repo for legacy use.

To learn more about the process of building modules, visit the [building modules reference documentation](https://docs.cosmos.network/main/building-modules/intro).

## IBC

The IBC module for the SDK is maintained by the IBC Go team in its [own repository](https://github.com/cosmos/ibc-go).

Additionally, the [capability module](https://github.com/cosmos/ibc-go/tree/fdd664698d79864f1e00e147f9879e58497b5ef1/modules/capability) is from v0.50+ maintained by the IBC Go team in its [own repository](https://github.com/cosmos/ibc-go/tree/fdd664698d79864f1e00e147f9879e58497b5ef1/modules/capability).

## CosmWasm

The CosmWasm module enables smart contracts, learn more by going to their [documentation site](https://book.cosmwasm.com/), or visit [the repository](https://github.com/CosmWasm/cosmwasm).

## EVM

Read more about writing smart contracts with solidity at the official [`evm` documentation page](https://evm.cosmos.network/).
