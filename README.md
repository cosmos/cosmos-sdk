# Basecoin Examples

This repository contains example code, showing how you can build your own cryptocurrency on top of basecoin.  You can clone one of these subdirectories as a starting place for your own application.  Each subdirectory is thought out like a stand-alone application, and all code is under Apache 2.0 license.

It also contains some step-by-step tutorials on getting stated... at least until we find a better place to put them.

**NOTE**: all this code is currently based on a non-master branch of basecoin to make use of the newest functionality, so make sure to update dependencies properly here, using `make get_vendor_deps` **prior** to running any code.  It is important to copy the glide files to any clone as well.

## Contents

1. [All your coins belong to me](#initial-setup)
1. [More money, more money](#minting-coin)
1. [Escows and other financial instruments](#financial-instruments)
1. [Deploying with tendermint](#deployment)

## Initial Setup

Before beginning with these guides, please make sure you understand how to [install and compile](https://github.com/tendermint/basecoin/blob/master/README.md) the basecoin code. Then, we will go through initializing the state, inspecting the state, and sending money with a simple cli.

**TODO**

## Minting Coin

**Working Code Here**

You just read about the amazing [plugin system](https://github.com/tendermint/basecoin/blob/master/README.md#plugins), and want to use it to print your own money.  Me too!  Let's get started with a simple plugin extension to basecoin, called [mintcoin](./mintcoin/README.md).

## Financial Instruments

Sure, printing money and sending it is nice, but sometimes I don't fully trust the guy at the other end. Maybe we could add an escrow service? Or how about options for currency trading, since we support multiple currencies? No problem, this is also just a plugin away.  Checkout our [trader application](./trader).

**TODO**

## Deployment

Up until this point, we have only been testing the code as a stand-alone abci app, which is nice for developing, but it is no blockchain.  Just a blockchain-ready application.

This section will demonstrate how to launch your basecoin-based application along with a tendermint testnet and initialize the genesis block for fun and profit.

**TODO** Maybe we link to a blog post for this???
