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
        - [Ante Handler](core/app2.md#ante-handler) - The AnteHandler
          authenticates transactions
    - [App3 - Modules](core/app3.md)
        - [Accounts](core/app3.md#accounts) - Accounts are the prototypical object kept in the store
          provides Account lookup on a KVStore
        - [Transactions](core/app3.md#transactions) - `StdTx` is the default
          implementation of `Tx`
        - [CoinKeeper](core/app3.md#coin-keeper) - CoinKeeper allows for coin
          transfer on an underlying AccountMapper
    - [App4 - Validator Set Changes](core/app4.md)
        - [InitChain](core/app4.md#init-chain) - Initialize the application
          state
        - [BeginBlock](core/app4.md#begin-block) - BeginBlock logic runs at the
          beginning of every block
        - [EndBlock](core/app4.md#end-block) - EndBlock logic runs at the
          end of every block
    - [App5 - Basecoin](core/app5.md) - 
        - [Directory Structure](core/app5.md#directory-structure) - Keep your
          application code organized
        - [Clients](core/app5.md#clients) - Hook up your app to standard CLI and REST
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
