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

## Options

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

**TODO** implement this fully

## Attaching a GUI

**TODO** showcase matt's ui and examples of how to extend it?
