# Basecoin Example

Here we explain how to get started with a basic Basecoin blockchain, how
to send transactions between accounts using the ``basecli`` tool, and
what is happening under the hood.

## Setup and Install

You will need to have go installed on your computer. Please refer to the [cosmos testnet tutorial](https://cosmos.network/validators/tutorial), which will always have the most updated instructions on how to get setup with go and the cosmos repository. 

Once you have go installed, run the command:

```
go get github.com/cosmos/cosmos-sdk
```

There will be an error stating `can't load package: package github.com/cosmos/cosmos-sdk: no Go files`, however you can ignore this error, it doesn't affect us. Now change directories to:

```
cd $GOPATH/src/github.com/cosmos/cosmos-sdk
```

And run :

```
make get_tools // run make update_tools if you already had it installed
make get_vendor_deps
make install_examples
```
Then run `make install_examples`, which creates binaries for `basecli` and `basecoind`. You can look at the Makefile if you want to see the details on what these make commands are doing. 

## Using basecli and basecoind

Check the versions by running:

```
basecli version
basecoind version
```

They should read something like `0.17.1-5d18d5f`, but the versions will be constantly updating so don't worry if your version is higher that 0.17.1. That's a good thing.

Note that you can always check help in the terminal by running `basecli -h` or `basecoind -h`. It is good to check these out if you are stuck, because updates to the code base might slightly change the commands, and you might find the correct command in there. 

Let's start by initializing the basecoind daemon. Run the command

```
basecoind init
```

And you should see something like this:

```
{
  "chain_id": "test-chain-z77iHG",
  "node_id": "e14c5056212b5736e201dd1d64c89246f3288129",
  "app_message": {
    "secret": "pluck life bracket worry guilt wink upgrade olive tilt output reform census member trouble around abandon"
  }
}
```

This creates the `~/.basecoind folder`, which has config.toml, genesis.json, node_key.json, priv_validator.json. Take some time to review what is contained in these files if you want to understand what is going on at a deeper level. 


## Generating keys

The next thing we'll need to do is add the key from priv_validator.json to the gaiacli key manager. For this we need the 16 word seed that represents the private key, and a password. You can also get the 16 word seed from the output seen above, under `"secret"`. Then run the command:

```
basecli keys add alice --recover
```

Which will give you three prompts:

```
Enter a passphrase for your key:
Repeat the passphrase:
Enter your recovery seed phrase:
```

You just created your first locally stored key, under the name alice, and this account is linked to the private key that is running the basecoind validator node. Once you do this, the  ~/.basecli folder is created, which will hold the alice key and any other keys you make. Now that you have the key for alice, you can start up the blockchain by running 

```
basecoind start
```

You should see blocks being created at a fast rate, with a lot of output in the terminal.

Next we need to make some more keys so we can use the send transaction functionality of basecoin. Open a new terminal, and run the following commands, to make two new accounts, and give each account a password you can remember:

```
basecli keys add bob
basecli keys add charlie
```

You can see your keys with the command:

```
basecli keys list
```

You should now see alice, bob and charlie's account all show up. 

```
NAME: 	ADDRESS:					                PUBKEY:
alice   90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD	1624DE62201D47E63694448665F5D0217EA8458177728C91C373047A42BD3C0FB78BD0BFA7
bob     29D721F054537C91F618A0FDBF770DA51EF8C48D	1624DE6220F54B2A2CA9EB4EE30DE23A73D15902E087C09CC5616456DDDD3814769E2E0A16
charlie 2E8E13EEB8E3F0411ACCBC9BE0384732C24FBD5E	1624DE6220F8C9FB8B07855FD94126F88A155BD6EB973509AE5595EFDE1AF05B4964836A53
```


## Send transactions

Lets send bob and charlie some tokens. First, lets query alice's account so we can see what kind of tokens she has:

```
basecli account 90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD
```

Where `90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD` is alice's address we got from running `basecli keys list`. You should see a large amount of "mycoin" there. If you search for bob's or charlie's address, the command will fail, because they haven't been added into the blockchain database yet since they have no coins. We need to send them some!

The following command will send coins from alice, to bob:

```
basecli send --name=alice --amount=10000mycoin --to=29D721F054537C91F618A0FDBF770DA51EF8C48D 
--sequence=0 --chain-id=test-chain-AE4XQo
```

Flag Descriptions: 
- `name` is the name you gave your key
- `mycoin` is the name of the token for this basecoin demo, initialized in the genesis.json file
- `sequence` is a tally of how many transactions have been made by this account. Since this is the first tx on this account, it is 0
- `chain-id` is the unique ID that helps tendermint identify which network to connect to. You can find it in the terminal output from the gaiad daemon in the header block , or in the genesis.json file  at `~/.basecoind/config/genesis.json`

Now if we check bobs account, it should have `10000 mycoin`. You can do so by running :

```
basecli account 29D721F054537C91F618A0FDBF770DA51EF8C48D
```

Now lets send some from bob to charlie. Make sure you send less than bob has, otherwise the transaction will fail:

```
basecli send --name=bob --amount=5000mycoin --to=2E8E13EEB8E3F0411ACCBC9BE0384732C24FBD5E 
--sequence=0 --chain-id=test-chain-AE4XQo
```

Note how we use the ``--name`` flag to select a different account to send from.

Lets now try to send from bob back to alice:

```
basecli send --name=bob --amount=3000mycoin --to=90B0B9BE0914ECEE0B6DB74E67B07A00056B9BBD 
--sequence=1 --chain-id=test-chain-AE4XQo
```

Notice that the sequence is now 1, since we have already recorded bobs 1st transaction as `sequence 0`. Also note the ``hash`` value in the response  in the terminal - this is the hash of the transaction. We can query for the transaction with this command:

```
basecli tx <INSERT HASH HERE>
```

It will return the details of the transaction hash, such as how many coins were send and to which address, and on what block it occurred.

That is the basic implementation of basecoin!


## Reset the basecoind blockchain and basecli data

**WARNING:** Running these commands will wipe out any existing
information in both the ``~/.basecli`` and ``~/.basecoind`` directories,
including private keys. This should be no problem considering that basecoin
is just an example, but it is always good to pay extra attention when 
you are removing private keys, in any scenario involving a blockchain. 

To remove all the files created and refresh your environment (e.g., if
starting this tutorial again or trying something new), the following
commands are run:

```
basecoind unsafe_reset_all
rm -rf ~/.basecoind
rm -rf ~/.basecli
```

## Technical Details on how Basecoin Works

This section describes some of the more technical aspects for what is going on under the hood of Basecoin.

## Proof

Even if you don't see it in the UI, the result of every query comes with
a proof. This is a Merkle proof that the result of the query is actually
contained in the state. And the state's Merkle root is contained in a
recent block header. Behind the scenes, ``basecli`` will not only
verify that this state matches the header, but also that the header is
properly signed by the known validator set. It will even update the
validator set as needed, so long as there have not been major changes
and it is secure to do so. So, if you wonder why the query may take a
second... there is a lot of work going on in the background to make sure
even a lying full node can't trick your client.

## Accounts and Transactions

For a better understanding of how to further use the tools, it helps to
understand the underlying data structures, so lets look at accounts and transactions.

### Accounts

The Basecoin state consists entirely of a set of accounts. Each account
contains an address, a public key, a balance in many different coin denominations,
and a strictly increasing sequence number for replay protection. This
type of account was directly inspired by accounts in Ethereum, and is
unlike Bitcoin's use of Unspent Transaction Outputs (UTXOs). 

```
type BaseAccount struct {
  Address  sdk.Address   `json:"address"`
  Coins    sdk.Coins     `json:"coins"`
  PubKey   crypto.PubKey `json:"public_key"`
  Sequence int64         `json:"sequence"`
}
```

You can also add more fields to accounts, and basecoin actually does so. Basecoin
adds a Name field in order to show how easily the base account structure can be
modified to suit any applications needs. It takes the `auth.BaseAccount` we see above, 
and extends it with `Name`.

```
type AppAccount struct {
  auth.BaseAccount
  Name string `json:"name"`
}
```

Within accounts, coin balances are stored. Basecoin is a multi-asset cryptocurrency, so each account can have many
different kinds of tokens, which are held in an array.  

```
type Coins []Coin

type Coin struct {
  Denom  string `json:"denom"`
  Amount int64  `json:"amount"`
}
```

If you want to add more coins to a blockchain, you can do so manually in
the ``~/.basecoin/genesis.json`` before you start the blockchain for the
first time.

Accounts are serialized and stored in a Merkle tree under the key
``base/a/<address>``, where ``<address>`` is the address of the account.
Typically, the address of the account is the 20-byte ``RIPEMD160`` hash
of the public key, but other formats are acceptable as well, as defined
in the `Tendermint crypto
library <https://github.com/tendermint/go-crypto>`__. The Merkle tree
used in Basecoin is a balanced, binary search tree, which we call an
`IAVL tree <https://github.com/tendermint/iavl>`__.

### Transactions

Basecoin defines a transaction type, the `SendTx`, which allows tokens
to be sent to other accounts. The `SendTx` takes a list of inputs and
a list of outputs, and transfers all the tokens listed in the inputs
from their corresponding accounts to the accounts listed in the output.
The `SendTx` is structured as follows:
```
type SendTx struct {
  Gas     int64      `json:"gas"`
  Fee     Coin       `json:"fee"`
  Inputs  []TxInput  `json:"inputs"`
  Outputs []TxOutput `json:"outputs"`
}

type TxInput struct {
  Address   []byte           `json:"address"`   // Hash of the PubKey
  Coins     Coins            `json:"coins"`     //
  Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed TxInput
  Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
  PubKey    crypto.PubKey    `json:"pub_key"`   // Is present iff Sequence == 0
}

type TxOutput struct {
  Address []byte `json:"address"` // Hash of the PubKey
  Coins   Coins  `json:"coins"`   //
}
```
Note the `SendTx` includes a field for `Gas` and `Fee`. The
`Gas` limits the total amount of computation that can be done by the
transaction, while the `Fee` refers to the total amount paid in fees.
This is slightly different from Ethereum's concept of `Gas` and
`GasPrice`, where `Fee = Gas x GasPrice`. In Basecoin, the `Gas`
and `Fee` are independent, and the `GasPrice` is implicit. 

In Basecoin, the `Fee` is meant to be used by the validators to inform
the ordering of transactions, like in Bitcoin. And the `Gas` is meant
to be used by the application plugin to control its execution. There is
currently no means to pass `Fee` information to the Tendermint
validators, but it will come soon... so this version of Basecoin does 
not actually fully implement fees and gas, but it still allows us 
to send transactions between accounts. 

Note also that the `PubKey` only needs to be sent for
`Sequence == 0`. After that, it is stored under the account in the
Merkle tree and subsequent transactions can exclude it, using only the
`Address` to refer to the sender. Ethereum does not require public
keys to be sent in transactions as it uses a different elliptic curve
scheme which enables the public key to be derived from the signature
itself.

Finally, note that the use of multiple inputs and multiple outputs
allows us to send many different types of tokens between many different
accounts at once in an atomic transaction. Thus, the `SendTx` can
serve as a basic unit of decentralized exchange. When using multiple
inputs and outputs, you must make sure that the sum of coins of the
inputs equals the sum of coins of the outputs (no creating money), and
that all accounts that provide inputs have signed the transaction.

