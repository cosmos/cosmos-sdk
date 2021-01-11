# ADR 028: Public Key Addresses

## Changelog

- 2020/08/18: Initial version

## Status

Proposed

## Abstract

This ADR defines a canonical 20-byte address format for new public key algorithms, multisig public keys, and module
accounts using string prefixes.

## Context

Issue [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) identified that public key
address spaces are currently overlapping. One initial proposal was extending the address length and
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

In discussions in [\#5694](https://github.com/cosmos/cosmos-sdk/issues/5694), we agreed to go with an
approach similar to this where essentially we take the first 20 bytes of the `sha256` hash of
the key type concatenated with the key bytes, summarized as `Sha256(KeyTypePrefix || Keybytes)[:20]`.

## Decision

### Legacy Public Key Addresses Don't Change

`secp256k1` and multisig public keys are currently in use in existing Cosmos SDK zones. They use the following
address formats:

- secp256k1: `ripemd160(sha256(pk_bytes))[:20]`
- legacy amino multisig: `sha256(aminoCdc.Marshal(pk))[:20]`

We don't want to change existing addresses. So the addresses for these two key types will remain the same.

The current multisig public keys use amino serialization to generate the address. We will retain
those public keys and their address formatting, and call them "legacy amino" multisig public keys
in protobuf. We will also create multisig public keys without amino addresses to be described below.


### Canonical Address Format

We have three types of accounts we would like to create addresses for in the future:
- regular public key addresses for new signature algorithms (ex. `sr25519`).
- public key addresses for multisig public keys that don't use amino encoding
- module accounts: basically any accounts which cannot sign transactions and
which are managed internally by modules

To address all of these use cases we propose the following basic `AddressHash` function,
based on the discussions in [\#5694](https://github.com/cosmos/cosmos-sdk/issues/5694):

```go
func AddressHash(prefix string, contents []byte) []byte {
	preImage := []byte(prefix)
	if len(contents) != 0 {
		preImage = append(preImage, 0)
		preImage = append(preImage, contents...)
	}
	return sha256.Sum256(preImage)[:20]
}
```

`AddressHash` always take a string `prefix` as a starting point which should represent the
type of public key (ex. `sr25519`) or module account being used (ex. `staking` or `group`).
For public keys, the `contents` parameter is used to specify the binary contents of the public
key. For module accounts, `contents` can be left empty (for modules which don't manage "sub-accounts"),
or can be some module-specific content to specify different pools (ex. `bonded` or `not-bonded` for `staking`)
or managed accounts (ex. different accounts managed by the `group` module).

In the `preImage`, the byte value `0` is used as the separator between `prefix` and `contents`. This is a logical
choice given that `0` is an invalid value for a string character and is commonly used as a null terminator.

### Canonical Public Key Address Prefixes

All public key types will have a unique protobuf message type such as:

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

Thus the canonical address for new public key types would be `AddressHash(proto.MessageName(pk), pk.Bytes)`.

### Multisig Addresses

For new multisig public keys, we define a custom address format not based on any encoding scheme
(amino or protobuf). This avoids issues with non-determinism in the encoding scheme. It also
ensures that multisig public keys which differ simply in the ordering of keys have the same
address by sorting child public keys first.

First we define a proto message for multisig public keys:
```proto
package cosmos.crypto.multisig;

message PubKey {
  uint32 threshold = 1;
  repeated google.protobuf.Any public_keys = 2;
}
```

We define the following `Address()` function for this public key:

```
func (multisig PubKey) Address() {
  // first gather all the addresses of each nested public key
  var addresses [][]byte
  for key := range multisig.Keys {
    addresses = append(joinedAddresses, key.Address())
  }

  // then sort them in ascending order
  addresses = Sort(addresses)

  // then concatenate them together
  var joinedAddresses []byte
  for addr := range addresses {
    joinedAddresses := append(joinedAddresses, addr...)
  }

  // form the string prefix from the message name (cosmos.crypto.multisig.PubKey) and the threshold joined together
  prefix := fmt.Sprintf("%s/%d", proto.MessageName(multisig), multisig.Threshold)

  // use the standard AddressHash function
  return AddressHash(prefix, joinedAddresses)
}
``` 

## Consequences

### Positive
- a simple algorithm for generating addresses for new public keys and module accounts

### Negative
- addresses do not communicate key type, a prefixed approach would have done this

### Neutral
- protobuf message names are used as key type prefixes

## References
