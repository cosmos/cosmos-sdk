# ICS XXX: Cosmos Signed Messages

>TODO: Replace with valid ICS number and possibly move to new location.

  * [Changelog](#changelog)
  * [Abstract](#abstract)
  * [Specification](#specification)
    + [Preliminary](#preliminary)
    + [Encoding](#encoding)
      - [Schema](#schema)
      - [encodeStruct](#encodestruct)
      - [encodeSchema](#encodeschema)
      - [encodeData](#encodedata)
    + [DomainSeparator](#domainseparator)
  * [API](#api)
  * [References](#references)

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

Let us consider the set of signed messages to be: `B ∪ TS`, where `B` is the set
of byte arrays and `TS` is the set of human-readable typed structures. Thus, the
set can can be encoded in a deterministic and injective way via the following
rules, where `||` denotes concatenation:

* `encode(b : B)` = `0x0000 || bytes("Signed Cosmos SDK Message: \n") || l || b`, where
  * `b`: the bytes to be signed
  * `l`: little endian uint64 encoding of the length of `b`
* `encode(ds : TS, ts : TS)` = `0x000001 || encodeStruct(ds) || encodeStruct(ts)`, where
  * `ds`: the application domain separator which is also a human-readable typed structure ([see below](#domainseparator))
  * `ts`: the human-readable typed structure to be signed

The prefix bytes disambiguate the encoding cases from one another as well as
separating them from collision of transactions to be signed. The `amino`
serialization protocol escapes the set of disambiguation and prefix bytes with a
single `0x00` byte so there should be no collision with those structures.

#### Schema

To achieve deterministic and injective encoding, Cosmos signed messages over
type structures will use an existing known standard -- [JSON schema](http://json-schema.org/).
The domain separator and typed structures to be encoded must be specified with
a schema adhering to the JSON schema [specification](http://json-schema.org/specification.html).

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "cosmos/signing/typeData/schema",
  "title": "The Cosmos signed message typed data schema.",
  "type": "object",
  "definitions": {
    "typeDef": {
      "type": "array",
      "items": {
        "description": "The list of properties a schema definition contains.",
        "type": "object",
        "properties": {
          "name": {
            "description": "The name of the schema rule.",
            "type": "string",
            "minLength": 1
          },
          "description": {
            "description": "The description of the schema rule.",
            "type": "string"
          },
          "type": {
            "description": "The type of the schema rule.",
            "type": "string",
            "minLength": 1
          }
        },
        "required": [
          "name",
          "type"
        ]
      }
    }
  },
  "properties": {
    "types": {
      "type": "object",
      "properties": {
        "CosmosDomain": {
          "description": "The application domain separator schema.",
          "$ref": "#/definitions/typeDef"
        }
      },
      "additionalProperties": {
        "description": "The application type schemas.",
        "$ref": "#/definitions/typeDef"
      },
      "required": [
        "CosmosDomain"
      ]
    },
    "primaryType": {
      "description": "The name of primary message type to sign. This is needed for when there is more than one type defined besides `CosmosDomain`.",
      "type": "string",
      "minLength": 1
    },
    "domain": {
      "description": "The domain separator to sign.",
      "type": "object"
    },
    "message": {
      "description": "The message data to sign.",
      "type": "object"
    }
  },
  "required": [
    "types",
    "primaryType",
    "domain",
    "message"
  ]
}
```

#### encodeStruct

The specification of encoding a human-readable typed structures, which includes 
the domain separator, is as follows, where `||` denotes concatenation:

`encodeStruct(ts : TS)` = `sha256(sha256(encodeSchema(ts)) || sha256(encodeData(ts)))`

**Note**: The typed structure `ts` should include the JSON instance and schema.

#### encodeSchema

The **schema** of a typed structure is encoded as the name of the type and the
concatenation of it's schema fields. If the schema internally references other
schemas, then those are appended to the encoding. In order to make the encoding
deterministic, the encoding should sort types by their type name in lexicographic
ascending order. The specification is as follows, where `||` denotes concatenation:

`encodeSchema(ts : TS)` = `"typeName(" || name || " " || type || ", " || ... || ")"`

e.g.
```json
"Address(address string)Coin(amount integer, denom string)Transaction(coin Coin, from Address, to Address)"
```

##### Alternatives

1. Instead of concatenating schema signatures, we could also take each type
definition and insert them into a sorted JSON object and hash the byte
representation of that object. This would avoid having to perform weird
concatenations.

#### encodeData

The **data** of a typed structure is encoded as the concatenation of values in
the typed data sorted by the field names in lexicographic ascending order. The
specification is as follows, where `||` denotes concatenation:

`encodeData(ts : TS)` = <code>sha256(value<sub>1</sub>) || sha256(value<sub>2</sub>) || ... || sha256(value<sub>n</sub>)`</code>, where <code>value<sub>1</sub> < value<sub>2</sub></code>

### DomainSeparator

Encoding structures can still lead to potential collisions and while this may be
okay or even desired, it introduces a concern in that it could lead to two compatible
signatures. The domain separator prevents collisions of otherwise identical
structures. It is designed to be unique per application use and is directly used
in the signature encoding itself. The domain separator is also extensible where
the protocol and application designer may introduce or omit fields to their needs,
but we will provide a "standard" structure that can be used for proper separation
of concerns:

```json
{
  "types": {
    "CosmosDomain": [
      {
        "name": "name",
        "description": "The name of the signing origin or application.",
        "type": "string"
      },
      {
        "name": "chainID",
        "description": "The corresponding Cosmos chain identifier.",
        "type": "string",
      },
      {
        "name": "version",
        "description": "The major version of the domain separator.",
        "type": "integer",
      },
    ],
  },
}
```

**Note**: The user-agent should refuse signing if the `chainID` does not match
the currently active chain!

## API

Application developers and designers should formalize a standard set of APIs that
adhere to the following specification:

<hr>

**cosmosSignBytes**

Params:
* `data`: arbitrary byte length data to sign
* `address`: 20 byte account address to sign data with

Returns:
* `signature`: the Cosmos signature derived using signing algorithm `S`

<hr>

**cosmosSignBytesPass**

Params:
* `data`: arbitrary byte length data to sign
* `address`: 20 byte account address to sign data with
* `password`: password of the account to sign data with

Returns:
* `signature`: the Cosmos signature derived using signing algorithm `S`

<hr>

**cosmosSignTyped**

Params:
* `typedData`: type typed data structure, including the domain separator, to encode and sign
* `address`: 20 byte account address to sign data with

Returns:
* `signature`: the Cosmos signature derived using signing algorithm `S`

<hr>

**cosmosSignTypedPass**

Params:
* `typedData`: type typed data structure, including the domain separator, to encode and sign
* `address`: 20 byte account address to sign data with
* `password`: password of the account to sign data with

Returns:
* `signature`: the Cosmos signature derived using signing algorithm `S`

## References

1. https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
2. https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
