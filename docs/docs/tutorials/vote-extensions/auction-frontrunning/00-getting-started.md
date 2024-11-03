# Getting Started

## Table of Contents

- [Getting Started](#overview-of-the-project)
- [Understanding Front-Running](./01-understanding-frontrunning.md)
- [Mitigating Front-running with Vote Extensions](./02-mitigating-front-running-with-vote-extesions.md)
- [Demo of Mitigating Front-Running](./03-demo-of-mitigating-front-running.md)

## Getting Started

### Overview of the Project

This tutorial outlines the development of a module designed to mitigate front-running in nameservice auctions. The following functions are central to this module:

* `ExtendVote`: Gathers bids from the mempool and includes them in the vote extension to ensure a fair and transparent auction process.
* `PrepareProposal`: Processes the vote extensions from the previous block, creating a special transaction that encapsulates bids to be included in the current proposal.
* `ProcessProposal`: Validates that the first transaction in the proposal is the special transaction containing the vote extensions and ensures the integrity of the bids.

In this advanced tutorial, we will be working with an example application that facilitates the auctioning of nameservices. To see what frontrunning and nameservices are [here](./01-understanding-frontrunning.md) This application provides a practical use case to explore the prevention of auction front-running, also known as "bid sniping", where a validator takes advantage of seeing a bid in the mempool to place their own higher bid before the original bid is processed.

The tutorial will guide you through using the Cosmos SDK to mitigate front-running using vote extensions. The module will be built on top of the base blockchain provided in the `tutorials/base` directory and will use the `auction` module as a foundation. By the end of this tutorial, you will have a better understanding of how to prevent front-running in blockchain auctions, specifically in the context of nameservice auctioning.

## What are Vote extensions?

Vote extensions is arbitrary information which can be inserted into a block. This feature is part of ABCI 2.0, which is available for use in the SDK 0.50 release and part of the 0.38 CometBFT release.

More information about vote extensions can be seen [here](https://docs.cosmos.network/main/build/abci/vote-extensions).

## Requirements and Setup

Before diving into the advanced tutorial on auction front-running simulation, ensure you meet the following requirements:

* [Golang >1.21.5](https://golang.org/doc/install) installed
* Familiarity with the concepts of front-running and MEV, as detailed in [Understanding Front-Running](./01-understanding-frontrunning.md)
* Understanding of Vote Extensions as described [here](https://docs.cosmos.network/main/build/abci/vote-extensions)

You will also need a foundational blockchain to build upon coupled with your own module. The `tutorials/base` directory has the necessary blockchain code to start your custom project with the Cosmos SDK. For the module, you can use the `auction` module provided in the `tutorials/auction/x/auction` directory as a reference but please be aware that all of the code needed to implement vote extensions is already implemented in this module.

This will set up a strong base for your blockchain, enabling the integration of advanced features such as auction front-running simulation.
