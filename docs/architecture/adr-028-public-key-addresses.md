# ADR 028: Public Key Addresses

## Changelog

- 2020/08/18: Initial version
- 2020/08/15: Analysis and algorithm update

## Status

LAST CALL 2021-01-22

## Abstract

This ADR defines an address format for all addressable SDK accounts. That includes: new public key algorithms, multisig public keys, and module
accounts.

## Context

Issue [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) identified that public key
address spaces are currently overlapping. We confirmed that it significantly decreases security of Cosmos SDK.


### Problem

An attacker can control an input for an address generation function. This leads to a birthday attack, which significantly decreases the security space.
To overcome this, we need to separate the inputs for different kind of account types:
a security break of one account type shouldn't impact the security of other account type.


### Initial proposals

One initial proposal was extending the address length and
adding prefixes for different types of addresses.

@ethanfrey explained an alternate approach originally used in https://github.com/iov-one/weave:

> I spent quite a bit of time thinking about this issue while building weave... The other cosmos Sdk.
> Basically I define a condition to be a type and format as human readable string with some binary data appended. This condition is hashed into an Address (again at 20 bytes). The use of this prefix makes it impossible to find a preimage for a given address with a different condition (eg ed25519 vs secp256k1).
> This is explained in depth here https://weave.readthedocs.io/en/latest/design/permissions.html
> And the code is here, look mainly at the top where we process conditions. https://github.com/iov-one/weave/blob/master/conditions.go

And explained how this approach should be sufficiently collision resistant:

> Yeah, AFAIK, 20 bytes should be collision resistance when the preimages are unique and not malleable. A space of 2^160 would expect some collision to be likely around 2^80 elements (birthday paradox). And if you want to find a collision for some existing element in the database, it is still 2^160. 2^80 only is if all these elements are written to state.
> The good example you brought up was eg. a public key bytes being a valid public key on two algorithms supported by the codec. Meaning if either was broken, you would break accounts even if they were secured with the safer variant. This is only as the issue when no differentiating type info is present in the preimage (before hashing into an address).
> I would like to hear an argument if the 20 bytes space is an actual issue for security, as I would be happy to increase my address sizes in weave. I just figured cosmos and ethereum and bitcoin all use 20 bytes, it should be good enough. And the arguments above which made me feel it was secure. But I have not done a deeper analysis.

This lead to the first proposal (which we proved to be not good enough):
we take the first 20 bytes of the `sha256` hash of the public key we concatenated with the key bytes, summarized as `sha256(keyTypePrefix || keybytes)[:20]`.


### Review and Discussions

