<!--
parent:
  order: false
-->

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

- [Auth](auth/spec/README.md) - Authentication of accounts and transactions for Cosmos SDK application.
- [Bank](bank/spec/README.md) - Token transfer functionalities.
- [Capability](capability/spec/README.md) - Object capability implementation.
- [Crisis](crisis/spec/README.md) - Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
- [Distribution](distribution/spec/README.md) - Fee distribution, and staking token provision distribution.
- [Evidence](evidence/spec/README.md) - Evidence handling for double signing, misbehaviour, etc.
- [Governance](gov/spec/README.md) - On-chain proposals and voting.
- [IBC](ibc/spec/README.md) - IBC protocol for transport, authentication adn ordering.
- [IBC Transfer](ibc/spec/README.md) - Cross-chain fungible token transfer implementation through IBC.
- [Mint](mint/spec/README.md) - Creation of new units of staking token.
- [Params](params/spec/README.md) - Globally available parameter store.
- [Slashing](slashing/spec/README.md) - Validator punishment mechanisms.
- [Staking](staking/spec/README.md) - Proof-of-Stake layer for public blockchains.
- [Upgrade](upgrade/spec/README.md) - Software upgrades handling and coordination.

To learn more about the process of building modules, visit the [building modules reference documentation](../docs/building-modules/README.md).
