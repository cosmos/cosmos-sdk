# Quick Start

This guide serves as a practical introduction to building blockchains with the Cosmos SDK.  It shows how to scaffold the code for a basic blockchain node, build and run it. Several important concepts of the Cosmos SDK are introduced along the way. 

## Setup

::: tip
To follow this guide, you need to [install golang](https://golang.org/doc/install) and set [your $GOPATH environment variable](https://golang.org/doc/code.html#GOPATH)
:::

::: warning
Make sure you are using the latest stable version of golang available on https://golang.org/dl/
::: 

First, download the [`scaffold`](https://github.com/cosmos/scaffold) tool:

```bash
git clone https://github.com/cosmos/scaffold
```

The `scaffold` tool lets you easily scaffold boilerplate Cosmos SDK applications. Once you have downloaded it, simply install it on your machine:

```bash
cd scaffold
make
```

## Create a Basic Cosmos SDK Blockchain

To create a basic Cosmos SDK application, simply type in the following command:

```bash
scaffold app lvl-1 <username|org> <repo>
```

where `username|org` is the name of your github/gitlab/atlassian username or organisation, and `repo` the name of the distant repository you would push your application too. These arguments are used to configure the imports so that people can easily download and install your application once (if) you upload it. 

The command above creates a starter application in a new folder named after the `repo` argument. This application contains the [basic logic most SDK applications](../intro/sdk-app-architecture.md) need as well as a set of standard [modules](../building-modules/intro.md) already hooked up. These include:

- [`auth`](../../x/auth/spec/): Accounts, signatures and fees.
- [`bank`](../../x/bank/spec/): Token transfers.
- [`staking`](../../x/staking/spec/): Proof-of-Stake logic, which is a way of managing validator set changes in public decentralised networks. Also includes delegation logic. 
- [`slashing`](../../x/slashing/spec/): Slash validators that misebehave. Complementary to the `staking` module.
- [`distribution`](../../x/distribution/spec/): Distribution of rewards and fees earned by participants in the Proof-of-Stake system (delegators and validators). 
- [`params`](../../x/params/spec/): Global parameter store of the application. 
- [`supply`](../../x/supply/spec/): Handles global token supply of the application. Enables modules to hold tokens. 
- [`genutil`](../../x/genutil) and [`genaccounts`](../../x/genaccounts): Utility modules to facilitate creation of genesis file. 

Now, go into the application's folder. The structure should look like the following:

```
├── app/
│   ├── app.go
│   └── export.go
├── cmd/
│   ├── acli/
│   │   └── main.go
│   ├── aud/
│   │   └── main.go
├── Makefile
├── go.mod
└── x/
```

where:

- `app.go` is the [main file](../basics/app-anatomy.md#core-application-file) defining the application logic. This is where the state is intantiated and modules are declared. This is also where the Cosmos SDK is imported as a dependency to help build the application.
- `export.go` is a helper file used to export the state of the application into a new genesis file. It is helpful when you want to upgrade your chain to a new (breaking) version. 
- `acli/main.go` builds the command-line interface for your blockchain application. It enables end-users to create transactions and query the chain for information. 
- `aud/main.go` builds the main [daemon client](../basics/app-anatomy.md#node-client) of the chain. It is used to run a full-node that will connect to peers and sync its local application state with the latest state of the network. 
- `go.mod` helps manage dependencies. The two main dependencies used are the Cosmos SDK to help build the application, and Tendermint to replicate it. 
- `x/` is the folder to place all the custom modules built specifically for the application. In general, most of the modules used in an application have already been built by third-party developers and only need to be imported in `app.go`. These modules do not need to be cloned into the application's `x/` folder. This is why the basic application shown above, which uses several modules, works despite having an empty `x/` folder. 

## Run your Blockchain

First, install the two main entrypoints of your blockchain, `aud` and `acli`:

```bash
go mod tidy
make install
```

Make sure the clients are properly installed:

```bash
acli --help
aud --help
```

Now that you have your daemon client `aud` and your command-line interface `acli` installed, go ahead and initialize your chain:

```bash
aud init <node-moniker> --chain-id test
```

The command above creates all the configuration files needed for your node to run, as well as a default genesis file, which defines the initial state of the network. Before starting the chain, you  need to populate the state with at least one account. To do so, first create a new [account](../basics/accounts.md) named `validator` (feel free to choose another name):

```bash
acli keys add validator
``` 

Now that you have created a local account, go ahead and grant it `stake` tokens in your chain's genesis file. Doing so will also make sure your chain is aware of this account's existence:

```bash
aud add-genesis-account $(acli keys show validator -a) 100000000stake
``` 

Now that your account has some tokens, you need to add a validator to your chain. Validators are special full-nodes that participate in the consensus process (implemented in the [underlying consensus engine](../intro/sdk-app-architecture.md#tendermint)) in order to add new blocks to the chain. Any account can declare its intention to become a validator operator, but only those with sufficient delegation get to enter the active set (for example, only the top 125 validator candidates with the most delegation get to be validators in the Cosmos Hub). For this guide, you will add your local node (created via the `init` command above) as a validator of your chain. Validators can be declared before a chain is first started via a special transaction included in the genesis file called a `gentx`:

```bash
// create a gentx
aud gentx --name validator --amount 100000stake

// add the gentx to the genesis file
aud collect-gentxs
```

A `gentx` does three things: 

    1. Makes the `validator` account you created into a validator operator account (i.e. the account that controls the validator).
    2. Self-delegates the provided `amount` of staking tokens. 
    3. Link the operator account with a Tendermint node pubkey that will be used for signing blocks. If no `--pubkey` flag is provided, it defaults to the local node pubkey created via the `aud init` command above. 

For more on `gentx`, use the following command:

```bash
aud gentx --help
```

Now that everyting is set up, you can finally start your node:

```bash
aud start
```

You should see blocks come in. 

## Send Tokens and Increase Delegation

Now that your chain is running, it is time to try sending tokens from the first account you created to a second account. In a new terminal window, start by running the following query command:

```bash
acli query account $(acli keys show validator -a) --chain-id test
```

You should see the current balance of the account you created, equal to the original balance of `stake` you granted it minus the amount you delegated via the `gentx`. Now, create a second account:

```bash
acli keys add receiver
```

The command above creates a local key-pair that is not yet registered on the chain. An account is registered the first time it receives tokens from another account. Now, run the following command to send tokens to the second account: 

```bash
acli tx send $(acli keys show validator -a) $(acli keys show receiver -a) 1000stake --chain-id test
```

Check that the second account did receive the tokens:

```bash
acli query account $(acli keys show receiver -a) --chain-id test
```

Finally, delegate some of the stake tokens sent to the `receiver` account to the validator:

```bash
acli tx staking delegate $(acli keys show validator --bech val -a) 500stake --from receiver --chain-id test
``` 

Try to query the total delegations to `validator`:

```bash
acli query staking delegations-to $(acli keys show validator --bech val -a) --chain-id test
```

You should see two delegations, the first one made from the `gentx`, and the second one you just performed from the `receiver` account. 

## Next

Congratulations on making it to the end of this short introduction guide! If you want to learn more, check out the following resources:

- [How to build a full SDK application from scratch](https://tutorials.cosmos.network/nameservice/tutorial/00-intro.html).
- [Read the Cosmos SDK Documentation](../intro/overview.md). 

