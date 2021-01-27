# ADR 028: Public Key Addresses

## Changelog

- 2020/08/18: Initial version
- 2021/01/15: Analysis and algorithm update

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
we concatenate a key type with a public key, hash it and take the first 20 bytes of that hash, summarized as `sha256(keyTypePrefix || keybytes)[:20]`.


### Review and Discussions

In [\#5694](https://github.com/cosmos/cosmos-sdk/issues/5694) we discussed various solutions.
We agreed that 20 bytes it's not future proof, and extending the address length is the only way to allow addresses of different types, various signature types, etc.
This disqualifies the initial proposal.

In the issue we discussed various modifications:
+ Choice of the hash function.
+ Move the prefix out of the hash function: `keyTypePrefix + sha256(keybytes)[:20]` [post-hash-prefix-proposal].
+ Use double hashing: `sha256(keyTypePrefix + sha256(keybytes)[:20])`.
+ Increase to keybytes hash slice from 20 byte to 32 or 40 bytes. We concluded that 32 bytes, produced by a good hash functions is future secure.

### Requirements

+ Support currently used tools - we don't want to break an ecosystem, or add a long adaptation period. Ref: https://github.com/cosmos/cosmos-sdk/issues/8041
+ Try to keep the address length small - addresses are widely used in state, both as part of a key and object value.


### Scope

This ADR only defines a process for the generation of address bytes. For end-user interactions with addresses (through the API, or CLI, etc.), we still use bech32 to format these addresses as strings. This ADR doesn't change that.
Using bech32 for string encoding gives us support for checsum error codes and handling of user typos.


## Decision

We define the following account types, for which we define the address function:

1. simple accounts: represented by a regular public key (ie: secp256k1, sr25519)
2. naive multisig: accounts composed by other addressable objects (ie: naive multisig)
3. composed accounts with a native address key (ie: bls, group module accounts)
4. module accounts: basically any accounts which cannot sign transactions and which are managed internally by modules


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

As in other parts of the Cosmos SDK, we will use `sha256`.

### Simple Account Address

We start with defining a hash base algorithm for generating addresses. Notably, it's used for accounts represented by a single key pair. For each public key schema we have to have an associated `typ` string, which we are discussing in a section below. `hash` is a cryptographic hash function defined in the previous section.

```go
const A_LEN = 32

func Hash(typ string, key []byte) []byte {
    return hash(hash(typ) + key)[:A_LEN]
}
```

The `+` is a bytes concatenation, which doesn't use any separator.

This algorithm is an outcome from a consulting session with a cryptographer.
Motivation: this algorithm keeps the address relatively small (length of the `typ` doesn't impact on the length of the final address)
and it's more secure than [post-hash-prefix-proposal] (which uses first 20 bytes of a pubkey hash, that significantly reduce the address space).
Moreover the cryptographer motivated the choice to add `typ` in the hash to protect against switch table attack.

We use `address.Hash` function for generating address for all accounts represented by a single key:
* simple public keys: `address.Hash(keyType, pubkey)`
+ aggregated keys (eg: BLS): `address.Hash(keyType, aggregatedPubKey)`
+ module accounts: `addoress.Hash("module", moduleID)`
  For module subaccount, or module specific content we append a specific key to the moduleKey:
  `address.Hash("module", moduleSubaccountKey)`  (note: `moduleSubaccountKey` already composes the `modleKey`).


### Composed Account Addresses

For simple composed accounts (like new naive multisig), we generalize the `address.Hash`. The address is constructed by recursively creating addresses for the sub accounts, sorting the addresses and composing it into a single address. It ensures that the ordering of keys doesn't impact the created address.

```go
// We don't need a PubKey interface - we need anything which is addressable.
type Addressable interface {
    Address() []byte
}

func Composed(typ string, subaccounts []Addressable) []byte {
    addresses = map(subaccounts, \a -> a.Address())
    addresses = sort(addresses)
    return address.Hash(typ, addresses[0] + ... + addresses[n])
}
```

The `typ` parameter should contain all significant attributes with deterministic serialization.

Implementation Tip: account implementations should cache address in their structure.


#### Multisig Addresses

For new multisig public keys, we define the `typ` parameter not based on any encoding scheme (amino or protobuf). This avoids issues with non-determinism in the encoding scheme.

Example:

```proto
package cosmos.crypto.multisig;

message PubKey {
  uint32 threshold = 1;
  repeated google.protobuf.Any pubkeys = 2;
}
```

```go
func (multisig PubKey) Address() {
  // first gather all nested pub keys
  var keys []address.Addressable  // cryptotypes.PubKey implements Addressable
  for _, _key := range multisig.Pubkeys {
    keys = append(keys, key.GetCachedValue().(cryptotypes.PubKey))
  }

  // form the type from the message name (cosmos.crypto.multisig.PubKey) and the threshold joined together
  prefix := fmt.Sprintf("%s/%d", proto.MessageName(multisig), multisig.Threshold)

  // use the Composed function defined above
  return address.Composed(prefix, keys)
}
```


### Account Types

The Account Types used in various account classes SHOULD be unique for each class.
Since both public keys and accounts are serialized in the state, we propose to use the protobuf message name string.

Example: all public key types have a unique protobuf message type similar to:

```proto
package cosmos.crypto.sr25519;

message PubKey {
  bytes key = 1;
}
```

All protobuf messages have unique fully qualified names, in this example `cosmos.crypto.sr25519.PubKey`.
These names are derived directly from .proto files in a standardized way and used
in other places such as the type URL in `Any`s. We can easily obtain the name using
`proto.MessageName(msg)`.



## Consequences

### Backwards Compatibility

This ADR is compatible to what was committed and directly supported in the SDK repository.

### Positive

- a simple algorithm for generating addresses for new public keys, complex accounts and module accounts
- the algorithm generalizes for _native composed keys_
- increase security and collision resistance of addresses
- the approach is extensible for future use-cases - one can use other address types, as long as they don't conflict with the address length specified here (20 or 32 bytes).
- support multiple types accounts.

### Negative

- addresses do not communicate key type, a prefixed approach would have done this
- addresses are 60% longer and will consume more storage space
- requires a refactor of KVStore store keys to handle variable length addresses

### Neutral

- protobuf message names are used as key type prefixes


## Further Discussions

Some accounts can have a fixed name or may be constructed in other way (eg: module accounts). We were discussing an idea of an account with a predefined name (eg: `me.regen`), which could be used by institutions.
Without going into details, this kind of addresses are compatible with the hash based addresses described here as long as they don't have the same length.
More specifically, any special account address, must not have length equal to 20 byte nor 32 bytes.


## References

* [Notes](https://hackmd.io/_NGWI4xZSbKzj1BkCqyZMw) from consulting meeting with [Alan Szepieniec](https://scholar.google.be/citations?user=4LyZn8oAAAAJ&hl=en).
