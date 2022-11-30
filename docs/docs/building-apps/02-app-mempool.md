---
sidebar_position: 1
---

# Application mempool

:::note Synopsis
This sections describes how the app side mempool can be used and replaced. 
:::


Since `0.47` the application has its own mempool to allow much more granular block building than previous versions. This change was enabled by [ABCI 1.0](https://github.com/tendermint/tendermint/blob/main/spec/abci/README.md). Notably it introduces the prepare and process proposal steps of ABCI. 

## Prepare Proposal

Prepare proposal handles construction of the block, meaning that when a proposer is preparing to propose a block it asks the application if the txs it collected from the mempool are the right ones, at which point the application will check its own mempool for txs that it would like to include. Now, reading mempool twice in the previous sentence is confusing, lets break it down. Tendermint has a mempool that handles bradcasting transactions to other nodes in the network, but it does not handle ordering of these transactions. The ordering happens at the application level in its own mempool. Allowing the application to handle ordering enables the application to define how it would like the block constructed. 

Currently, there is a default `PrepareProposal` implementation provided by the application.

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/baseapp/baseapp.go#L866-L904
```

This default implementation can be overridden by the application developer in favor of a custom implementation:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app.go#L197-L199
```


## Process Proposal

Process proposal handles the validation of what is in a block, meaning that after a block has been proposed the other validators have the right to vote no or yes on a block. The validator in the default implementation of `PrepareProposal` runs the transaction in a non execution fashion, it runs the antehandler and gas operations to make sure the transaction is valid. 

Here is the implementation of the default implementation:

```go reference
https://github.com/cosmos/cosmos-sdk/blob/bcff22a3767b9c5dd7d1d562aece90cf72e05e85/baseapp/baseapp.go#L906-L930
```

Like `PrepareProposal` this implementation is the default and can be modified by the application developer. 

```go reference
https://github.com/cosmos/cosmos-sdk/blob/main/simapp/app.go#L200-L203
```

## Mempool

Now that we have walked through the `PrepareProposal` & `ProcessProposal`, we can move on to walking through the mempool. 

There are countless designs that an application developer can write for a mempool, the core team opted to provide a simple implementation of a nonce mempool. The nonce mempool is a mempool that keeps transactions from an sorted by nonce in order to avoid the issues with nonces. 

It works by storing the transation in a list sorted by the transaction nonce. When the proposer asks for transactions to be included in a block it randomly selects a sender and gets the first transaction in the list. It repeats this until the mempool is empty or the block is full. 

### Configurations

#### MaxTxs

Its an integer value that sets the mempool in one of three modes, bounded, unbounded, or disabled.

- **negative**: Disabled, mempool does not insert new tx and return early.
- **zero**: Unbounded mempool has no tx limit and will never fail with ErrMempoolTxMaxCapacity.
- **positive**: Bounded, it fails with ErrMempoolTxMaxCapacity when maxTx value is the same as CountTx()
