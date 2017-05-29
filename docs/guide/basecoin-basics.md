# Basecoin Basics

Here we explain how to get started with a simple Basecoin blockchain, 
how to send transactions between accounts using the `basecoin` tool,
and what is happening under the hood.

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

Now we are ready to send some transactions. First, open another window.
If you take a look at the `~/.basecoin/genesis.json` file, you will see one account listed under the `app_options`.
This account corresponds to the private key in `~/.basecoin/key.json`.
We also included the private key for another account, in `~/.basecoin/key2.json`.

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

For a better understanding of the options, it helps to understand the underlying data structures.

## Accounts

The Basecoin state consists entirely of a set of accounts.
Each account contains a public key,
a balance in many different coin denominations,
and a strictly increasing sequence number for replay protection.
This type of account was directly inspired by accounts in Ethereum,
and is unlike Bitcoin's use of Unspent Transaction Outputs (UTXOs).
Note Basecoin is a multi-asset cryptocurrency, so each account can have many different kinds of tokens.

```golang
type Account struct {
	PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
	Sequence int           `json:"sequence"`
	Balance  Coins         `json:"coins"`
}

type Coins []Coin

type Coin struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}
```

Accounts are serialized and stored in a Merkle tree under the key `base/a/<address>`, where `<address>` is the address of the account.
Typically, the address of the account is the 20-byte `RIPEMD160` hash of the public key, but other formats are acceptable as well,
as defined in the [Tendermint crypto library](https://github.com/tendermint/go-crypto).
The Merkle tree used in Basecoin is a balanced, binary search tree, which we call an [IAVL tree](https://github.com/tendermint/go-merkle).

## Transactions

Basecoin defines a simple transaction type, the `SendTx`, which allows tokens to be sent to other accounts.
The `SendTx` takes a list of inputs and a list of outputs,
and transfers all the tokens listed in the inputs from their corresponding accounts to the accounts listed in the output.
The `SendTx` is structured as follows:

```golang
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

Note the `SendTx` includes a field for `Gas` and `Fee`.
The `Gas` limits the total amount of computation that can be done by the transaction,
while the `Fee` refers to the total amount paid in fees.
This is slightly different from Ethereum's concept of `Gas` and `GasPrice`,
where `Fee = Gas x GasPrice`. In Basecoin, the `Gas` and `Fee` are independent,
and the `GasPrice` is implicit.

In Basecoin, the `Fee` is meant to be used by the validators to inform the ordering 
of transactions, like in Bitcoin.  And the `Gas` is meant to be used by the application 
plugin to control its execution.  There is currently no means to pass `Fee` information 
to the Tendermint validators, but it will come soon...

Note also that the `PubKey` only needs to be sent for `Sequence == 0`.
After that, it is stored under the account in the Merkle tree and subsequent transactions can exclude it,
using only the `Address` to refer to the sender. Ethereum does not require public keys to be sent in transactions
as it uses a different elliptic curve scheme which enables the public key to be derived from the signature itself.

Finally, note that the use of multiple inputs and multiple outputs allows us to send many 
different types of tokens between many different accounts at once in an atomic transaction. 
Thus, the `SendTx` can serve as a basic unit of decentralized exchange. When using multiple 
inputs and outputs, you must make sure that the sum of coins of the inputs equals the sum of 
coins of the outputs (no creating money), and that all accounts that provide inputs have signed the transaction.

## Conclusion

In this guide, we introduced the `basecoin` tool, demonstrated how to use it to send tokens between accounts,
and discussed the underlying data types for accounts and transactions, specifically the `Account` and the `SendTx`.
In the [next guide](basecoin-plugins.md), we introduce the basecoin plugin system, which uses a new transaction type, the `AppTx`,
to extend the functionality of the Basecoin system with arbitrary logic.
