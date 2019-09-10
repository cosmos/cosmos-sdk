# ADR 012: State Accessors

## Changelog

- 2019 Sep 04: Initial draft

## Context

The application logic currently uses raw `KVStore` interface and `Codec` to access the state. While this gives large degree of freedom to the module writers, it is hard to write/read and modularize.

First, each time when a logic tries to access on the state, it have to marshal the value and set, or get the value and unmarshal. Usually this is done by declaring `Keeper.GetXXX` and `Keeper.SetXXX` functions, which are boilerplate codes hard to maintain.

Second, this makes harder to align with object capability theorem: the right to access on the state is defined as a `StoreKey`, which gives the full access on the entier Merkle tree, 
so a logic cannot send the access right to a specific key-value pair(or a subspace) to another logic safely.

Finally, because the getter/setter functions are defined as a method of `Keeper`, the reviewers have to consider the whole Merkle tree space when they reviewing a function accessing any part of the state. 
There is no static way to know which part of the state that the function wants to access.


> This section describes the forces at play, including technological, political, social, and project local. These forces are probably in tension, and should be called out as such. The language in this section is value-neutral. It is simply describing facts.
> {context body}

## Decision

We will define a type named `Value`:
```go
type Value struct {
  m Mapping
  key []byte
}
```

The `Value` works as a reference for a key-value pair in the state, where `Value.m` defines the key-value space it will access and `Value.key` defines the exact key for the reference.

We will define a type named `Mapping`:

```go
type Mapping struct {
  storeKey sdk.StoreKey
  cdc *codec.Codec
  prefix []byte
}
```

The `Mapping` works as a reference for a key-value space in the state, where `Mapping.storeKey` defines the IAVL tree and `Mapping.prefix` defines the optional subspace prefix.


We will give `Value` the following core methods:

```go
func (Value) Get(ctx Context, ptr interface{}) // Get and unmarshal stored data, noop if not exists, panic if cannot unmarshal
func (Value) GetSafe(ctx Context, ptr interface{}) // Get and unmarshal stored data, return error if not exists or cannot unmarshal
func (Value) Set(ctx Context, o interface{}) // Marshal and set argument
func (Value) Exists(ctx Context) bool // Check if value exists
func (Value) Delete(ctx Context) // Delete value
```

We will give `Mapping` the following methods:

```go
func (Mapping) Value(key []byte) Value // Constructs key-value pair reference corresponding to the key argument in the Mapping space
func (Mapping) Get(ctx Context, key []byte, ptr interface{}) // Get and unmarshal stored data, noop if not exists, panic if cannot unmarshal
func (Mapping) GetSafe(ctx Context, key []byte, ptr interface{}) // Get and unmarshal stored data, return error if not exists or cannot unmarshal
func (Mapping) Set(ctx Context, key []byte, o interface{}) // Marshal and set argument
func (Mapping) Has(ctx Context, key []byte) bool // Check if value exists
func (Mapping) Delete(ctx Context, key []byte) // Delete value
```

where `Mapping.{Get, GetSafe, Set, Has, Delete}(ctx Context, key []byte, args ...T)` are defined as `Mapping.Value(key).{Get, GetSafe, Set, Exists, Delete}(ctx Context, args ...T)`

We will define a family of types derived from `Value`:

```go
type Boolean struct { Value }
type Enum struct { Value }
type Integer struct { Value; enc IntEncoding }
type String struct { Value }
// extensible
```

where the encoding schemes can be different, `o` arguments in core methods are typed, `ptr` arguments in core methods are replaced by explicit return types.

We will define a family of types derived from `Mapping`:

```go
type Indexer struct { m Mapping, enc IntEncoding }
```

wherer the `key` arguments in core methods are typed.

Some of the properties of the accessor types are:
- State access happens only when a function which takes `ctx Context` as an argument is invoked
- Accessor type structs gives right to access to the state only that the struct is referring, no other
- Marshalling/Unmarshalling happens implicitly within the core methods

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Status

Proposed

> A decision may be "proposed" if the project stakeholders haven't agreed with it yet, or "accepted" once it is agreed. If a later ADR changes or reverses a decision, it may be marked as "deprecated" or "superseded" with a reference to its replacement.
> {Deprecated|Proposed|Accepted}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## References

#4554
- {reference link}
