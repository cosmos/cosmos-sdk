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
        - [BaseApp](core/app1.md#baseapp) - BaseApp is the base layer of the application
    - [App2 - Amino](core/app2.md)
        - [Amino](core/app2.md#amino) - Amino is the primary serialization library used in the SDK
    - [App3 - Authentication](core/app3.md)
        - [Accounts](core/app3.md#accounts) - Accounts are the prototypical object kept in the store
        - [Transactions](core/app3.md#transactions) - Transactions wrap messages and provide authentication
    - [App4 - Modules and Keepers](core/app4.md)
        - [Keepers](core/app4.md#keepers) - Keepers are the interfaces between handlers
    - [App5 - Advanced](core/app5.md)
        - [Validator Set Changes](core/app5.md#validators) - Change the
          validator set 
    - [App6 - Basecoin](core/app6.md) - 
        - [Directory Structure](core/app6.md#directory-structure) - Keep your
          application code organized
        - [Clients](core/app6.md#clients) - Hook up your app to standard CLI and REST
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
