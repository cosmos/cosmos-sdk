# ADR 012: State Accessors

## Changelog

- 2019 Sep 04: Initial draft

## Context

SDK modules currently use the `KVStore` interface and `Codec` to access their respective state. While
this provides a large degree of freedom to module developers, it is hard to modularize and the UX is
mediocre.

First, each time a module tries to access the state, it has to marshal the value and set or get the
value and finally unmarshal. Usually this is done by declaring `Keeper.GetXXX` and `Keeper.SetXXX` functions,
which are repetitive and hard to maintain.

Second, this makes it harder to align with the object capability theorem: the right to access the
state is defined as a `StoreKey`, which gives full access on the entire Merkle tree, so a module cannot
send the access right to a specific key-value pair (or a set of key-value pairs) to another module safely.

Finally, because the getter/setter functions are defined as methods of a module's `Keeper`, the reviewers
have to consider the whole Merkle tree space when they reviewing a function accessing any part of the state.
There is no static way to know which part of the state that the function is accessing (and which is not).

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
func (Value) Get(ctx Context, ptr interface{}) {} // Get and unmarshal stored data, noop if not exists, panic if cannot unmarshal
func (Value) GetSafe(ctx Context, ptr interface{}) {} // Get and unmarshal stored data, return error if not exists or cannot unmarshal
func (Value) GetRaw(ctx Context) []byte {} // Get stored data as raw byteslice
func (Value) Set(ctx Context, o interface{}) {} // Marshal and set argument
func (Value) Exists(ctx Context) bool {} // Check if value exists
func (Value) Delete(ctx Context) {} // Delete value
```

We will give `Mapping` the following core methods:

```go
func (Mapping) Value(key []byte) Value {} // Constructs key-value pair reference corresponding to the key argument in the Mapping space
func (Mapping) Get(ctx Context, key []byte, ptr interface{}) {} // Get and unmarshal stored data, noop if not exists, panic if cannot unmarshal
func (Mapping) GetSafe(ctx Context, key []byte, ptr interface{}) {} // Get and unmarshal stored data, return error if not exists or cannot unmarshal
func (Mapping) GetRaw(ctx Context, key []byte) []byte {} // Get stored data as raw byteslice
func (Mapping) Set(ctx Context, key []byte, o interface{}) {} // Marshal and set argument
func (Mapping) Has(ctx Context, key []byte) bool {} // Check if value exists
func (Mapping) Delete(ctx Context, key []byte) {} // Delete value
```

where `Mapping.{Get, GetSafe, GetRaw, Set, Has, Delete}(ctx, key, args...)` are defined as `Mapping.Value(key).{Get, GetSafe, GetRaw, Set, Exists, Delete}(ctx, args...)`

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
// extensible
```

wherer the `key` arguments in core methods are typed.

Some of the properties of the accessor types are:
- State access happens only when a function which takes `ctx Context` as an argument is invoked
- Accessor type structs gives right to access to the state only that the struct is referring, no other
- Marshalling/Unmarshalling happens implicitly within the core methods

## Status

Proposed

## Consequences

### Positive

- Serialization will be done automatically
- Shorter code size, less boilerplate
- References to the state can be transfered safely
- Explicit scope of accessing

### Negative

- Serialization format will be hidden
- Different architecture from the current
- Type-specific types(`Boolean`, `Integer`...) have to be defined manually

### Neutral

## References

#4554
