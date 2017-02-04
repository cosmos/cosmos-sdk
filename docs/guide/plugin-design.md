# Basecoin Plugins

Basecoin is an extensible cryptocurrency module.
Each Basecoin account contains a ED25519 public key,
a balance in many different coin denominations,
and a strictly increasing sequence number for replay protection (like in Ethereum).
Accounts are serialized and stored in a merkle tree using the account's address as the key,
where the address is the RIPEMD160 hash of the public key.

Sending tokens around is done via the `SendTx`, which takes a list of inputs and a list of outputs,
and transfers all the tokens listed in the inputs from their corresponding accounts to the accounts listed in the output.
The `SendTx` is structured as follows:

```
type SendTx struct {
  Gas     int64      `json:"gas"` // Gas
  Fee     Coin       `json:"fee"` // Fee
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

type Coins []Coin

type Coin struct {
  Denom  string `json:"denom"`
  Amount int64  `json:"amount"`
}

```

Note it also includes a field for `Gas` and `Fee`. The `Gas` limits the total amount of computation that can be done by the transaction,
while the `Fee` refers to the total amount paid in fees. This is slightly different from Ethereum's concept of `Gas` and `GasPrice`,
where `Fee = Gas x GasPrice`. In Basecoin, the `Gas` and `Fee` are independent.


Basecoin also defines another transaction type, the `AppTx`:

