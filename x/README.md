<!--
parent:
  order: false
-->

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

- [Auth](./x/auth/spec/README.md) - Authentication of accounts and transactions for Cosmos SDK application.
- [Bank](./x/bank/spec/README.md) - Token transfer functionalities.
- [Governance](./x/gov/spec/README.md) - On-chain proposals and voting.
- [Staking](./x/staking/spec/README.md) - Proof-of-stake layer for public blockchains.
- [Slashing](./x/slashing/spec/README.md) - Validator punishment mechanisms.
- [Distribution](./x/distribution/spec/README.md) - Fee distribution, and staking token provision distribution.
- [Crisis](./x/crisis/spec/README.md) - Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
- [Mint](./x/mint/spec/README.md) - Creation of new units of staking token.
- [Params](./x/params/spec/README.md) - Globally available parameter store.
- [Supply](./x/supply/spec/README.md) - Total token supply of the chain.

To learn more about the process of building modules, visit the [building modules reference documentation](../docs/building-modules/README.md).
