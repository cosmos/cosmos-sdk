# Mintcoin - minting your own crypto-cash

Mintcoin is a Basecoin plugin that allows more coins to be created in the network
by registering some accounts as "Central Bankers" that can issue more money.

For more details about Basecoin, the tools, and the plugin architecture, [see the docs](https://github.com/tendermint/basecoin).

## Install

Run `make all` in this directory.
This will update all dependencies, run the test suite, and install the `mintcoin` binary to your `$GOPATH/bin`.

## Setting Initial State

The state is initialized using a `genesis.json` containing a list of issuers.
These are the accounts that can issue new coins.
An example can be found in `data/genesis.json`.

`mintcoin` uses the `SetOption` plugin method to enable new issuers to be added or removed with the `add` and `remove` keys, respectively. The value must be the hex-encoded address of the issuer to add or remove.

Once an address is added, the private key that belongs to that address can sign MintTx transactions
that create money.

## Minting Money

The `mintcoin` plugin expects the `Data` in the `AppTx` to contain a serialized `MintTx`:

```
type MintTx struct {
	Credits Credits
}

type Credits []Credit

type Credit struct {
	Addr   []byte
	Amount types.Coins
}
```

If the sender of the `AppTx` is a registered issuer,
the corresponding amounts in the embedded `MintTx` will be credited to the listed accounts.

## Testing with a CLI

Alright, now let's set ourselves up as issuers and send some shiny new bills to our friends!

First we do the usual tendermint reset routine:

```
tendermint init
tendermint unsafe_reset_all
```

Now we can start Basecoin with the mintcoin plugin and the default genesis:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/mintcoin/data
mintcoin start --in-proc
```

In another window, we can run the client tool:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/mintcoin/data
mintcoin account 0xD397BC62B435F3CF50570FBAB4340FE52C60858F
```

This was the account registered in the genesis; it has the right number of coins.

Let's mint some new coins:

```
mintcoin tx mint --chain_id mint_chain_id --amount 1blank --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 1000BTC
mintcoin tx mint --chain_id mint_chain_id --amount 1blank --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 5cosmo
mintcoin tx mint --chain_id mint_chain_id --amount 1blank --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 5000FOOD
```

Here, we're sending `1000 BTC`, `5 cosmo`, and `5000 FOOD` to the account with address `D397BC62B435F3CF50570FBAB4340FE52C60858`.
Note that we have to provide some non-zero `--amount` for the transaction, and we have to specify the `--chain_id`,
which must match the `chain_id` in the `genesis.json`.

Let's take another look at the account:

```
mintcoin account 0xD397BC62B435F3CF50570FBAB4340FE52C60858F
```

It's got all the coins!

Alright, let's issue some coins to our friend:

```
mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301
mintcoin tx mint --chain_id mint_chain_id --amount 1blank --mintto 4793A333846E5104C46DD9AB9A00E31821B2F301 --mint 1234BTC
mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301
```

Now they can send us some coins for our labour:

```
mintcoin tx send --chain_id mint_chain_id --from key2.json --to D397BC62B435F3CF50570FBAB4340FE52C60858F --amount 333BTC
mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301
mintcoin account D397BC62B435F3CF50570FBAB4340FE52C60858F
```

If we try to issue coins from the wrong account, we'll get an error:

```
mintcoin tx mint --from key2.json --chain_id mint_chain_id --amount 1blank --mintto 4793A333846E5104C46DD9AB9A00E31821B2F301 --mint 1234BTC
```

## Attaching a GUI

Coming soon! For now, see [the repository](https://github.com/tendermint/js-basecoin)
