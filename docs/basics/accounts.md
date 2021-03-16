<!--
order: 4
-->

# Accounts

This document describes the in-built accounts system of the Cosmos SDK. {synopsis}

### Pre-requisite Readings

- [Anatomy of an SDK Application](./app-anatomy.md) {prereq}

## Account Definition

In the Cosmos SDK, an _account_ designates a pair of _public key_ `PubKey` and _private key_ `PrivKey`. The `PubKey` can be derived to generate various `Addresses`, which are used to identify users (among other parties) in the application. `Addresses` are also associated with [`message`s](../building-modules/messages-and-queries.md#messages) to identify the sender of the `message`. The `PrivKey` is used to generate [digital signatures](#signatures) to prove that an `Address` associated with the `PrivKey` approved of a given `message`.

For HD key derivation the Cosmos SDK uses a standard called [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki). This allows users to create an HD wallet - a set of accounts derived from an initial secret seed. A seed is usually created from a 12- or 24-word mnemonic. This single seed allows one to derive any number of `PrivKey`s using a one-way cryptographic function. Then, a `PubKey` can be derived from the `PrivKey`. Naturally, the mnemonic is the most sensitive information, as private keys can always be re-generated if the mnemonic is preserved.

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

In the Cosmos SDK, keys are stored and managed via an object called a [`Keyring`](#keyring).


## Keys, accounts, addresses and signatures

The principal way of authenticating a user is done through [digital signatures](https://en.wikipedia.org/wiki/Digital_signature). Users signs transactions using their own private key. Signature verification is done with the associated public key. For on-chain signature verification purposes, we store the public key in an `Account` object (alongside other data required for a proper transaction validation).

Currently, the Cosmos SDK supports the following digital key schemes for creating digital signatures:

- `secp256k1`, as implemented in the [SDK's `crypto/keys/secp256k1` package](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keys/secp256k1/secp256k1.go).
- `secp256r1`, as implemented in the [SDK's `crypto/keys/secp256r1` package](https://github.com/cosmos/cosmos-sdk/blob/master/crypto/keys/secp256r1/pubkey.go),
- `tm-ed25519`, as implemented in the [SDK's `crypto/keys/ed25519` package](https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keys/ed25519/ed25519.go). This scheme is only supported for the consensus validation, not user transaction.


|              | Address length | Public key length | Used for transaction | Used for consensus |
|              | in bytes       |          in bytes | authentication       | (tendermint)       |
|--------------+----------------+-------------------+----------------------+--------------------|
| `secp256k1`  | 20             |                33 | yes                  | no                 |
| `secp256k1`  | 32             |                33 | yes                  | no                 |
| `tm-ed25519` | -- not used -- |                32 | no                   | yes                |



## Addresses

`Addresses` and `PubKey`s are both public information that identify actors in the application. As we noted above, `Account` is used to store authentication information. The basic account implementation is provided by a `BaseAccount` object.

Each account is identified using `Address` - a sequence of bytes derived from a public key. Moreover, in SDK, we define 3 types of addresses, which specify a context where an account is used:

- `AccAddress` is used to identify users (e.g. the sender of a `message`).
- `ValAddress` is used to identify validator operators.
- `ConsAddress` is used to identify validator nodes participating in consensus. They are derived using the **`ed25519`** curve.

These types implement the `Address` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/address.go#L104-L113

Address construction algorithm is defined in [ADR-28](https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-028-public-key-addresses.md).
Here is the standard way to obtain an account address from a `pub` public key:

```go
sdk.AccAddress(pub.Address().Bytes())
```

Of note, the `Marshal()` and `Bytes()` method both return the same raw `[]byte` form of the address, the former being needed for Protobuf compatibility.

Addresses and public keys are formatted using [bech32](https://en.bitcoin.it/wiki/Bech32) and implemented by the `String` method. This is the only format a user should use when interacting with a blockchain. Bech32 human readable part (bech32 prefix) is used to denote an address type. Example:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/types/address.go#L235-L249


|                    | Address bech32 Prefix | Pubkey bech32 Prefix |
| ------------------ | --------------------- | -------------------- |
| Accounts           | cosmos                | cosmospub            |
| Validator Operator | cosmosvaloper         | cosmosvaloperpub     |
| Consensus Nodes    | cosmosvalcons         | cosmosvalconspub     |


### Public Keys

Public keys in Cosmos SDK are defined by `crypto/PubKey` interface. Since they are saved in a store, it extends the `proto.Message` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/master/crypto/types/types.go#L8-L17

For `secp256k1` and `secp256r1` serialization, a compressed format is used. The first byte is a `0x02` byte if the `y`-coordinate is the lexicographically largest of the two associated with the `x`-coordinate. Otherwise the first byte is a `0x03`. This prefix is followed with the `x`-coordinate.

Similarly to `Address`, bech32 is used to format `PubKey` and for all communication with a blockchain:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/types/address.go#L579-L729


## Keyring

A `Keyring` is an object that stores and manages accounts. In the Cosmos SDK, a `Keyring` implementation follows the `Keyring` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/crypto/keyring/keyring.go#L51-L92

The default implementation of `Keyring` comes from the third-party [`99designs/keyring`](https://github.com/99designs/keyring) library.

A few notes on the `Keyring` methods:

- `Sign(uid string, payload []byte) ([]byte, sdkcrypto.PubKey, error)` strictly deals with the signature of the `payload` bytes. Some preliminary work should be done beforehand to prepare and encode the transaction into a canonical `[]byte` form. Protobuf being not deterministic, it has been decided in [ADR-020](../architecture/adr-020-protobuf-transaction-encoding.md) that the canonical `payload` to sign is the `SignDoc` struct, deterministically encoded using [ADR-027](adr-027-deterministic-protobuf-serialization.md). Note that signature verification is not implemented in the SDK by default, it is deferred to the [`anteHandler`](../core/baseapp.md#antehandler).
  +++ https://github.com/cosmos/cosmos-sdk/blob/v0.42.1/proto/cosmos/tx/v1beta1/tx.proto#L47-L64

- `NewAccount(uid, mnemonic, bip39Passwd, hdPath string, algo SignatureAlgo) (Info, error)` creates a new account based on the [`bip44 path`](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki) and persists it on disk (note that the `PrivKey` is [encrypted with a passphrase before being persisted](https://github.com/cosmos/cosmos-sdk/blob/v0.40.0-rc3/crypto/armor.go), it is **never stored unencrypted**). In the context of this method, the `account` and `address` parameters refer to the segment of the BIP44 derivation path (e.g. `0`, `1`, `2`, ...) used to derive the `PrivKey` and `PubKey` from the mnemonic (note that given the same mnemonic and `account`, the same `PrivKey` will be generated, and given the same `account` and `address`, the same `PubKey` and `Address` will be generated). Finally, note that the `NewAccount` method derives keys and addresses using the algorithm specified in the last argument `--algo`. The following keys are supported by the keyring:

- `secp256k1`
- `ed25519`

- `ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error)` exports a private key in ASCII-armored encrypted format, using the given passphrase. You can then either import it again into the keyring using the `ImportPrivKey(uid, armor, passphrase string)` function, or decrypt it into a raw private key using the `UnarmorDecryptPrivKey(armorStr string, passphrase string)` function.


## Next {hide}

Learn about [gas and fees](./gas-fees.md) {hide}
