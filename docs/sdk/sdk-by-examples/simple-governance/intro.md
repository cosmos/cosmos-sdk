# SDK By Examples - Simple Governance Application 

In this tutorial, you will learn the basics of coding an application with the Cosmos-SDK. Applications built on top of the SDK are called *Application-specific blockchains*. They are decentralised applications running on their own blockchains. The application we will build in this tutorial is a simple governance application.

Before getting in the bulk of the code, we will start by some introductory content on Tendermint, Cosmos and the programming philosophy of the SDK. Let us get started!

## Table of contents:

### Introduction - Prerequsite reading

- [Intro to Tendermint and Cosmos](../../../introduction/tendermint-cosmos.md)
- [Tendermint Core and ABCI](../../../introduction/tendermint.md)
- [Intro to Cosmos-SDK](../../overview.md)
- [Starting your own project](start.md)

### Setup and design phase

- [Setup](setup.md)
- [Application design](app-design.md)

### Implementation of the application

**Important note: All the code for this application can be found [here](https://github.com/cosmos/cosmos-sdk/tree/fedekunze/module_tutorial/examples/simpleGov). Snippets will be provided throughout the tutorial, but please refer to the provided link for the full implementation details**

- [Application initialization](app-init.md)
- Simple Governance module
    + [Module initialization](module-init.md)
    + [Types](module-types.md)
    + [Keeper](module-keeper.md)
    + [Handler](module-handler.md)
    + [Wire](module-wire.md)
    + [Errors](module-errors.md)
    + Command-Line Interface and Rest API
        * [Command-Line Interface](module-cli.md)
        * [Rest API](module-rest.md)
- Bridging it all together
    + [Application structure](app-structure.md)
    + [Application CLI and Rest Server](app-commands.md)
        * [Application CLI](app-cli.md)
        * [Rest Server](app-rest.md)
    + [Makefile](app-makefile.md)
    + [Application constructor](app-constructor.md)
    + [Application codec](app-codec.md)
- Running the application
    + [Installation](run-install.md)
    + [Submit a proposal](submit-proposal.md)
    + [Cast a vote](cast-vote.md)

## Useful links

If you have any question regarding this tutorial or about development on the SDK, please reach out us through our official communication channels:

- [Cosmos-SDK Riot Channel](https://riot.im/app/#/room/#cosmos-sdk:matrix.org)
- [Telegram](https://t.me/cosmosproject)

Or open an issue on the SDK repo:

- [Cosmos-SDK repo](https://github.com/cosmos/cosmos-sdk/)
