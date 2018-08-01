# ICS XXX: Cosmos Signed Messages

>TODO: Replace with valid ICS number and possibly move to new location.

## Changelog

## Abstract

Having the ability to sign messages off-chain has proven to be a fundamental aspect
of nearly any blockchain. The notion of signing messages off-chain has many 
added benefits such as saving on computational costs and reducing transaction
throughput and overhead. Within the context of the Cosmos, some of the major
applications of signing such data includes, but is not limited to, providing a
cryptographic secure and verifiable means of proving validator identity and
possibly associating it with some other framework or organization. In addition,
having the ability to sign Cosmos messages with a Ledger or similar HSM device.

A standardized protocol for hashing, signing, and verifying messages that can be
implemented by the Cosmos SDK and other third-party organizations is needed. Such a
standardized protocol subscribes to the following:

* Contains a specification of machine-verifiable and human-readable typed structured data
* Contains a framework for deterministic and injective encoding of structured data
* Utilizes cryptographic secure hashing and signing algorithms
* A framework for supporting extensions and domain separation
* Is invulnerable to chosen ciphertext attacks
* Has protection against potentially signing transactions a user did not intend to

This specification is only concerned with the rationale and the standardized
implementation of Cosmos signed messages. It does **not** concern itself with the
concept of replay attacks as that will be left up to the higher-level application
implementation. If you view signed messages in the means of authorizing some
action or data, then such an application would have to either treat this as 
idempotent or have mechanisms in place to reject known signed messages.

## Specification

> The proposed implementation is motivated and borrows heavily from EIP-712<sup>1</sup>
and in general Ethereum's `eth_sign` and `eth_signTypedData` methods<sup>2</sup>.

### Preliminary

The Cosmos message signing protocol will be parameterized with a cryptographic
secure hashing algorithm `SHA-256` and a signing algorithm `S` that contains 
the operations `sign` and `verify` which provide a digital signature over a set
of bytes and verification of a signature respectively.

Note, our goal here is not to provide context and reasoning about why necessarily
these algorithms were chosen apart from the fact they are the defacto algorithms
used in Tendermint and the Cosmos SDK and that they satisfy our needs for such
cryptographic algorithms such as having resistance to collision and second
pre-image attacks, as well as being deterministic and uniform.

### Encoding

Our goal is to create a deterministic, injective, machine-verifiable means of
encoding human-readable typed and structured data.

Let us consider the set of signed messages to be: `B ∪ S`, where `B` is the set
of byte strings and `S` is the set of human-readable typed structures. Thus, the
set can can be encoded in a deterministic and injective way via the following
rules, where `||` denotes concatenation:

* `encode(b : B)` = `p1 || bytes("Signed Cosmos SDK Message: \n") || l || b`, where
  * `l`: little endian uint64 encoding of the length of `b`
  * `p1`: prefix to distinguish from normal transactions and other encoding cases
* `encode(s : S, domainSeparator : B)` = `p2 || domainSeparator || amino(s)`, where
  * `domainSeparator`: 32 byte encoding of the domain separator [see below](###DomainSeparator)
  * `p2`: prefix to distinguish from normal transactions and other encoding cases

> TODO: Figure out byte(s) prefix in the encoding to not have collisions with
typical transaction signatures (JSON-encoded) and to distinguish the individual
cases. This may require introducing prefixes to transactions.

### DomainSeparator

Encoding structures can still lead to potential collisions and while this may be
ok or even desired, it introduces a concern in that it could lead to two compatible
signatures. The domain separator prevents collisions of otherwise identical
structures. It is designed to unique per application use and is directly used in
the signature encoding itself. The domain separator is also extensible where the
protocol and application designer may introduce or omit fields to their liking,
but we will provide a typical structure that can be used for proper separation
of concerns:

```golang
type DomainSeparator struct {
    name    string  // A user readable name of the signing origin or application.
    chainID string  // The corresponding Cosmos chain identifier.
    version uint16  // Version of the domain separator. A single major version should suffice.
    salt    []byte  // Random data to further provide disambiguation.
}
```

Application designers may choose to omit or introduce additional fields to a
domain separator. However, users should provided with the exact information for
which they will be signing (i.e. a user should always know the chain ID they are
signing for).

Given the set of all domain separators, the encoding of the domain separator
is as follows:

* `encode(domainSeparator : B)` = `sha256(amino(domainSeparator))`

## API

We now formalize a standard set of APIs that providers must implement:

```
cosmosSignBytes(b : B, addr : B) {
    return secp256k1Sign(sha256(encode(b)), addressPrivKey(addr))
}
```

```
cosmosSignBytesPass(b : B, pass : B) {
    return secp256k1Sign(sha256(encode(b)), passPrivKey(addr))
}
```

```
cosmosSignTyped(s : S, domainSeparator : B, addr : B) {
     return secp256k1Sign(sha256(encode(s, domainSeparator)), addressPrivKey(addr))
}
```

```
cosmosSignTypedPass(s : S, domainSeparator : B, pass : B) {
     return secp256k1Sign(sha256(encode(s, domainSeparator)), passPrivKey(addr))
}
```

## References

1. https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
2. https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
