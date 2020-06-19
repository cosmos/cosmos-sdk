<!--
parent:
  order: false
-->

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

- [Auth](auth/) - Authentication of accounts and transactions for Cosmos SDK application.
- [Bank](bank/) - Token transfer functionalities.
- [Capability](capability/docs/README.md) - Object capability implementation.
- [Crisis](crisis/) - Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
- [Distribution](distribution/) - Fee distribution, and staking token provision distribution.
- [Evidence](evidence/) - Evidence handling for double signing, misbehaviour, etc.
- [Governance](gov/) - On-chain proposals and voting.
- [IBC](ibc/) - IBC protocol for transport, authentication adn ordering.
- [IBC Transfer](ibc/) - Cross-chain fungible token transfer implementation through IBC.
- [Mint](mint/) - Creation of new units of staking token.
- [Params](params/) - Globally available parameter store.
- [Slashing](slashing/) - Validator punishment mechanisms.
- [Staking](staking/) - Proof-of-Stake layer for public blockchains.
- [Upgrade](upgrade/) - Software upgrades handling and coordination.

To learn more about the process of building modules, visit the [building modules reference documentation](/building-modules/intro.html).
