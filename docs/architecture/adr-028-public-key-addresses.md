# ADR 028: Public Key Addresses

## Changelog

- 2020/08/18: Initial version

## Status

Proposed

## Context

Issue [\#3685](https://github.com/cosmos/cosmos-sdk/issues/3685) identified that public key
addresses are currently overlapping. One initial proposal was extending the address length and
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

`secp256k1` and multisig public keys are currently in use in existing Cosmos SDK zones. We
don't want to change existing addresses. So the addresses for these two key types will remain the same.

The current multisig public keys use amino serialization to generate the address. We will retain
those public keys and their address formatting, and call them "legacy amino" multisig public keys
in protobuf. We will also create multisig public keys without amino addresses to be described below.

### Canonical Public Key Addresses

Following on the discussion in [\#5694](https://github.com/cosmos/cosmos-sdk/issues/5694), we propose the
following approach.

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
key type prefix when creating addresses.

We define the canonical address format for new (non-legacy) public keys as
`Sha256(fmt.Sprintf("%s/%x, proto.MessageName(key), key.Bytes())[:20]`. This takes
the first 20 bytes of an SHA-256 hash of a string with the proto message name for the key
type joined by an `/` with the hex encoding of the key bytes.

For all basic public keys, key bytes should just be the raw unencoded public key bytes.

### Multisig Addresses

For new multisig public keys, we define a custom address format not based on any encoding scheme
(amino or protobuf).

First we define a proto message for multisig public keys:
```proto
package cosmos.crypto.multisig;

message PubKey {
  uint32 threshold = 1;
  repeated google.protobuf.Any public_keys = 2;
}
```

Each nested public key has its own address, so we can use that address as a starting
point for forming the multisig address. Let's create an array of strings, `hexAddresses []string`,
which is the hex-encoded address of each nested pubkey. We join these hex encoded addresses
with a `/`, i.e. `joinedHexAddresses := strings.Join(hexAddresses, "/")`. We then form the address of the multisig pubkey,
using `Sha256(fmt.Sprintf("cosmos.crypto.multisig.PubKey/%d/%s", pk.Threshold, joinedHexAddresses)[:20]`.

## Consequences

### Positive
- a simple algorithm for generating addresses for new public keys

### Negative
- addresses do not communicate key type, a prefixed approach would have done this

### Neutral
- protobuf message names are used as key type prefixes
- public key bytes are hex encoded before generating addresses

## References
