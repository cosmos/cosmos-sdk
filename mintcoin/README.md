# Mintcoin - minting your own crypto-cash

This directory is an example of extending basecoin with a plugin architecture to allow minting your own money. This directory is designed to be stand-alone and can be copied to your own repo as a starting point.  Just make sure to change the import path in `./cmd/mintcoin`.

First, make sure everything is working on your system, by running `make all` in this directory, this will update all dependencies, run the test quite, and install the `mintcoin` binary.  After that, you can run all commands with mintcoin.

## Setting Initial State

You first need to declare who the bankers are who can issue new coin. To do so, we make use of the `SetOption` abci command.  For debug purposes, we can run this over the `abci-cli`. When deployed as part of tendermint, you need to initialize this the same for all nodes, by passing in a [genesis file](https://github.com/tendermint/basecoin-examples/blob/master/mintcoin/cmd/mintcoin/main.go#L20) (this is different than a genesis block) upon starting the mintcoin binary.

If you register the plugin with the default name "mint", two options keys are supported - `mint/add` and `mint/remove`.  Both take a hex-encoded address as the second argument.  Once an address is added, the private key that belongs to that address can sign MintTx transactions, and thus create money.

## Minting Money

To create money, we need to create a [MintTx](https://github.com/tendermint/basecoin-examples/blob/master/mintcoin/mint_data.go#L39-L50) transaction, and then call Serialize() to get the app-specific tx bytes.  Then you must wrap it in a [basecoin AppTx](https://github.com/tendermint/basecoin/blob/master/types/tx.go#L154-L160), setting `Name` to "mint", and `Data` to the bytes returned by `Serialize`.  You can then sign this AppTx with the private key...

Confused yet, luckily there is a shiny, new cli thanks to Bucky. So all this crypto power is just a few keystrokes away....

## Testing with a CLI

Once we authorized some keys to mint cash, let's do it.  And send those shiny new bills to our friends.  But before playing with mintcoin, make sure you check out [the basecoin cli tutorial](../tutorial.md)

Setup tendermint:

```
tendermint init
tendermint unsafe_reset_all
```

Setup basecoin server with default genesis:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/mintcoin/data
mintcoin start --in-proc --mint-plugin --counter-plugin
```

Run basecoin client in another window:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/mintcoin/data
mintcoin account D397BC62B435F3CF50570FBAB4340FE52C60858F

# apparently we need to send some coin with every transaction, even if the
# app doesn't care.  this also means the sending account must have at least
# one of some currency in order to mint the first coins.

# also note that we must specify the chain_id, as we are not using


mintcoin apptx --chain_id mint_chain_id --amount 1 mint --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 1000 --mintcoin BTC

mintcoin apptx --chain_id mint_chain_id --amount 1 mint --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 5 --mintcoin cosmo

mintcoin apptx --chain_id mint_chain_id --amount 1 mint --mintto D397BC62B435F3CF50570FBAB4340FE52C60858F --mint 5000 --mintcoin FOOD

mintcoin account D397BC62B435F3CF50570FBAB4340FE52C60858F

# let's stop being greedy and pay back someone else

mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301
mintcoin apptx --chain_id mint_chain_id --amount 1 mint --mintto 4793A333846E5104C46DD9AB9A00E31821B2F301 --mint 1234 --mintcoin BTC
mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301

# and they can give us a little kickback
mintcoin sendtx --chain_id mint_chain_id --from priv_validator2.json --to D397BC62B435F3CF50570FBAB4340FE52C60858F --amount 333 --coin BTC
mintcoin account 4793A333846E5104C46DD9AB9A00E31821B2F301
mintcoin account D397BC62B435F3CF50570FBAB4340FE52C60858F
```

## Attaching a GUI

**TODO** showcase matt's ui and examples of how to extend it?
