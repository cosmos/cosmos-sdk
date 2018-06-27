# Transactions

In the previous app we built a simple `bank` with one message type for sending
coins and one store for storing accounts.
Here we build `App2`, which expands on `App1` by introducing 

- a new message type for issuing new coins
- a new store for coin metadata (like who can issue coins)
- a requirement that transactions include valid signatures

Along the way, we'll be introduced to Amino for encoding and decoding
transactions and to the AnteHandler for processing them.


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

## Amino

Now that we have two implementations of `Msg`, we won't know before hand 
which type is contained in a serialized `Tx`. Ideally, we would use the
`Msg` interface inside our `Tx` implementation, but the JSON decoder can't
decode into interface types. In fact, there's no standard way to unmarshal 
into interfaces in Go. This is one of the primary reasons we built 
[Amino](https://github.com/tendermint/go-amino) :).

While SDK developers can encode transactions and state objects however they
like, Amino is the recommended format. The goal of Amino is to improve over the latest version of Protocol Buffers,
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

TODO: JSON, types table

## Tx

TODO

## AnteHandler

Now that we have an implementation of `Tx` that includes more than just the Msgs, 
we need to specify how that extra information is validated and processed. This
is the role of the `AnteHandler`. The word `ante` here denotes "before", as the
`AnteHandler` is run before a `Handler`. While an app may have many Handlers,
one for each set of messages, it may have only a single `AnteHandler` that
corresponds to its single implementation of `Tx`.


The AnteHandler resembles a Handler:


```go
type AnteHandler func(ctx Context, tx Tx) (newCtx Context, result Result, abort bool)
```

Like Handler, AnteHandler takes a Context that restricts its access to stores
according to whatever capability keys it was granted. Instead of a `Msg`,
however, it takes a `Tx`.

Like Handler, AnteHandler returns a `Result` type, but it also returns a new
`Context` and an `abort bool`. TODO explain (do we still need abort? )

For `App2`, we simply check if the PubKey matches the Address, and the Signature validates with the PubKey:

```go
TODO
```

## App2

Let's put it all together now to get App2:

```go
TODO
```

## Conclusion

We've expanded on our first app by adding a new message type for issuing coins,
and by checking signatures. We learned how to use Amino for decoding into
interface types, allowing us to support multiple Msg types, and we learned how
to use the AnteHandler to validate transactions.

Unfortunately, our application is still insecure, because any valid transaction
can be replayed multiple times to drain someones account! Besides, validating
signatures and preventing replays aren't things developers should have to think
about.

In the next section, we introduce the built-in SDK modules `auth` and `bank`,
which respectively provide secure implementations for all our transaction authentication
and coin transfering needs.