In [\#5694](https://github.com/cosmos/cosmos-sdk/issues/5694) we discussed various solutions.
We agreed that 20 bytes it's not future proof, and extending the address length is the only way to allow addresses of different types, various signature types, etc.
This disqualifies the initial proposal.

In the issue we discussed various modifications:
+ Choice of the hash function.
+ Move the prefix out of the hash function: `keyTypePrefix || sha256(keybytes)[:20]` [post-hash-prefix-proposal].
+ Use double hashing: `sha256(keyTypePrefix || sha256(keybytes)[:20])`.
+ Increase to keybytes hash slice from 20 byte to 32 or 40 bytes. We concluded that 32 bytes, produced by a good hash functions is future secure.

### Requirements

+ Support currently used tools - we don't want to break an ecosystem, or add a long adaptation period.
+ Try to keep the address length small - addresses are widely used in state, both as part of a key and object value.


### Scope

This ADR defines only an address bytes. The for the API level we already use bech32 and this ADR doesn't change that.
Bech32 support checsum error codes and handles user typos.


## Decision

### Legacy Public Key Addresses Don't Change

Currently (Jan 2021), the only officially supported SDK user accounts are `secp256k1` basic accounts and legacy amino multisig.
They are used in existing Cosmos SDK zones. They use the following address formats:

- secp256k1: `ripemd160(sha256(pk_bytes))[:20]`
- legacy amino multisig: `sha256(aminoCdc.Marshal(pk))[:20]`

We don't want to change existing addresses. So the addresses for these two key types will remain the same.

The current multisig public keys use amino serialization to generate the address. We will retain
those public keys and their address formatting, and call them "legacy amino" multisig public keys
in protobuf. We will also create multisig public keys without amino addresses to be described below.

### Hash Function Choice

We propose to use [blake2b](https://www.blake2.net/) as a hash function choice:
+ The main arguments are speed and separating from `sha256` which is widely used
  by miners and could potentially be used to find collisions.
+ The function was in the final round of the 2012 NIST hash function competition.
+ It's well studied with security covered in many academic papers.
+ Faster than `sha2` on non ASICs chipsets.
+ It's getting more traction in other blockchains (Pokadot, Sia, Zcash, ...). Related [zcash discussion](https://github.com/zcash/zcash/issues/706#issuecomment-187807410).
+ It's already widely supported by all major programming languages.
+ Cryptography consulting reviled no argument against `blake2b`


### Base Address Algorithm

We start with defining a base algorithm for generating addresses. Notably, it's used for Base Accounts (accounts represented by a single key-pair) addresses. For each Public Key schema we need to have an associated `typ` string, which we will discuss in a section below. `hash` is a cryptographic hash function defined in the previous section.

```go
const A_LEN = 32

func BaseAddress(typ string, pubkey []byte) []byte {
    return hash(hash(typ) + pubkey)[:A_LEN]
}
```

The `+` is a bytes concatenation, which doesn't use any separator.

This algorithm is an outcome after a consulting session with a cryptographer.
Motivation: this algorithm keeps the address relatively small (length of the `typ` doesn't impact on the length of the final address)
and it's more secure than [post-hash-prefix-proposal] (with reducing the pubkey hash to 20 bytes, we essentially significantly reduce the address space).
Moreover the cryptographer motivated the choice to add `typ` in the hash to protect against switch table attack.


### Composed Account Address Algorithm

We will generalize `BaseAddress` algorithm to define an address for an account which is represented by a set of sub accounts (example: group module accounts, multisig acconts...).
The address is constructed by recursively creating addresses for the sub accounts, sorting the addresses and composing it into a single address:

```go

type Acc interface {
    Typ() string
    SubAccounts() []Acc
    ...
}

type BaseAccount interface {
    Acc
    PubKey() crypto.PubKey
}

func Address(acc Acc) []byte {
    typ := acc.Typ()
    if acc is BaseAccount {
        return BaseAddress(typ, acc.PubKey())
    }
    subacconts := acc.SubAcconts()
    addresses := map(subaccount, Address)
    addresses = sort(addresses)
    n := len(addresses) - 1

    return BaseAddress(typ, addresses[0] + ... + addresses[n])
}
```

Implementation Tip: `Acc` implementations should cache address in their attributes.


### Native Composed Accounts

For accounts with a well specified public key composed of other public keys (various algorithms for aggregated signatures),
we will use a public key defined by the composition algorithm and we will call it _composed pubkey_. Example: BLS multisig.
The address algorithm for such accounts is same as the `BaseAccount`. In the example below, `na` is an object representing a native composed account.

```
na.address = BaseAddress(na.typ, na.ComposedPubKey)
```

### Account Types

The Account Types used in various account classes SHOULD be unique for each class.
Since both public keys and accounts are serialized in the state, we propose to use the protobuf message name string (`proto.MessageName(msg)`).

Example: all public key types have a unique protobuf message type similar to:

```proto
package cosmos.crypto.sr25519;

message PubKey {
  bytes key = 1;
}
```

All protobuf messages have unique fully qualified names, in this example `cosmos.crypto.sr25519.PubKey`.
These names are derived directly from .proto files in a standardized way and used
in other places such as the type URL in `Any`s. Since there is an easy and obvious
way to get this name for every protobuf type, we can use this message name as the
key type `prefix` when creating addresses. For all basic public keys, `contents`
should just be the raw unencoded public key bytes.



## Consequences

### Positive

- a simple algorithm for generating addresses for new public keys, complex accounts and module accounts
- the algorithm generalizes for _native composed keys_
- increase security and collision resistance of addresses
- the approach is extensible for future use-cases - one can use shorter addresses (>20 and < 32) for other use-cases.

### Negative

- addresses do not communicate key type, a prefixed approach would have done this
- addresses are 60% longer and will consume more storage space

### Neutral
- protobuf message names are used as key type prefixes


## References

* [Notes](https://hackmd.io/_NGWI4xZSbKzj1BkCqyZMw) from consulting meeting with [Alan Szepieniec](https://scholar.google.be/citations?user=4LyZn8oAAAAJ&hl=en).
* Blake2b security analysis: [1](https://eprint.iacr.org/2013/467), [2](https://eprint.iacr.org/2014/1012), [3](https://eprint.iacr.org/2015/515).
