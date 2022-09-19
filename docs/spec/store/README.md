# Store

The store package defines the interfaces, types and abstractions for Cosmos SDK
modules to read and write to merkleized state within a Cosmos SDK application.
The store package provides many primitives for developers to use in order to
work with both state storage and state commitment. Below we describe the various
abstractions.

## Types

### `Store`

The bulk of the store interfaces are defined [here](https://github.com/cosmos/cosmos-sdk/blob/main/store/types/store.go),
where the base primitive interface, for which other interfaces build off of, is
the `Store` type. The `Store` interface defines the ability to tell the type of
the implementing store and the ability to cache wrap via the `CacheWrapper` interface.

### `CacheWrapper` & `CacheWrap`

One of the most important features a store has the ability to perform is the
ability to cache wrap. Cache wrapping is essentially the underlying store wrapping
itself within another store type that performs caching for both reads and writes
with the ability to flush writes via `Write`.

TODO: Write about the CacheKV store


### `KVStore`

One of the most important interfaces that both developers and modules interface
with, which also provides the basis of most state storage and commitment operations,
is the `KVStore`. The `KVStore` interface provides basic CRUD abilities and
prefix-based iteration, including reverse iteration.

Typically, each module has it's own dedicated `KVStore` instance, which it can
get access to via the `sdk.Context` and the use of a pointer-based named key,
`KVStoreKey`. The `KVStoreKey` provides pseudo-OCAP. How a `KVStoreKey` maps
exactly to a `KVStore` will be illustrated below through the `CommitMultiStore`.
