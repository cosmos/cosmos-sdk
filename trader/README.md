# Trader - let's play with finances

Once you have checkout out [mintcoin]('../mintcoin/README.md'), a simple plugin to allow minting money, with only global state and one transaction type, let's explore a more complicated example.  Trader will simulate some basic financial actions that are very common in the modern world, and show how we can add these features to basecoin, in our quest to make the best financial system ever.  Stable and secure and feature-rich, and super fast.  This is going to be so good.

## Escrow

This first instrument we implement is an [escrow](./escrow).  We can send money to an apptx to create an escrow.  Thereby we specify who the intended recipient is, and who can releae the money (or return it).  Note that we give this "arbiter" the power to send the money to the recipient or return it to the sender, but no way to take the money and put it in their own pocket.  Removing more locations for fraud.

When we create an escrow, we get a unique address back.  Save this.  Later on, the arbiter can send a message to this address to release the money to either the intended recipient, or back to the sender if they failed to deliver on their promise. And just in case, if too much time passed and no one did anything, you can always expire the escrow and recover the money (eg. if the arbiter lost their private key - ouch!).


## Testing with a CLI

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

## Attaching a GUI

**TODO** showcase matt's ui and examples of how to extend it?
