# Cosmos SDK Documentation

NOTE: This documentation is a work-in-progress!

- [Overview](overview) 
    - [Overview](overview/overview.md) - An overview of the Cosmos-SDK
    - [The Object-Capability Model](overview/capabilities.md) - Security by
      least-privilege
    - [Application Architecture](overview/apps.md) - Layers in the application architecture
- [Install](install.md) - Install the library and example applications
- [Core](core)
    - [BaseApp](core/baseapp.md) - BaseApp is the base layer of the appication
    - [The MultiStore](core/multistore.md) - MultiStore is a rich Merkle database
    - [Messages](core/messages.md) - Messages contain the content of a transaction
    - [Handlers](core/handlers.md) - Handlers are the workhorse of the app!
    - [Amino](core/amino.md) - Amino is the primary serialization library used in the SDK
    - [Accounts](core/accounts.md) - Accounts are the prototypical object kept in the store
    - [Transactions](core/transactions.md) - Transactions wrap messages and provide authentication
    - [Keepers](core/keepers.md) - Keepers are the interfaces between handlers
- [Modules](modules)
    - [Bank](modules/bank.md)
    - [Staking](modules/staking.md)
    - [Slashing](modules/slashing.md)
    - [Provisions](modules/provisions.md)
    - [Governance](modules/governance.md)
    - [IBC](modules/ibc.md)
- [Clients](clients)
    - [Running a Node](clients/node.md)
    - [Key Management](clients/keys.md) - Managing user keys
    - [CLI](clients/cli.md) - Queries and transactions via command line
    - [Light Client Daemon](clients/lcd.md) - Queries and transactions via REST
      API
