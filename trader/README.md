# Trader - let's play with finances

Once you have checkout out [mintcoin]('../mintcoin/README.md'), a simple plugin to allow minting money, with only global state and one transaction type, let's explore a more complicated example.  Trader will simulate some basic financial actions that are very common in the modern world, and show how we can add these features to basecoin, in our quest to make the best financial system ever.  Stable and secure and feature-rich, and super fast.  This is going to be so good.

## Escrow

This first instrument we implement is an [escrow](./escrow).  We can send money to an apptx to create an escrow.  Thereby we specify who the intended recipient is, and who can releae the money (or return it).  Note that we give this "arbiter" the power to send the money to the recipient or return it to the sender, but no way to take the money and put it in their own pocket.  Removing more locations for fraud.

When we create an escrow, we get a unique address back.  Save this.  Later on, the arbiter can send a message to this address to release the money to either the intended recipient, or back to the sender if they failed to deliver on their promise. And just in case, if too much time passed and no one did anything, you can always expire the escrow and recover the money (eg. if the arbiter lost their private key - ouch!).


### Data structure

We create a separate data location for each escrow.  It is stored in the following format, where `Sender`, `Recipient`, and `Arbiter` are all addresses.  The value is determined from the money sent in creation, and the same amount is paid back when the escrow is resolved.  The address of the escrow is determined from the hash for the data bytes, to guarantee uniqueness.

```
// EscrowData is our principal data structure in the db
type EscrowData struct {
  Sender     []byte
  Recipient  []byte
  Arbiter    []byte
  Expiration uint64 // height when the offer expires (0 = never)
  Amount     types.Coins
}
```

There are two basic operations one can perform on an escrow - creating it, and resolving it.  Resolving it can be done either by a clear decision of the arbiter, or by a simple expiration.  Thus, there are three transaction types for these two concepts.

### Testing with a CLI

By this point, you have probably played with the basecoin-based cli a few times and are feeling real comfortable-like.  So, let's just jump right in with a few commands:

Setup basecoin server with default genesis:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/trader
make all
tendermint unsafe_reset_all
cd data
trader start --in-proc --escrow-plugin
```

Run basecoin client in another window.  In this example, priv_validator.json will be the sender, priv_validator2.json the arbiter.  And some empty account the receiver.

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/trader/data

# check the three accounts
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # sender
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # arbiter
trader account 2ABAA2CCFA1F618CF9C97F1FD59FC3EE4968FE8A   # receiver

# let's make an escrow
trader apptx --chain_id trader_chain_id --from priv_validator.json --amount 400 escrow create --recv 2ABAA2CCFA1F618CF9C97F1FD59FC3EE4968FE8A --arbiter 4793A333846E5104C46DD9AB9A00E31821B2F301

#-> TODO: need to get ESCROW_ID locally, broadcastTx response....
ESCROW_ID=9D2C197899F922359D7AB13D18123B2749077FB8

# fails cuz the sender cannot release funds
trader apptx --chain_id trader_chain_id --from priv_validator.json --amount 1 escrow pay --escrow $ESCROW_ID

# note, that it didn't cost anything to fail :)
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # sender

# succeeds as the arbiter can
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 1 escrow pay --escrow $ESCROW_ID

# but you pay the fees when the call works (1 blank here)
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # arbiter

# and the money was sent
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # sender
trader account 2ABAA2CCFA1F618CF9C97F1FD59FC3EE4968FE8A   # receiver

# but an error the second time the arbiter tries to send the same money (no re-entrant contracts)
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 1 escrow pay --escrow $ESCROW_ID

# TODO: let's demo expiry and more

# ASIDE: digging in with a debugger....
dlv debug ../cmd/trader/main.go -- apptx --chain_id trader_chain_id --from priv_validator.json --amount 400 escrow create --recv 2ABAA2CCFA1F618CF9C97F1FD59FC3EE4968FE8A --arbiter 4793A333846E5104C46DD9AB9A00E31821B2F301
```

