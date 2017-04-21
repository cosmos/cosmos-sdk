# Basecoin Basics

Here we explain how to get started with a simple Basecoin blockchain, 
and how to send transactions between accounts using the `basecoin` tool.

## Install

Installing basecoin is simple:

```
go get -u github.com/tendermint/basecoin/cmd/basecoin
```

If you have trouble, see the [installation guide](install.md).

## Initialization

To initialize a new Basecoin blockchain, run:

```
basecoin init 
```

This will create the necessary files for a Basecoin blockchain with one validator and one account in `~/.basecoin`.
For more options on setup, see the [guide to using the Basecoin tool](/docs/guide/basecoin-tool.md).

## Start

Now we can start basecoin:

```
basecoin start
```

You should see blocks start streaming in!

## Send transactions

Now we are ready to send some transactions.
If you take a look at the `genesis.json` file, you will see one account listed there.
This account corresponds to the private key in `key.json`.
We also included the private key for another account, in `key2.json`.

Leave basecoin running and open a new terminal window.
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

By default, the CLI looks for a `key.json` to sign the transaction with.
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
