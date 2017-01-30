# Basecoin Tutorial

This tutorial explains how to get started with a simple Basecoin blockchain, and how to send transactions between accounts.

## Install

See the instructions for [installing basecoin](https://github.com/tendermint/basecoin#installation) and [tendermint](https://github.com/tendermint/tendermint#install).

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

## Start

Now we can start basecoin:

```
basecoin start --in-proc
```

This will initialize the chain with the `genesis.json` file from the current directory.  If you want to specify another location, you can run:

```
basecoin start --in-proc --dir PATH/TO/CUSTOM/DATA
```

This will start basecoin with the Tendermint node running in the same process.
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


## Next steps

In future tutorials, we'll discuss how to use `basecoin apptx` to interact with an application plugin that enables basecoin to do more than just send money around.
Stay tuned!
