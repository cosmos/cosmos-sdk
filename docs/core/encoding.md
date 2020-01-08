<!--
order: 6
synopsis: Amino and Protobuf are the primary binary wire encoding schemes in the Cosmos SDK.
-->

# Encoding

> NOTE: This document a WIP.

## Pre-requisite Readings {hide}

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}

## Encoding

The Cosmos SDK utilizes two binary wire encoding protocols, [Amino](https://github.com/tendermint/go-amino/)
and [Protocol Buffers](https://developers.google.com/protocol-buffers), where Amino
is an object encoding specification. It is a subset of Proto3 with an extension for
interface support. See the [Proto3 spec](https://developers.google.com/protocol-buffers/docs/proto3)
for more information on Proto3, which Amino is largely compatible with (but not with Proto2).

Due to Amino having significant performance drawbacks, being reflection-based, and
not having any meaningful cross-language/client support, Protocol Buffers, specifically
[gogoprotobuf](https://github.com/gogo/protobuf/), is being used in place of Amino.
Note, this process of using Protocol Buffers over Amino is still an ongoing process.

Binary wire encoding of types in the Cosmos SDK can be broken down into two main
categories, client encoding and store encoding. Client encoding mainly revolves
around transaction processing and signing, whereas store encoding revolves around
types used in state-machine transitions and what is ultimately stored in the Merkle
tree.

For store encoding, protobuf definitions can exist for any type and will typically
have an Amino-based "intermediary" type. Specifically, the protobuf-based type
definition is used for serialization and persistence, whereas the Amino-based type
is used for business logic in the state-machine where they may converted back-n-forth.
Note, the Amino-based types may slowly be phased-out in the future so developers
should take note to use the protobuf message definitions where possible.

### Amino

Every module uses an Amino codec to serialize types and interfaces. This codec typically
has types and interfaces registered in that module's domain only (e.g. messages),
but there are exceptions like `x/gov`. Each module exposes a `RegisterCodec` function
that allows a user to provide a codec and have all the types registered. An application
will call this method for each necessary module.

Where there is no protobuf-based type definition for a module, Amino is used to encode
and decode raw wire bytes to the concrete type or interface:

```go
bz := keeper.cdc.MustMarshalBinaryBare(typeOrInterface)
keeper.cdc.MustUnmarshalBinaryBare(bz, &typeOrInterface)
```

Note, there are length-prefixed variants of the above functionality and this is
typically used for when the data needs to be streamed or grouped together
(e.g. `ResponseDeliverTx.Data`)

Another important use of the Amino is the encoding and decoding of
[transactions](./transactions.md). Transactions are defined by the application or
the SDK, but passed to the underlying consensus engine in order to be relayed to
other peers. Since the underlying consensus engine is agnostic to the application,
it only accepts transactions in the form of raw bytes. The encoding is done by an
object called `TxEncoder` and the decoding by an object called `TxDecoder`.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/tx_msg.go#L45-L49

A standard implementation of both these objects can be found in the [`auth` module](https://github.com/cosmos/cosmos-sdk/blob/master/x/auth):

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/auth/types/stdtx.go#L241-L266

### Gogoproto

> TODO: Add section on gogoproto semantics and design decisions that are used in the SDK.

## Next {hide}

Learn about [events](./events.md) {hide}
