# Cosmos SDK Documentation

NOTE: This documentation is a work-in-progress!

- [Overview](overview) 
    - [Overview](overview/overview.md) - An overview of the Cosmos-SDK
    - [The Object-Capability Model](overview/capabilities.md) - Security by
      least-privilege
    - [Application Architecture](overview/apps.md) - Layers in the application architecture
- [Install](install.md) - Install the library and example applications
- [Core](core)
    - [Introduction](core/intro.md) - Intro to the tutorial
    - [App1 - The Basics](core/app1.md)
        - [Messages](core/app1.md#messages) - Messages contain the content of a transaction
        - [Stores](core/app1.md#kvstore) - KVStore is a Merkle Key-Value store. 
        - [Handlers](core/app1.md#handlers) - Handlers are the workhorse of the app!
        - [Tx](core/app1.md#tx) - Transactions are the ultimate input to the
          application
        - [BaseApp](core/app1.md#baseapp) - BaseApp is the base layer of the application
    - [App2 - Transactions](core/app2.md)
        - [Amino](core/app2.md#amino) - Amino is the primary serialization library used in the SDK
        - [Ante Handler](core/app2.md#antehandler) - The AnteHandler
          authenticates transactions
    - [App3 - Modules: Auth and Bank](core/app3.md)
        - [auth.Account](core/app3.md#accounts) - Accounts are the prototypical object kept in the store
        - [auth.AccountMapper](core/app3.md#account-mapper) - AccountMapper gets and sets Account on a KVStore
        - [auth.StdTx](core/app3.md#stdtx) - `StdTx` is the default implementation of `Tx`
        - [auth.StdSignBytes](core/app3.md#signing) - `StdTx` must be signed with certain
          information
        - [auth.AnteHandler](core/app3.md#antehandler) - The `AnteHandler`
          verifies `StdTx`, manages accounts, and deducts fees
        - [bank.CoinKeeper](core/app3.md#coinkeeper) - CoinKeeper allows for coin
          transfers on an underlying AccountMapper
    - [App4 - ABCI](core/app4.md)
        - [ABCI](core/app4.md#abci) - ABCI is the interface between Tendermint
          and the Cosmos-SDK
        - [InitChain](core/app4.md#initchain) - Initialize the application
          store
        - [BeginBlock](core/app4.md#beginblock) - BeginBlock runs at the
          beginning of every block and updates the app about validator behaviour
        - [EndBlock](core/app4.md#endblock) - EndBlock runs at the
          end of every block and lets the app change the validator set.
        - [Query](core/app4.md#query) - Query the application store
        - [CheckTx](core/app4.md#checktx) - CheckTx only runs the AnteHandler
    - [App5 - Basecoin](core/app5.md) - 
        - [Directory Structure](core/app5.md#directory-structure) - Keep your
          application code organized
        - [Tendermint Node](core/app5.md#tendermint-node) - Run a full
          blockchain node with your app
        - [Clients](core/app5.md#clients) - Hook up your app to CLI and REST
            interfaces for clients to use!

- [Modules](modules)
    - [Bank](modules/bank.md)
    - [Staking](modules/staking.md)
    - [Slashing](modules/slashing.md)
    - [Provisions](modules/provisions.md)
    - [Governance](modules/governance.md)
    - [IBC](modules/ibc.md)

- [Clients](clients)
    - [Running a Node](clients/node.md) - Run a full node!
    - [Key Management](clients/keys.md) - Managing user keys
    - [CLI](clients/cli.md) - Queries and transactions via command line
    - [Light Client Daemon](clients/lcd.md) - Queries and transactions via REST
      API