## Currency Options

Moving on to a more complex example, we will create an [currency option](./options).  This is the option to buy one set of coins (eg. 100 ETH) for another set of coins (eg. 2 BTC).  There are two parties in the option - the issuer and the holder.  The issuer bonds a certain set of Coin in the option and sets the price.  The holder is the account that has the right to exercise the option, that is send the trade value, which goes to the original issuer, while the bonded value is released to the holder.  If the option is not used in a given time, the bond returns to the issuer.

On first glance, this is a similar set up to escrow, bonded coins that can be released by another transaction.  However, there is one additional step.  The option can be bought and sold without exercising it.  That is the holder can transfer the option to a new holder in return for some coin.  And this transfer operation should be done atomically, without room for one party cheating.

### Data structure

We create a separate data location for each option.  It is stored in the following format, where `Sender`, `Recipient`, and `Arbiter` are all addresses.  The value is determined from the money sent in creation, and the same amount is paid back when the escrow is resolved.  The address of the escrow is determined from the hash for the data bytes, to guarantee uniqueness.

```
// OptionData is our principal data structure in the db
type OptionData struct {
  OptionIssue
  OptionHolder
}

// OptionIssue is the constant part, created wth the option, never changes
type OptionIssue struct {
  // this is for the normal option functionality
  Issuer     []byte
  Serial     int64       // this serial number is from the apptx that created it
  Expiration uint64      // height when the offer expires (0 = never)
  Bond       types.Coins // this is stored upon creation of the option
  Trade      types.Coins // this is the money that can exercise the option
}

// OptionHolder is the dynamic section of who can excercise the options
type OptionHolder struct {
  // this is for buying/selling the option (should be a separate struct?)
  Holder    []byte
  NewHolder []byte      // set to allow for only one buyer, empty for any buyer
  Price     types.Coins // required payment to transfer ownership
}
```

We can perform the following actions on an option:

* Create the option (sending Bond to the apptx, Holder=Issuer at first)
* Offer the option for sale (Specifying Price)
* Purchase the option (by sending Price to apptx, changes Holder)
* Exercise the option (only the holder can do by sending Trade to apptx)
* Disolve the option (either at expiration, or if Issuer=Holder and wants the Bond back)

Thus, we need a transaction type for each of these actions

### Code Design

I have attempted to abstract out some common patterns by this point in time. These patterns could be useful for anyone else attempting to build a basecoin plugin, so I will cover them briefly...

