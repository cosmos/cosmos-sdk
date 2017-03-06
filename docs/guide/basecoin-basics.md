# Basecoin Basics

Here we explain how to get started with a simple Basecoin blockchain, 
and how to send transactions between accounts using the `basecoin` tool.

## Install

Make sure you have [basecoin installed](install.md).
You will also need to [install Tendermint](https://tendermint.com/intro/getting-started/download).

**Note** All code is on the 0.9 pre-release branch, you may have to 
[install Tendermint from source](https://tendermint.com/docs/guides/install) 
until 0.9 is released.  (Make sure to add `git checkout develop` to the linked install instructions)

## Initialization

Basecoin is an ABCI application that runs on Tendermint, so we first need to initialize Tendermint:

```
tendermint init
```

This will create the necessary files for a single Tendermint node in `~/.tendermint`.
If you had previously run Tendermint, make sure you reset the chain
(note this will delete all chain data, so back it up if you need it):

```
tendermint unsafe_reset_all
```

Now we need some initialization files for basecoin.
We have included some defaults in the basecoin directory, under `data`.
For purposes of convenience, change to that directory:

```
cd $GOPATH/src/github.com/tendermint/basecoin/data
```

The directory contains a genesis file and two private keys.

You can generate your own private keys with `tendermint gen_validator`,
and construct the `genesis.json` as you like.
Note, however, that you must be careful with the `chain_id` field,
as every transaction must contain the correct `chain_id`
(default is `test_chain_id`).

## Start

Now we can start basecoin:

```
basecoin start --in-proc
```

This will initialize the chain with the `genesis.json` file from the current directory. 
If you want to specify another location, you can run:

```
basecoin start --in-proc --dir PATH/TO/CUSTOM/DATA
```

Note that `--in-proc` stands for "in process", which means
basecoin will be started with the Tendermint node running in the same process.
To start Tendermint in a separate process instead, use:

```
basecoin start
```

and in another window:

```
tendermint node
```

In either case, you should see blocks start streaming in!
Note, however, that currently basecoin currently requires the
`develop` branch of Tendermint for this to work.

## Send transactions

Now we are ready to send some transactions.
If you take a look at the `genesis.json` file, you will see one account listed there.
This account corresponds to the private key in `key.json`.
We also included the private key for another account, in `key2.json`.

Let's check the balance of these two accounts:

```
basecoin account 0x1B1BE55F969F54064628A63B9559E7C21C925165
basecoin account 0x1DA7C74F9C219229FD54CC9F7386D5A3839F0090
```

The first account is flush with cash, while the second account doesn't exist.
Let's send funds from the first account to the second:

```
basecoin tx send --to 0x1DA7C74F9C219229FD54CC9F7386D5A3839F0090 --amount 10mycoin
```

By default, the CLI looks for a `priv_validator.json` to sign the transaction with,
so this will only work if you are in the `$GOPATH/src/github.com/tendermint/basecoin/data`.
To specify a different key, we can use the `--from` flag.

Now if we check the second account, it should have `10` 'mycoin' coins!

```
basecoin account 0x1DA7C74F9C219229FD54CC9F7386D5A3839F0090
```

We can send some of these coins back like so:

```
basecoin tx send --to 0x1B1BE55F969F54064628A63B9559E7C21C925165 --from key2.json --amount 5mycoin
```

Note how we use the `--from` flag to select a different account to send from.

If we try to send too much, we'll get an error:

```
basecoin tx send --to 0x1B1BE55F969F54064628A63B9559E7C21C925165 --from key2.json --amount 100mycoin
```

See `basecoin tx send --help` for additional details.

## Plugins

The `tx send` command creates and broadcasts a transaction of type `SendTx`,
which is only useful for moving tokens around.
Fortunately, Basecoin supports another transaction type, the `AppTx`,
which can trigger code registered via a plugin system.

In the [next tutorial](example-plugin.md),
we demonstrate how to implement a plugin
and extend the CLI to support new transaction types!
But first, you may want to learn a bit more about [Basecoin's design](basecoin-design.md)
