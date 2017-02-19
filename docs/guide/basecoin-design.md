# Basecoin Design

Basecoin is designed to be a simple cryptocurrency application with limited built-in functionality,
but with the capacity to be extended by arbitrary plugins.
Its basic data structures are inspired by Ethereum, but it is much simpler, as there is no built-in virtual machine.

## Accounts

The Basecoin state consists entirely of a set of accounts.
Each account contains an ED25519 public key,
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

Accounts are serialized and stored in a Merkle tree using the account's address as the key,
In particular, an account is stored in the Merkle tree under the key `base/a/<address>`,
where `<address>` is the address of the account.
In Basecoin, the address of an account is the 20-byte `RIPEMD160` hash of the public key.
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

There are a few things to note. First, the `SendTx` includes a field for `Gas` and `Fee`.
The `Gas` limits the total amount of computation that can be done by the transaction,
while the `Fee` refers to the total amount paid in fees.
This is slightly different from Ethereum's concept of `Gas` and `GasPrice`,
where `Fee = Gas x GasPrice`. In Basecoin, the `Gas` and `Fee` are independent,
and the `GasPrice` is implicit.

In Tendermint, the `Fee` is meant to be used by the validators to inform the ordering 
of transactions, like in bitcoin.  And the `Gas` is meant to be used by the application 
plugin to control its execution.  There is currently no means to pass `Fee` information 
to the Tendermint validators, but it will come soon...

Second, notice that the `PubKey` only needs to be sent for `Sequence == 0`.
After that, it is stored under the account in the Merkle tree and subsequent transactions can exclude it,
using only the `Address` to refer to the sender. Ethereum does not require public keys to be sent in transactions
as it uses a different elliptic curve scheme which enables the public key to be derived from the signature itself.

Finally, note that the use of multiple inputs and multiple outputs allows us to send many 
different types of tokens between many different accounts at once in an atomic transaction. 
Thus, the `SendTx` can serve as a basic unit of decentralized exchange. When using multiple 
inputs and outputs, you must make sure that the sum of coins of the inputs equals the sum of 
coins of the outputs (no creating money), and that all accounts that provide inputs have signed the transaction.

## Plugins

Basecoin actually defines a second transaction type, the `AppTx`,
which enables the functionality to be extended via custom plugins.
To learn more about the `AppTx` and plugin system, see the [plugin design document](plugin-design.md).
To implement your first plugin, see [plugin tutorial](example-plugin.md).
