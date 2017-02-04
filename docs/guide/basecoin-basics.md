# Basecoin Basics

Here we explain how to get started with a simple Basecoin blockchain, and how to send transactions between accounts using the `basecoin` tool.

## Install

Make sure you have [basecoin installed](install.md).
You will also need to [install tendermint](https://tendermint.com/intro/getting-started/download).

## Initialization

Basecoin is an ABCI application that runs on Tendermint, so we first need to initialize Tendermint:

```
tendermint init
```

This will create the necessary files for a single Tendermint node in `~/.tendermint`.
If you had previously run tendermint, make sure you reset the chain
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

## Start

Now we can start basecoin:

```
basecoin start --in-proc
```

This will initialize the chain with the `genesis.json` file from the current directory.  If you want to specify another location, you can run:

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

## Send transactions

Now we are ready to send some transactions.
If you take a look at the `genesis.json` file, you will see one account listed there.
This account corresponds to the private key in `priv_validator.json`.
We also included the private key for another account, in `priv_validator2.json`.

Let's check the balance of these two accounts:

```
basecoin account 0xD397BC62B435F3CF50570FBAB4340FE52C60858F
basecoin account 0x4793A333846E5104C46DD9AB9A00E31821B2F301
```

The first account is flush with cash, while the second account doesn't exist.
Let's send funds from the first account to the second:

```
basecoin sendtx --to 0x4793A333846E5104C46DD9AB9A00E31821B2F301 --amount 10
```

By default, the CLI looks for a `priv_validator.json` to sign the transaction with,
so this will only work if you are in the `$GOPATH/src/github.com/tendermint/basecoin/data`.
To specify a different key, we can use the `--from` flag.

Now if we check the second account, it should have `10` coins!

```
basecoin account 0x4793A333846E5104C46DD9AB9A00E31821B2F301
```

We can send some of these coins back like so:

```
basecoin sendtx --to 0xD397BC62B435F3CF50570FBAB4340FE52C60858F --from priv_validator2.json --amount 5
```

Note how we use the `--from` flag to select a different account to send from.

If we try to send too much, we'll get an error:

```
basecoin sendtx --to 0xD397BC62B435F3CF50570FBAB4340FE52C60858F --from priv_validator2.json --amount 100
```

See `basecoin sendtx --help` for additional details.

## Plugins


The `sendtx` command creates and broadcasts a transaction of type `SendTx`,
which is only useful for moving tokens around.
Fortunately, Basecoin supports another transaction type, the `AppTx`, 
which can trigger code registered via a plugin system.

For instance, we implemented a simple plugin called `counter`, 
which just counts the number of transactions it processed.
To run it, kill the other processes, run `tendermint unsafe_reset_all`, and then 

```
basecoin start --in-proc --counter-plugin
```

Now in another window, we can send transactions with:

```
TODO
```

## Next steps

1. Learn more about [Basecoin's design](basecoin-design.md)
1. Make your own [cryptocurrency using Basecoin plugins](example-counter.md)
1. Learn more about [plugin design](plugin-design.md)
1. See some [more example applications](more-examples.md)
1. Learn how to use [InterBlockchain Communication (IBC)](ibc.md)
1. [Deploy testnets](deployment.md) running your basecoin application.
