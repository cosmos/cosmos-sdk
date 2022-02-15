<!--
order: 4
-->

# Accounts

This document describes the in-built account and public key system of the Cosmos SDK. {synopsis}

### Pre-requisite Readings

- [Anatomy of an SDK Application](./app-anatomy.md) {prereq}

## Account Definition

In the Cosmos SDK, an _account_ designates a pair of _public key_ `PubKey` and _private key_ `PrivKey`. The `PubKey` can be derived to generate various `Addresses`, which are used to identify users (among other parties) in the application. `Addresses` are also associated with [`message`s](../building-modules/messages-and-queries.md#messages) to identify the sender of the `message`. The `PrivKey` is used to generate [digital signatures](#signatures) to prove that an `Address` associated with the `PrivKey` approved of a given `message`.

For HD key derivation the Cosmos SDK uses a standard called [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki). The BIP32 allows users to create an HD wallet (as specified in [BIP44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki)) - a set of accounts derived from an initial secret seed. A seed is usually created from a 12- or 24-word mnemonic. A single seed can derive any number of `PrivKey`s using a one-way cryptographic function. Then, a `PubKey` can be derived from the `PrivKey`. Naturally, the mnemonic is the most sensitive information, as private keys can always be re-generated if the mnemonic is preserved.

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

In the Cosmos SDK, keys are stored and managed by using an object called a [`Keyring`](#keyring).

## Keys, accounts, addresses, and signatures

The principal way of authenticating a user is done using [digital signatures](https://en.wikipedia.org/wiki/Digital_signature). Users sign transactions using their own private key. Signature verification is done with the associated public key. For on-chain signature verification purposes, we store the public key in an `Account` object (alongside other data required for a proper transaction validation).

In the node, all data is stored using Protocol Buffers serialization.

The Cosmos SDK supports the following digital key schemes for creating digital signatures:

- `secp256k1`, as implemented in the [SDK's `crypto/keys/secp256k1` package](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keys/secp256k1/secp256k1.go).
- `secp256r1`, as implemented in the [SDK's `crypto/keys/secp256r1` package](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/secp256r1/pubkey.go),
- `tm-ed25519`, as implemented in the [SDK `crypto/keys/ed25519` package](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keys/ed25519/ed25519.go). This scheme is supported only for the consensus validation.

|              | Address length | Public key length | Used for transaction | Used for consensus |
|              | in bytes       |          in bytes | authentication       | (tendermint)       |
|--------------+----------------+-------------------+----------------------+--------------------|
| `secp256k1`  | 20             |                33 | yes                  | no                 |
| `secp256r1`  | 32             |                33 | yes                  | no                 |
| `tm-ed25519` | -- not used -- |                32 | no                   | yes                |

## Addresses

`Addresses` and `PubKey`s are both public information that identifies actors in the application. `Account` is used to store authentication information. The basic account implementation is provided by a `BaseAccount` object.

Each account is identified using `Address` which is a sequence of bytes derived from a public key. In SDK, we define 3 types of addresses that specify a context where an account is used:

- `AccAddress` identifies users (the sender of a `message`).
- `ValAddress` identifies validator operators.
- `ConsAddress` identifies validator nodes that are participating in consensus. Validator nodes are derived using the **`ed25519`** curve.

These types implement the `Address` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/address.go#L71-L90

Address construction algorithm is defined in [ADR-28](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md).
Here is the standard way to obtain an account address from a `pub` public key:

```go
sdk.AccAddress(pub.Address().Bytes())
```

Of note, the `Marshal()` and `Bytes()` method both return the same raw `[]byte` form of the address. `Marshal()` is required for Protobuf compatibility.

For user interaction, addresses are formatted using [Bech32](https://en.bitcoin.it/wiki/Bech32) and implemented by the `String` method. The Bech32 method is the only supported format to use when interacting with a blockchain. The Bech32 human-readable part (Bech32 prefix) is used to denote an address type. Example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/address.go#L230-L244

|                    | Address Bech32 Prefix |
| ------------------ | --------------------- |
| Accounts           | cosmos                |
| Validator Operator | cosmosvaloper         |
| Consensus Nodes    | cosmosvalcons         |

### Public Keys

Public keys in Cosmos SDK are defined by `cryptotypes.PubKey` interface. Since public keys are saved in a store, `cryptotypes.PubKey` extends the `proto.Message` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/types/types.go#L8-L17

A compressed format is used for `secp256k1` and `secp256r1` serialization.

- The first byte is a `0x02` byte if the `y`-coordinate is the lexicographically largest of the two associated with the `x`-coordinate.
- Otherwise the first byte is a `0x03`.

This prefix is followed by the `x`-coordinate.

Public Keys are not used to reference accounts (or users) and in general are not used when composing transaction messages (with few exceptions: `MsgCreateValidator`, `Validator` and `Multisig` messages).
For user interactions, `PubKey` is formatted using Protobufs JSON ([ProtoMarshalJSON](https://github.com/cosmos/cosmos-sdk/blob/release/v0.42.x/codec/json.go#L12) function). Example:

+++ https://github.com/cosmos/cosmos-sdk/blob/7568b66/crypto/keyring/output.go#L23-L39

## Keyring

A `Keyring` is an object that stores and manages accounts. In the Cosmos SDK, a `Keyring` implementation follows the `Keyring` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keyring/keyring.go#L51-L89

The default implementation of `Keyring` comes from the third-party [`99designs/keyring`](https://github.com/99designs/keyring) library.

A few notes on the `Keyring` methods:

- `Sign(uid string, payload []byte) ([]byte, sdkcrypto.PubKey, error)` strictly deals with the signature of the `payload` bytes. You must prepare and encode the transaction into a canonical `[]byte` form. Because protobuf is not deterministic, it has been decided in [ADR-020](../architecture/adr-020-protobuf-transaction-encoding.md) that the canonical `payload` to sign is the `SignDoc` struct, deterministically encoded using [ADR-027](adr-027-deterministic-protobuf-serialization.md). Note that signature verification is not implemented in the SDK by default, it is deferred to the [`anteHandler`](../core/baseapp.md#antehandler).
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/proto/cosmos/tx/v1beta1/tx.proto#L47-L64

- `NewAccount(uid, mnemonic, bip39Passwd, hdPath string, algo SignatureAlgo) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk. The `PrivKey` is **never stored unencrypted**, instead it is [encrypted with a passphrase](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/armor.go)  before being persisted. In the context of this method, the key type and sequence number refer to the segment of the BIP44 derivation path (for example, `0`, `1`, `2`, ...) that is used to derive a private and a public key from the mnemonic. Using the same mnemonic and derivation path, the same `PrivKey`, `PubKey` and `Address` is generated. The following keys are supported by the keyring:

- `secp256k1`
- `ed25519`

- `ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)` exports a private key in ASCII-armored encrypted format using the given passphrase. You can then either import the private key again into the keyring using the `ImportPrivKey(uid, armor, passphrase string)` function or decrypt it into a raw private key using the `UnarmorDecryptPrivKey(armorStr string, passphrase string)` function.

## Next {hide}

Learn about [gas and fees](./gas-fees.md) {hide}
