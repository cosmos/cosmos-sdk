# Accounts, Fees and Gas

## Pre-requisite Reading

- [Anatomy of an SDK Application](./app-anatomy.md)

## Synopsis

This document describes the in-built accounts system of the Cosmos SDK, as well as the default strategies to handle fees and gas within a Cosmos SDK application.

- [Accounts](#accounts)
    + [KeyBase](#keybase)
    + [Addresses](#addresses)
    + [Signatures](#signatures)
- [Fees and Gas](#fees-and-gas)
    + [AnteHandler](#antehandler)
    + [Gas](#gas)
    + [Gas Meter](#gas-meter)
    + [Block Gas Meter](#block-gas-meter)

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

- `Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)` strictly deals with the signature of the `message` bytes. Some preliminary work should be done beforehand to prepare and encode the `message`  into a canonical `[]byte` form. See an example of `message` preparation from the `auth` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth/types/txbuilder.go#L177-L207).
- `CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)` creates a new mnemonic and prints it in the logs, but it **does not persist it on disk**. 
- `CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk (note that the `PrivKey` is [encrypted with a passphrase before being persisted](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/mintkey/mintkey.go), it is **never stored unencrypted**). In the context of this method, the `account` parameter refers to the derivation path (e.g. `0`, `1`, `2`, ...) used to derive the `PrivKey` from the mnemonic (note that given the same mnemonic and `account`, `CreateAccount` will always return the same `PrivKey`). As for the `address`  parameter, it refers to the derivation path used to derive `Addresses` from the `PubKey`. Any number of `Addresses` can be derived from the same `PubKey`, giving the possibility to developers to make use of single-use addresses for accounts. Finally, note that the `CreateAccount` method derives keys and addresses using `secp256k1` as implemented in the [Tendermint library](https://github.com/tendermint/tendermint/blob/master/crypto/secp256k1). As a result, it only works for creating account keys and addresses, not consensus keys. See [`Addresses`](#addresses) for more.

The current implementation of `dbKeybase` is basic and does not offer on-demand locking. If an instance of `dbKeybase` is created, the underlying `db` is locked meaning no other process can access it besides the one in which it was instantiated. This is the reason why the default SDK client uses another implementation of the `Keybase` interface called [`lazyKeybase`](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/lazy_keybase.go). `lazyKeybase` is simple wrapper around `dbKeybase` which locks the database only when operations are to be performed and unlocks it immediately after:

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

The `lazyKeybase` is typically used from 

### Addresses



### Signatures

## Fees and Gas

### AnteHandler

### Gas

### Gas Meter

### Block Gas Meter