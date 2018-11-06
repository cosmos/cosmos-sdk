# SDK By Examples - Simple Governance Application 

In this tutorial, you will learn the basics of coding an application with the Cosmos-SDK. Applications built on top of the SDK are called *Application-specific blockchains*. They are decentralised applications running on their own blockchains. The application we will build in this tutorial is a simple governance application.

Before getting in the bulk of the code, we will start by some introductory content on Tendermint, Cosmos and the programming philosophy of the SDK. Let us get started!

## Table of contents:

### Introduction - Prerequisite reading

- [Intro to Tendermint and Cosmos](/introduction/tendermint-cosmos.md)
- [Tendermint Core and ABCI](/introduction/tendermint.md)
- [Intro to Cosmos-SDK](/sdk/overview.md)


### [Setup and design](setup-and-design.md)

- [Starting your own project](setup-and-design.md#get-started)
- [Setup](setup-and-design.md#setup)
- [Application design](setup-and-design.md#application-design)

### Implementation of the application

**Important note: All the code for this application can be found [here](https://github.com/cosmos/cosmos-sdk/tree/fedekunze/module_tutorial/examples/simpleGov). Snippets will be provided throughout the tutorial, but please refer to the provided link for the full implementation details**

- [Application initialization](app-init.md)
- [Simple Governance module](simple-gov-module.md)
    + [Module initialization](simple-gov-module.md#module-initialization)
    + [Types](simple-gov-module.md#types)
    + [Keeper](simple-gov-module.md#keeper)
    + [Handler](simple-gov-module.md#handler)
    + [Wire](simple-gov-module.md#codec)
    + [Errors](simple-gov-module.md#errors)
    + [Command-Line Interface](simple-gov-module.md#command-line-interface)
    + [Rest API](simple-gov-module.md#rest-api)
- [Bridging it all together](bridging-it-all.md)
    + [Application structure](bridging-it-all.md#application-structure)
    + [Application CLI and Rest Server](bridging-it-all.md#cli-and-rest-server)
    + [Makefile](bridging-it-all.md#makefile)
    + [Application constructor](bridging-it-all.md#application-constructor)
    + [Application codec](bridging-it-all.md#application-codec)
- [Running the application](running-the-application.md)
    + [Installation](running-the-application.md#installation)
    + [Submit a proposal](running-the-application.md#submit-a-proposal)
    + [Cast a vote](running-the-application.md#cast-a-vote)

## Useful links

If you have any question regarding this tutorial or about development on the SDK, please reach out us through our official communication channels:

- [Cosmos-SDK Riot Channel](https://riot.im/app/#/room/#cosmos-sdk:matrix.org)
- [Telegram](https://t.me/cosmosproject)

Or open an issue on the SDK repo:

- [Cosmos-SDK repo](https://github.com/cosmos/cosmos-sdk/)
