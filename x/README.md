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
* [Consensus](./consensus/README.md) - Consensus module for modifying CometBFT's ABCI consensus params.
* [Distribution](./distribution/README.md) - Fee distribution, and staking token provision distribution.
* [Governance](./gov/README.md) - On-chain proposals and voting.
* [Genutil](./genutil/README.md) - Genesis utilities for the Cosmos SDK.
* [Mint](./mint/README.md) - Creation of new units of staking token.
* [Slashing](./slashing/README.md) - Validator punishment mechanisms.
* [Staking](./staking/README.md) - Proof-of-Stake layer for public blockchains.
* [Upgrade](./upgrade/README.md) - Software upgrades handling and coordination.
* [Evidence](./evidence/README.md) - Evidence handling for double signing, misbehaviour, etc.

## Supplementary Modules

Supplementary modules are modules that are maintained in the Cosmos SDK but are not necessary for
the core functionality of your blockchain.  They can be thought of as ways to extend the 
capabilities of your blockchain or further specialize it.

* [Authz](./authz/README.md) - Authorization for accounts to perform actions on behalf of other accounts.
* [Epochs](./epochs/README.md) - Registration so SDK modules can have logic to be executed on timed tickers.
* [Feegrant](./feegrant/README.md) - Grant fee allowances for executing transactions.
* [ProtocolPool](./protocolpool/README.md) - Extended management of community pool functionality.

## Quantum & Consciousness Modules

* [Nocturne](./nocturne/README.md) - A consciousness-oriented runtime for trauma compression, empathic mirroring, and ethical forgetting.

## Deprecated Modules

The following modules are deprecated.  They will no longer be maintained actively.

* [Crisis](../contrib/x/crisis/README.md) - _Deprecated_ halting the blockchain under certain circumstances (e.g. if an invariant is broken).
* [Params](./params/README.md) - _Deprecated_ Globally available parameter store.
* [NFT](../contrib/x/nft/README.md) - _Deprecated_ NFT module implemented based on [ADR43](https://docs.cosmos.network/main/build/architecture/adr-043-nft-module).
* [Group](../contrib/x/group/README.md) - _Deprecated_ Allows for the creation and management of on-chain multisig accounts.  
* [Circuit](../contrib/x/circuit/README.md) _Deprecated_ - Circuit breaker module for pausing messages.

To learn more about the process of building modules, visit the [building modules reference documentation](https://docs.cosmos.network/main/building-modules/intro).

## IBC

The IBC module for the SDK is maintained by the IBC Go team in its [own repository](https://github.com/cosmos/ibc-go).

Additionally, the [capability module](https://github.com/cosmos/ibc-go/tree/fdd664698d79864f1e00e147f9879e58497b5ef1/modules/capability) is from v0.50+ maintained by the IBC Go team in its [own repository](https://github.com/cosmos/ibc-go/tree/fdd664698d79864f1e00e147f9879e58497b5ef1/modules/capability).

## CosmWasm

The CosmWasm module enables smart contracts, learn more by going to their [documentation site](https://book.cosmwasm.com/), or visit [the repository](https://github.com/CosmWasm/cosmwasm).

## EVM

Read more about writing smart contracts with solidity at the official [`evm` documentation page](https://evm.cosmos.network/).

## Enterprise Modules

In addition to these core and supplementary modules, the Cosmos SDK maintains enterprise-grade modules in the `enterprise/` directory.

**Enterprise modules use different licenses than the Apache 2.0 core SDK modules.** Please review the LICENSE file in each enterprise module directory before use.

### Available Enterprise Modules

* [PoA (Proof of Authority)](../enterprise/poa/README.md) - Admin-controlled validator set for permissioned networks with governance integration.

For complete information about enterprise modules, licensing, and documentation, see the [Enterprise Modules documentation](../enterprise/README.md).