1. When working with a plugin, generally we prefix all Set/Get queries for the plugin-specific data with a standard prefix (eg. the plugin name). In order to avoid this boiler-plater, you can just wrap the store with a [PrefixStore](./prefix_store.go#L10), and then all `Get` and `Set` methods are automatically prefixed with the key.
1. The only non-local data we generally should touch is the `Accounts` themselves to update the balances.  We can wrap the `KVStore` with an [Accountant](./options/util.go#L9), to easily `GetAccount` and `SetAccount`, as well as `Refund` all coins sent on the transaction for errors, or `Pay` some coins to a given address.
1. We define a plugin-specific [transaction type](./options/data.go#L113-L119) along with writing and parsing them.  We can then just [register all supported transactions](./options/data.go#L13-L23) in the init function, and all parsing is taking care of.
1. The `Plugin` itself simply tracks the height, parses the transactions, and [delegates the work](./options/plugin.go#L37-L50) to the transactions themselves. The real code is in the transactions, which have a special interface to get all data they need.
1. Each new transaction you want to support, simply involves creating the [data structure](./options/tx.go#L11-L15), implementing the [Apply method]((./options/tx.go#L17-L42)), and [registering it](./options/data.go#L17) with go wire.
1. You can add a special [command for the plugin](./commands/options.go#L52-L63), with [subcommands for each transaction](./commands/options.go#L65-L76), and then a simple [parsing of args to data structure](./commands/options.go#L146-L161).  This is not required for the plugin to function, but with a little work, you can [integrate it](./commands/options.go#L140-L144) in the [basecoin cli](./cmd/trader/main.go#L7), and allow a much better workflow to debug and demo it until you have a gui.
1. Make sure all data has a unique, deterministic, constant address.  To do so, we can hash all or part of the data.  To make it constant, only hash the constant part of the data (`OptionIssue`), the mutable parts of the data (`OptionHolder`) should not be included in this hash for obvious reasons. To guarantee uniqueness (should the same user issue the same command twice), it is nice to include the sequence number of the create transaction as part of the immutible section.

Using these patterns should allow you to perform most actions you reasonably wish to perform with your basecoin plugin, while removing much boilerplate and bit switching.  It is also very extensible if you wish to add a new transaction type.  And allows easily setting up the scaffolding for [unit tests](./options/tx_test.go#L12-L43).

### Testing with a CLI

You know the deal by now, so....

In one window start the server:

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/trader
make all
tendermint unsafe_reset_all
cd data
trader start --in-proc --options-plugin
# dlv debug ../cmd/trader/main.go -- start --in-proc --options-plugin
```

Run basecoin client in another window.  In this example, priv_validator.json will be the issuer, priv_validator2.json the holder.

```
cd $GOPATH/src/github.com/tendermint/basecoin-examples/trader/data

# check the two accounts
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # issuer
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # holder

# let's make an option
trader apptx --chain_id trader_chain_id --from priv_validator.json --coin ETH --amount 400 options create --trade 4 --trade-coin BTC
# dlv debug ../cmd/trader/main.go -- apptx --chain_id trader_chain_id --from priv_validator.json --coin ETH --amount 400 options create --trade 4 --trade-coin BTC


#-> TODO: need to get OPTION_ID locally, broadcastTx response....
OPTION_ID=XXXXXXX
trader apptx --chain_id trader_chain_id options query $OPTION_ID

# we cannot exercise it cuz the we do not own the option yet
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 4 --coin BTC options exercise --option $OPTION_ID

# note, that it didn't cost anything to fail :) no 4 BTC loss....
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # sender

# so, let us offer this for sale (only the current holder can)
# also note this money is not used up (just needs to be non-zero to prevent spaming)
trader apptx --chain_id trader_chain_id --from priv_validator.json --amount 10 --coin ETH options sell --option $OPTION_ID --price 100 --price-coin blank

# and now the holder can buy the rights to the option.
# the money is used up to the price level (overpayment returned)
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 250 --coin blank options buy --option $OPTION_ID

# check the two accounts
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # issuer
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # holder

# notice the issuer is down the 400 ETH stored as a bond in the option
# and notice the only other change is the 100 blank payment from holder to issuer for ownership of the option
# all other coin is returned untouched after the transaction

# and now for the real trick, let's use this option
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 2 --coin BTC options exercise --option $OPTION_ID

# wait... it only works if you send the required amount
trader apptx --chain_id trader_chain_id --from priv_validator2.json --amount 4 --coin BTC options exercise --option $OPTION_ID

# now, look at this, the issuer got the 4 BTC, the holder the 400 ETH
# and we can even trade the rights to perform this operation :)
trader account D397BC62B435F3CF50570FBAB4340FE52C60858F  # issuer
trader account 4793A333846E5104C46DD9AB9A00E31821B2F301  # holder

# and the option has now disappeared, so you can't use it again
trader apptx --chain_id trader_chain_id options query $OPTION_ID
```

This is just the start.  There is also some methods for expiration and "disolving" the option if it will not be used.  You could add a UI to enable market trading and just do the resolution on the blockchain.  Or spend a couple days and build even more complex instruments, complete with unit tests and safety, so you don't go bankrupt from a bug.

## Attaching a GUI

**TODO** showcase matt's ui and examples of how to extend it?
