# Amino

In the previous app we build a simple `bank` with one message type for sending
coins and one store for storing accounts.
Here we build `App2`, which expands on `App1` by introducing another message type for issuing new coins, and another store
for storing information about who can issue coins and how many.

`App2` will allow us to better demonstrate the security model of the SDK, 
using object-capability keys to determine which handlers can access which
stores.

Having multiple implementations of `Msg` also requires a better transaction
decoder, since we won't know before hand which type is contained in the
serialized `Tx`. In effect, we'd like to unmarshal directly into the `Msg`
interface, but there's no standard way to unmarshal into interfaces in Go.
This is what Amino is for :)


## Message

Let's introduce a new message type for issuing coins:

```go
TODO
```

## Handler

We'll need a new handler to support the new message type:

```go
TODO
```

## BaseApp

```go
TODO
```

## Amino

The SDK is flexible about serialization - application developers can use any
serialization scheme to encode transactions and state. However, the SDK provides
a native serialization format called
[Amino](https://github.com/tendermint/go-amino).

The goal of Amino is to improve over the latest version of Protocol Buffers,
`proto3`. To that end, Amino is compatible with the subset of `proto3` that
excludes the `oneof` keyword.

While `oneof` provides union types, Amino aims to provide interfaces.
The main difference being that with union types, you have to know all the types
up front. But anyone can implement an interface type whenever and however 
they like.

To implement interface types, Amino allows any concrete implementation of an
interface to register a globally unique name that is carried along whenever the
type is serialized. This allows Amino to seamlessly deserialize into interface
types!

The primary use for Amino in the SDK is for messages that implement the
`Msg` interface. By registering each message with a distinct name, they are each
given a distinct Amino prefix, allowing them to be easily distinguished in
transactions.

Amino can also be used for persistent storage of interfaces.

To use Amino, simply create a codec, and then register types:

```
cdc := wire.NewCodec()

cdc.RegisterConcrete(MsgSend{}, "cosmos-sdk/Send", nil)
cdc.RegisterConcrete(MsgIssue{}, "cosmos-sdk/Issue", nil)
```
