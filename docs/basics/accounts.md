<!--
order: 3
synopsis: This document describes the in-built accounts system of the Cosmos SDK.
-->

# Accounts 

## Pre-requisite Readings {hide}

- [Anatomy of an SDK Application](./app-anatomy.md) {prereq}

## Account Definition

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

## Keybase

A `Keybase` is an object that stores and manages accounts. In the Cosmos SDK, a `Keybase` implementation follows the `Keybase` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/types.go#L13-L86

The default implementation of `Keybase` of the Cosmos SDK is `dbKeybase`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/keybase.go

A few notes on the `Keybase` methods as implemented in `dbKeybase`:

- `Sign(name, passphrase string, msg []byte) ([]byte, crypto.PubKey, error)` strictly deals with the signature of the `message` bytes. Some preliminary work should be done beforehand to prepare and encode the `message`  into a canonical `[]byte` form. See an example of `message` preparation from the `auth` module. Note that signature verification is not implemented in the SDK by default. It is deferred to the [`anteHandler`](#antehandler).
	+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/auth/types/txbuilder.go#L176-L209
- `CreateMnemonic(name string, language Language, passwd string, algo SigningAlgo) (info Info, seed string, err error)` creates a new mnemonic and prints it in the logs, but it **does not persist it on disk**. 
- `CreateAccount(name, mnemonic, bip39Passwd, encryptPasswd string, account uint32, index uint32) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk (note that the `PrivKey` is [encrypted with a passphrase before being persisted](https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/mintkey/mintkey.go), it is **never stored unencrypted**). In the context of this method, the `account` and `address` parameters refer to the segment of the BIP44 derivation path (e.g. `0`, `1`, `2`, ...) used to derive the `PrivKey` and `PubKey` from the mnemonic (note that given the same mnemonic and `account`, the same `PrivKey` will be generated, and given the same `account` and `address`, the same `PubKey` and `Address` will be generated). Finally, note that the `CreateAccount` method derives keys and addresses using `secp256k1` as implemented in the [Tendermint library](https://github.com/tendermint/tendermint/tree/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/secp256k1). As a result, it only works for creating account keys and addresses, not consensus keys. See [`Addresses`](#addresses) for more.

The current implementation of `dbKeybase` is basic and does not offer on-demand locking. If an instance of `dbKeybase` is created, the underlying `db` is locked meaning no other process can access it besides the one in which it was instantiated. This is the reason why the default SDK client uses another implementation of the `Keybase` interface called `lazyKeybase`:
	
	
+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/crypto/keys/lazy_keybase.go

`lazyKeybase` is simple wrapper around `dbKeybase` which locks the database only when operations are to be performed and unlocks it immediately after. With the `lazyKeybase`, it is possible for the [command-line interface](../interfaces/cli.md) to create a new account while the [rest server](../interfaces/rest.md) is running. It is also possible to pipe multiple CLI commands. 

## Addresses and PubKeys

`Addresses` and `PubKey`s are both public information that identify actors in the application. There are 3 main types of `Addresses`/`PubKeys` available by default in the Cosmos SDK:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`secp256k1`** curve. 
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`secp256k1`** curve. 
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve       | Address byte length | Pubkey byte length |
|--------------------|-----------------------|----------------------|-------------|---------------------|--------------------|
| Accounts           | cosmos                | cosmospub            | `secp256k1` | `20`                | `33`               |
| Validator Operator | cosmosvaloper         | cosmosvaloperpub     | `secp256k1` | `20`                | `33`               |
| Consensus Nodes    | cosmosvalcons         | cosmosvalconspub     | `ed25519`   | `20`                | `32`               | 

### PubKeys

`PubKey`s used in the Cosmos SDK follow the `Pubkey` interface defined in tendermint's `crypto` package:

+++ https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/crypto.go#L22-L27

For `secp256k1` keys, the actual implementation can be found [here](https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/secp256k1/secp256k1.go#L140). For `ed25519` keys, it can be found [here](https://github.com/tendermint/tendermint/blob/bc572217c07b90ad9cee851f193aaa8e9557cbc7/crypto/ed25519/ed25519.go#L135). 

Note that in the Cosmos SDK, `Pubkeys` are not manipulated in their raw form. Instead, they are double encoded using [`Amino`](../core/encoding.md#amino) and [`bech32`](https://en.bitcoin.it/wiki/Bech32). In the SDK is done by first calling the `Bytes()` method on the raw `Pubkey` (which applies amino encoding), and then the `ConvertAndEncode` method of `bech32`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L579-L729

### Addresses

The Cosmos SDK comes by default with 3 types of addresses:

- `AccAddress` for accounts.
- `ValAddress` for validator operators. 
- `ConsAddress` for validator nodes. 

Each of these address types are an alias for an hex-encoded `[]byte` array of length 20. Here is the standard way to obtain an address `aa` from a `Pubkey pub`:

```go
aa := sdk.AccAddress(pub.Address().Bytes())
```

These addresses implement the `Address` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L71-L80

Of note, the `Marhsal()` and `Bytes()` method both return the same raw `[]byte` form of the address, the former being needed for Protobuff compatibility. Also, the `String()` method is used to return the `bech32` encoded form of the address, which should be the only address format with which end-user interract. Next is an example:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/address.go#L229-L243

## Next {hide}

Learn about [gas and fees](./gas-fees.md) {hide}