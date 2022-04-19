<!--
order: 0
-->

# List of Modules

Here are some production-grade modules that can be used in Cosmos SDK applications, along with their respective documentation:

* [Auth](auth/spec/README.md) - Authentication of accounts and transactions for Cosmos SDK applications.
* [Authz](authz/spec/README.md) - Authorization for accounts to perform actions on behalf of other accounts.
* [Bank](bank/spec/README.md) - Token transfer functionalities.
* [Capability](capability/spec/README.md) - Object capability implementation.
* [Crisis](crisis/spec/README.md) - Halting the blockchain under certain circumstances (e.g. if an invariant is broken).
* [Distribution](distribution/spec/README.md) - Fee distribution, and staking token provision distribution.
* [Epoching](epoching/spec/README.md) - Allows modules to queue messages for execution at a certain block height.
* [Evidence](evidence/spec/README.md) - Evidence handling for double signing, misbehaviour, etc.
* [Feegrant](feegrant/spec/README.md) - Grant fee allowances for executing transactions.
* [Governance](gov/spec/README.md) - On-chain proposals and voting.
* [Mint](mint/spec/README.md) - Creation of new units of staking token.
* [Params](params/spec/README.md) - Globally available parameter store.
* [Slashing](slashing/spec/README.md) - Validator punishment mechanisms.
* [Staking](staking/spec/README.md) - Proof-of-Stake layer for public blockchains.
* [Upgrade](upgrade/spec/README.md) - Software upgrades handling and coordination.

To learn more about the process of building modules, visit the [building modules reference documentation](../docs/building-modules/README.md).

## IBC

The IBC module for the SDK has moved to its [own repository](https://github.com/cosmos/ibc-go).
