<!--
parent:
  order: false
-->

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

- [Auth](auth/spec/README.md) - Authentication of accounts and transactions for Cosmos SDK application.
- [Bank](bank/spec/README.md) - Token transfer functionalities.
- [Governance](gov/spec/README.md) - On-chain proposals and voting.
- [Staking](staking/spec/README.md) - Proof-of-stake layer for public blockchains.
- [Slashing](slashing/spec/README.md) - Validator punishment mechanisms.
- [Distribution](distribution/spec/README.md) - Fee distribution, and staking token provision distribution.
- [Crisis](crisis/spec/README.md) - Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
- [Mint](mint/spec/README.md) - Creation of new units of staking token.
- [Params](params/spec/README.md) - Globally available parameter store.
- [Supply](supply/spec/README.md) - Total token supply of the chain.

To learn more about the process of building modules, visit the [building modules reference documentation](../docs/building-modules/README.md).
