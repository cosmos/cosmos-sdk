# Accounts, Fees and Gas

## Pre-requisite Reading

- [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes the in-built accounts system of the Cosmos SDK, as well as the default strategies to handle fees and gas within a Cosmos SDK application.

- [Accounts](#accounts)
    + [KeyBase](#keybase)
    + [Addresses and Pubkeys](#addresses-and-pubkeys)
        * [Pubkeys](#pubkeys)
        * [Addresses](#addresses)
- [Fees and Gas](#fees-and-gas)
    + [Gas Meter](#gas-meter)
        * [Main Gas Meter](#main-gas-meter)
        * [Block Gas Meter](#block-gas-meter)
    + [AnteHandler](#antehandler)

## Accounts

In the Cosmos SDK, an *account* designates a pair of *public key* `PubKey` and *private key* `PrivKey`. The `PubKey` can be derived to generate various `Addresses`, which are used to identify users (among other parties) in the application. `Addresses` are also associated with [`message`s](../building-modules/messages-and-queries.md#messages) to identify the sender of the `message`. The `PrivKey` is used to generate [digital signatures](#signatures) to prove that an `Address` associated with the `PrivKey` approved of a given `message`. 

To derive `PubKey`s and `PrivKey`s, the Cosmos SDK uses a standard called [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki). This standard defines how to build an HD wallet, where a wallet is a set of accounts. At the core of every account, there is a seed, which takes the form of a 12 or 24-words mnemonic. From this mnemonic, it is possible to derive any number of `PrivKey`s using one-way cryptographic function. Then, a `PubKey` can be derived from the `PrivKey`. Naturally, the mnemonic is the most sensitive information, as private keys can always be re-generated if the mnemonic is preserved. 

```
     Account 0                         Account 1                         Account 2

+------------------+              +------------------+               +------------------+
|                  |              |                  |               |                  |
|    Address 0     |              |    Address 1     |               |    Address 2     |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Public key 0    |              |  Public key 1    |               |  Public key 2    |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Private key 0   |              |  Private key 1   |               |  Private key 2   |
|        ^         |              |        ^         |               |        ^         |
+------------------+              +------------------+               +------------------+
         |                                 |                                  |
         |                                 |                                  |
         |                                 |                                  |
         +--------------------------------------------------------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Master PrivKey   |
                                 |                   |
                                 +-------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Mnemonic (Seed)  |
                                 |                   |
                                 +-------------------+
```

In the Cosmos SDK, accounts are stored and managed via an object called a [`Keybase`](#keybase).

### Keybase

A `Keybase` is an object that stores and manages accounts. In the Cosmos SDK, a `Keybase` implementation follows the [`Keybase` interface](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/types.go#L14-L60):

```go
type Keybase interface {
	// CRUD on the keystore
	List() ([]Info, error)
	Get(name string) (Info, error)
	GetByAddress(address types.AccAddress) (Info, error)
	Delete(name, passphrase string, skipPass bool) error

	// Sign some bytes, looking up the private key to use
	Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)

	// CreateMnemonic creates a new mnemonic, and derives a hierarchical deterministic
	// key from that.
	CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)

	// CreateAccount creates an account based using the BIP44 path (44'/118'/{account}'/0/{index}
	CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)

	// Derive computes a BIP39 seed from th mnemonic and bip39Passwd.
	// Derive private key from the seed using the BIP44 params.
	// Encrypt the key to disk using encryptPasswd.
	// See https://github.com/cosmos/cosmos-sdk/issues/2095
	Derive(name, mnemonic, bip39Passwd, encryptPasswd string, params hd.BIP44Params) (Info, error)

	// CreateLedger creates, stores, and returns a new Ledger key reference
	CreateLedger(name string, algo SigningAlgo, hrp string, account, index uint32) (info Info, err error)

	// CreateOffline creates, stores, and returns a new offline key reference
	CreateOffline(name string, pubkey crypto.PubKey) (info Info, err error)

	// CreateMulti creates, stores, and returns a new multsig (offline) key reference
	CreateMulti(name string, pubkey crypto.PubKey) (info Info, err error)

	// The following operations will *only* work on locally-stored keys
	Update(name, oldpass string, getNewpass func() (string, error)) error
	Import(name string, armor string) (err error)
	ImportPrivKey(name, armor, passphrase string) error
	ImportPubKey(name string, armor string) (err error)
	Export(name string) (armor string, err error)
	ExportPubKey(name string) (armor string, err error)
	ExportPrivKey(name, decryptPassphrase, encryptPassphrase string) (armor string, err error)

	// ExportPrivateKeyObject *only* works on locally-stored keys. Temporary method until we redo the exporting API
	ExportPrivateKeyObject(name string, passphrase string) (crypto.PrivKey, error)

	// CloseDB closes the database.
	CloseDB()
}
```

The default implementation of `Keybase` of the Cosmos SDK is [`dbKeybase`](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/keybase.go). A few notes on the `Keybase` methods as implemented in `dbKeybase`:

- `Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)` strictly deals with the signature of the `message` bytes. Some preliminary work should be done beforehand to prepare and encode the `message`  into a canonical `[]byte` form. See an example of `message` preparation from the `auth` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/txbuilder.go#L177-L207). Note that signature verification is not implemented in the SDK by default. It is deferred to the [`anteHandler`](#antehandler).
- `CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)` creates a new mnemonic and prints it in the logs, but it **does not persist it on disk**. 
- `CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk (note that the `PrivKey` is [encrypted with a passphrase before being persisted](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/mintkey/mintkey.go), it is **never stored unencrypted**). In the context of this method, the `account` and `address` parameters refer to the segment of the BIP44 derivation path (e.g. `0`, `1`, `2`, ...) used to derive the `PrivKey` and `PubKey` from the mnemonic (note that given the same mnemonic and `account`, the same `PrivKey` will be generated, and given the same `account` and `address`, the same `PubKey` and `Address` will be generated). Finally, note that the `CreateAccount` method derives keys and addresses using `secp256k1` as implemented in the [Tendermint library](https://github.com/tendermint/tendermint/blob/master/crypto/secp256k1). As a result, it only works for creating account keys and addresses, not consensus keys. See [`Addresses`](#addresses) for more.

The current implementation of `dbKeybase` is basic and does not offer on-demand locking. If an instance of `dbKeybase` is created, the underlying `db` is locked meaning no other process can access it besides the one in which it was instantiated. This is the reason why the [default SDK client](https://github.com/cosmos/cosmos-sdk/tree/master/client/keys) uses another implementation of the `Keybase` interface called [`lazyKeybase`](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/lazy_keybase.go). `lazyKeybase` is simple wrapper around `dbKeybase` which locks the database only when operations are to be performed and unlocks it immediately after:

```go
// Example Get method of lazyKeybase

func (lkb lazyKeybase) Get(name string) (Info, error) {
	db, err := sdk.NewLevelDB(lkb.name, lkb.dir)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return newDbKeybase(db).Get(name)
}
```

With the `lazyKeybase`, it is possible for the [command-line interface](../interfaces/cli.md) to create a new account while the [rest server](../interfaces/rest.md) is running. It is also possible to pipe multiple CLI commands. 

### Addresses and PubKeys

`Addresses` and `PubKey`s are both public information that identify actors in the application. There are 3 main types of `Addresses`/`PubKeys` available by default in the Cosmos SDK:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`secp256k1`** curve. 
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`secp256k1`** curve. 
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve       | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-------------|---------------------|--------------------|
| Accounts           | cosmos                | cosmospub            | `secp256k1` | `20`                | `33`               |
| Validator Operator | cosmosvaloper         | cosmosvaloperpub     | `secp256k1` | `20`                | `33`               |
| Consensus Nodes    | cosmosvalcons         | cosmosvalconspub     | `ed25519`   | `20`                | `32`               | 

#### PubKeys

`PubKey`s used in the Cosmos SDK follow the [`Pubkey` interface defined in [tendermint's `crypto` package](https://github.com/tendermint/tendermint/blob/master/crypto/crypto.go#L22-L27):


```go
type PubKey interface {
	Address() Address
	Bytes() []byte
	VerifyBytes(msg []byte, sig []byte) bool
	Equals(PubKey) bool
}
```

For `secp256k1` keys, the actual implementation can be found [here](https://github.com/tendermint/tendermint/blob/master/crypto/secp256k1/secp256k1.go#L140). For `ed25519` keys, it can be found [here](https://github.com/tendermint/tendermint/blob/master/crypto/ed25519/ed25519.go#L135). 

Note that in the Cosmos SDK, `Pubkeys` are not manipulated in their raw form. Instead, they are double encoded using [`Amino`](../core/encoding.md#amino) and [`bech32`](https://en.bitcoin.it/wiki/Bech32). In the SDK is done by first calling the `Bytes()` method on the raw `Pubkey` (which applies amino encoding), and then the `ConvertAndEncode` method of `bech32`. You can check out the implementation [here](https://github.com/cosmos/cosmos-sdk/blob/master/types/address.go#L579-L729). 

#### Addresses

The Cosmos SDK comes by default with 3 types of addresses:

- `AccAddress` for accounts.
- `ValAddress` for validator operators. 
- `ConsAddress` for validator nodes. 

Each of these address types are an alias for an hex-encoded `[]byte` array of length 20. Here is the standard way to obtain an address `aa` from a `Pubkey pub`:

```go
aa := sdk.AccAddress(pub.Address().Bytes())
```

These addresses implement the [`Address` interface](https://github.com/cosmos/cosmos-sdk/blob/master/types/address.go#L72-L80):

```go
type Address interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}
```

Of note, the `Marhsal()` and `Bytes()` method both return the same raw `[]byte` form of the address, the former being needed for Protobuff compatibility. Also, the `String()` method is used to return the `bech32` encoded form of the address, which should be the only address format with which end-user interract. See an example [here](https://github.com/cosmos/cosmos-sdk/blob/master/types/address.go#L230-L243). 

## Fees and Gas

In the Cosmos SDK, `gas` is a special unit that is used to track the consumption of resources during execution. `gas` is typically consumed whenever read and writes are made to the store, but it can also be consumed if expensive computation needs to be done. It serves two main purposes:

- Make sure blocks are not consuming too many resources and will be finalized. This is implemented by default in the SDK via the [block gas meter](#block-gas-meter).
- Prevent spam and abuse from end-user. To this end, `gas` consumed during [`message`](../building-modules/messages-and-queries.md#messages) execution is typically priced, resulting in a `fee` (`fees = gas * gas-prices`). This fee generally has to be paid by the sender of the `message`. Note that the SDK does not enforce `gas` pricing by default, as there may be other ways to prevent spam (e.g. bandwidth schemes). Still, most applications will implement `fee` mechanisms to prevent spam. This is done via the [`AnteHandler`](#antehandler).  

### Gas Meter

In the Cosmos SDK, `gas` is a simple alias for `uint64`, and is managed by an object called a *gas meter*. Gas meters implement the [`GasMeter` interface](https://github.com/cosmos/cosmos-sdk/blob/master/store/types/gas.go#L32-L44):

```go
type GasMeter interface {
	GasConsumed() Gas
	GasConsumedToLimit() Gas
	Limit() Gas
	ConsumeGas(amount Gas, descriptor string)
	IsPastLimit() bool
	IsOutOfGas() bool
}
```

where:

- `GasConsumed()` returns the amount of gas that was consumed by the gas meter instance.
- `GasConsumedToLimit()` returns the amount of gas that was consumed by gas meter instance, or the limit if it is reached.
- `Limit()` returns the limit of the gas meter instance. `0` if the gas meter is infinite. 
- `ConsumeGas(amount Gas, descriptor string)` consumes the amount of `gas` provided. If the `gas` overflows, it panics with the `descriptor` message. If the gas meter is not infinite, it panics if `gas` consumed goes above the limit. 
- `IsPastLimit()` returns `true` if the amount of gas consumed by the gas meter instance is strictly above the limit, `false` otherwise. 
- `IsOutOfGas()` returns `true` if the amount of gas consumed by the gas meter instance is above or equal to the limit, `false` otherwise.

The gas meter is generally held in [`ctx`](../core/context.md), and consuming gas is done with the following pattern:

```go
ctx.GasMeter().ConsumeGas(amount, "description")
```

By default, the Cosmos SDK makes use of two different gas meters, the [main gas meter](#main-gas-metter[) and the [block gas meter](#block-gas-meter). 

#### Main Gas Meter

`ctx.GasMeter()` is the main gas meter of the application. The main gas meter is initialized in `BeginBlock` via `setDeliverState`, and then tracks gas consumption during execution sequences that lead to state-transitions, i.e. those originally triggered by [`BeginBlock`](../core/baseapp.md#beginblock), [`DeliverTx`](../core/baseapp.md#delivertx) and [`EndBlock`](../core/baseapp.md#endblock). At the beginning of each `DeliverTx`, the main gas meter **must be set to 0** in the [`AnteHandler`](#antehandler), so that it can track gas comsumption per-transaction. 

Gas comsumption can be done manually, generally by the module developer in the [`BeginBlocker`, `EndBlocker`](../building-modules/beginblock-endblock.md) or [`handler`](../building-modules/handler.md), but most of the time it is done automatically whenever there is a read or write to the store. This automatic gas consumption logic is implemented in a special store called [`GasKv`](../core/store.md#gaskv-store). 

#### Block Gas Meter

`ctx.BlockGasMeter()` is the gas meter used to track gas consumption per block and make sure it does not go above a certain limit. A new instance of the `BlockGasMeter` is created each time [`BeginBlock`](../core/baseapp.go#beginblock) is called. The `BlockGasMeter` is finite, and the limit of gas per block is defined in the application's consensus parameters. 

When a new [transaction](../core/transaction.md) is being processed via `DeliverTx`, the current value of `BlockGasMeter` is checked to see if it is above the limit. If it is, `DeliverTx` returns immediately. This can happen even with the first transaction in a block, as `BeginBlock` itself can consume gas. If not, the transaction is processed normally. At the end of `DeliverTx`, the gas tracked by `ctx.BlockGasMeter()` is increased by the amount consumed to process the transaction:

```go
ctx.BlockGasMeter().ConsumeGas(
				ctx.GasMeter().GasConsumedToLimit(),
				"block gas meter",
			)
```

### AnteHandler

The `AnteHandler` is a special `handler` that is run for every transaction during `CheckTx` and `DeliverTx`, before the `handler` of each `message` in the transaction. `AnteHandler` have a different signature than `handler`s:

```go
// AnteHandler authenticates transactions, before their internal messages are handled.
// If newCtx.IsZero(), ctx is used instead.
type AnteHandler func(ctx Context, tx Tx, simulate bool) (newCtx Context, result Result, abort bool)
```

Being a `handler`, the `anteHandler` is not implemented in the core SDK but in a module. This gives the possibility to developers to choose the version of `AnteHandler` that fits their application's needs. That said, most applications today use the default implementation defined in the [`auth` module](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth). Here is what the `anteHandler` is intended to do in a normal Cosmos SDK application:

- Verify that the transaction are of the correct type. Transaction types are defined in the module that implements the `anteHandler`, and they follow the [transaction interface](https://github.com/cosmos/cosmos-sdk/blob/master/types/tx_msg.go#L34-L41). This enables developers to play with various types for the transaction of their application. In the default `auth` module, the standard transaction type is [`StdTx`](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/stdtx.go#L23-L28).
- Verify signatures for each [`message`](../building-modules/messages-and-queries.md#messages) contained in the transaction. Each `message` should be signed by one or multiple sender(s), and these signatures must be verified in the `anteHandler`. 
- During `CheckTx`, verify that the gas prices provided with the transaction is greater than the local `min-gas-prices` (as a reminder, gas-prices can be deducted from the following equation: `fees = gas * gas-prices`). `min-gas-prices` is a parameter local to each full-node and used during `CheckTx` to discard transactions that do not provide a minimum amount of fees. This ensure that the mempool cannot be spammed with garbage transactions. 
- Verify that the sender of the transaction has enough funds to cover for the `fees`. When the end-user generates a transaction, they must indicate 2 of the 3 following parameters (the third one being implicit): `fees`, `gas` and `gas-prices`. This signals how much they are willing to pay for nodes to execute their transaction. The provided `gas` value is stored in a parameter called `GasWanted` for later use. 
- Set `newCtx.GasMeter` to 0, with a limit of `GasWanted`. **This step is extremely important**, as it not only makes sure the transaction cannot consume infinite gas, but also that `ctx.GasMeter` is reset in-between each `DeliverTx` (`ctx` is set to `newCtx` after `anteHandler` is run). 

As explained above, the `anteHandler` returns a maximum limit of `gas` the transaction can consume during execution called `GasWanted`. The actual amount consumed in the end is denominated `GasUsed`, and we must therefore have `GasUsed =< GasWanted`. Both `GasWanted` and `GasUsed` are relayed to the underlying consensus engine when [`DeliverTx`](../core/baseapp.md#delivertx) returns. 

## Next

Learn about [baseapp](../core/baseapp.md). 