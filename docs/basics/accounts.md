<!--
order: 4
-->

# Accounts

This document describes the in-built accounts system of the Cosmos SDK. {synopsis}

### Pre-requisite Readings

- [Anatomy of an SDK Application](./app-anatomy.md) {prereq}

## Account Definition

In the Cosmos SDK, an _account_ designates a pair of _public key_ `PubKey` and _private key_ `PrivKey`. The `PubKey` can be derived to generate various `Addresses`, which are used to identify users (among other parties) in the application. `Addresses` are also associated with [`message`s](../building-modules/messages-and-queries.md#messages) to identify the sender of the `message`. The `PrivKey` is used to generate [digital signatures](#signatures) to prove that an `Address` associated with the `PrivKey` approved of a given `message`.

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

In the Cosmos SDK, accounts are stored and managed via an object called a [`Keyring`](#keyring).

## Keyring

A `Keyring` is an object that stores and manages accounts. In the Cosmos SDK, a `Keyring` implementation follows the `Keyring` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/keyring/keyring.go#L50-L88

The default implementation of `Keyring` comes from the third-party [`99designs/keyring`](https://github.com/99designs/keyring) library.

A few notes on the `Keyring` methods:

- `Sign(uid string, payload []byte) ([]byte, tmcrypto.PubKey, error)` strictly deals with the signature of the `payload` bytes. Some preliminary work should be done beforehand to prepare and encode the transaction into a canonical `[]byte` form. Protobuf being not deterministic, it has been decided in [ADR-020](../architecture/adr-020-protobuf-transaction-encoding.md) that the canonical `payload` to sign is the `SignDoc` struct, deterministically encoded using [ADR-027](adr-027-deterministic-protobuf-serialization.md). Note that signature verification is not implemented in the SDK by default, it is deferred to the [`anteHandler`](../core/baseapp.md#antehandler).
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/proto/cosmos/tx/v1beta1/tx.proto#L47-L64

- `NewAccount(uid, mnemonic, bip39Passwd, hdPath string, algo SignatureAlgo) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk (note that the `PrivKey` is [encrypted with a passphrase before being persisted](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/armor.go), it is **never stored unencrypted**). In the context of this method, the `account` and `address` parameters refer to the segment of the BIP44 derivation path (e.g. `0`, `1`, `2`, ...) used to derive the `PrivKey` and `PubKey` from the mnemonic (note that given the same mnemonic and `account`, the same `PrivKey` will be generated, and given the same `account` and `address`, the same `PubKey` and `Address` will be generated). Finally, note that the `NewAccount` method derives keys and addresses using the algorithm specified in the last argument `algo`. Currently, the SDK supports two public key algorithms:

  - `secp256k1`, as implemented in the [SDK's `crypto/keys/secp256k1` package](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/keys/secp256k1/secp256k1.go),
  - `ed25519`, as implemented in the [SDK's `crypto/keys/ed25519` package](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/keys/ed25519/ed25519.go).

- `ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)` exports a private key in ASCII-armored encrypted format, using the given passphrase. You can then either import it again into the keyring using the `ImportPrivKey(uid, armor, passphrase string)` function, or decrypt it into a raw private key using the `UnarmorDecryptPrivKey(armorStr string, passphrase string)` function.

Also see the [`Addresses`](#addresses) section for more information.

## Addresses and PubKeys

`Addresses` and `PubKey`s are both public information that identify actors in the application. There are 3 main types of `Addresses`/`PubKeys` available by default in the Cosmos SDK:

- Addresses and Keys for **accounts**, which identify users (e.g. the sender of a `message`). They are derived using the **`secp256k1`** curve.
- Addresses and Keys for **validator operators**, which identify the operators of validators. They are derived using the **`secp256k1`** curve.
- Addresses and Keys for **consensus nodes**, which identify the validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

|                    | Address bech32 Prefix | Pubkey bech32 Prefix | Curve       | Address byte length | Pubkey byte length |
| ------------------ | --------------------- | -------------------- | ----------- | ------------------- | ------------------ |
| Accounts           | cosmos                | cosmospub            | `secp256k1` | `20`                | `33`               |
| Validator Operator | cosmosvaloper         | cosmosvaloperpub     | `secp256k1` | `20`                | `33`               |
| Consensus Nodes    | cosmosvalcons         | cosmosvalconspub     | `ed25519`   | `20`                | `32`               |

### PubKeys

`PubKey`s used in the Cosmos SDK are Protobuf messages and have the following methods:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/types/types.go#L8-L17

- For `secp256k1` keys, the actual implementation can be found [here](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/keys/secp256k1/secp256k1.go).
- For `ed25519` keys, it can be found [here](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/keys/ed25519/ed25519.go).

In both case, the actual key (as raw bytes) is the compressed form of the pubkey. The first byte is a `0x02` byte if the `y`-coordinate is the lexicographically largest of the two associated with the `x`-coordinate. Otherwise the first byte is a `0x03`. This prefix is followed with the `x`-coordinate.

Note that in the Cosmos SDK, `Pubkeys` are not manipulated in their raw bytes form. Instead, they are encoded to string using [`Amino`](../core/encoding.md#amino) and [`bech32`](https://en.bitcoin.it/wiki/Bech32). In the SDK, it is done by first calling the `Bytes()` method on the raw `Pubkey` (which applies amino encoding), and then the `ConvertAndEncode` method of `bech32`.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/address.go#L579-L729

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

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/address.go#L76-L85

Of note, the `Marshal()` and `Bytes()` method both return the same raw `[]byte` form of the address, the former being needed for Protobuf compatibility. Also, the `String()` method is used to return the `bech32` encoded form of the address, which should be the only address format with which end-user interract. Here is an example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/address.go#L235-L249

## Next {hide}

Learn about [gas and fees](./gas-fees.md) {hide}
