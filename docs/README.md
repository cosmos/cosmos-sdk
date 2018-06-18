# Cosmos SDK Documentation

NOTE: This documentation is a work-in-progress!

- [Overview](overview) 
    - [Overview](overview/overview.md) - An overview of the Cosmos-SDK
    - [The Object-Capability Model](overview/capabilities.md) - Security by
      least-privilege
    - [Application Architecture](overview/apps.md) - Layers in the application architecture
- [Install](install.md) - Install the library and example applications
- [Core](core)
    - [Messages](core/messages.md) - Messages contain the content of a transaction
    - [Handlers](core/handlers.md) - Handlers are the workhorse of the app!
    - [BaseApp](core/baseapp.md) - BaseApp is the base layer of the application
    - [The MultiStore](core/multistore.md) - MultiStore is a rich Merkle database
    - [Amino](core/amino.md) - Amino is the primary serialization library used in the SDK
    - [Accounts](core/accounts.md) - Accounts are the prototypical object kept in the store
    - [Transactions](core/transactions.md) - Transactions wrap messages and provide authentication
    - [Keepers](core/keepers.md) - Keepers are the interfaces between handlers
    - [Clients](core/clients.md) - Hook up your app to standard CLI and REST
      interfaces for clients to use!
    - [Advanced](core/advanced.md) - Trigger logic on a timer, use custom
      serialization formats, advanced Merkle proofs, and more!

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
