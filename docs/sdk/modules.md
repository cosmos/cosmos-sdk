# Modules

::: tip Note
🚧 We are actively working on documentation for SDK modules.
:::

The Cosmos SDK has all the necessary pre-built modules to add functionality on top of a `BaseApp`, which is the template to build a blockchain dApp in Cosmos. In this context, a `module` is a fundamental unit in the Cosmos SDK.

Each module is an extension of the `BaseApp`'s functionalities that defines transactions, handles application state and manages the state transition logic. Each module also contains handlers for messages and transactions, as well as REST and CLI for secure user interactions.

Some of the most important modules in the SDK are:

- `Auth`: Defines a standard account structure (`BaseAccount`) and how transaction signers are authenticated.
- `Bank`: Defines how coins or tokens (i.e cryptocurrencies) are transferred.
- `Fees`:
- `Governance`: Governance related implementation including proposals and voting.
- `IBC`: Defines the intereoperability of blockchain zones according to the specifications of the [IBC Protocol](https://cosmos.network/whitepaper#inter-blockchain-communication-ibc).
- `Slashing`:
- `Staking`: Proof of Stake related implementation including bonding and delegation transactions, inflation, fees, unbonding, etc.
