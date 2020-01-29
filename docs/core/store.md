<!--
order: 5
synopsis: A store is a data structure that holds the state of the application. 
-->

# Store

## Pre-requisite Readings {hide}

- [Anatomy of an SDK application](../basics/app-anatomy.md) {prereq}

## Introduction to SDK Stores

The Cosmos SDK comes with a large set of stores to persist the state of applications. By default, the main store of SDK applications is a `multistore`, i.e. a store of stores. Developers can add any number of key-value stores to the multistore, depending on their application needs. The multistore exists to support the modularity of the Cosmos SDK, as it lets each module declare and manage their own subset of the state. Key-value stores in the multistore can only be accessed with a specific capability `key`, which is typically held in the [`keeper`](../building-modules/keeper.md) of the module that declared the store. 

```
+-----------------------------------------------------+
|                                                     |
|    +--------------------------------------------+   |
|    |                                            |   |
|    |  KVStore 1 - Manage by keeper of Module 1  |
|    |                                            |   |
|    +--------------------------------------------+   |
|                                                     |
|    +--------------------------------------------+   |
|    |                                            |   |
|    |  KVStore 2 - Manage by keeper of Module 2  |   |
|    |                                            |   |
|    +--------------------------------------------+   |
|                                                     |
|    +--------------------------------------------+   |
|    |                                            |   |
|    |  KVStore 3 - Manage by keeper of Module 2  |   |
|    |                                            |   |
|    +--------------------------------------------+   |
|                                                     |
|    +--------------------------------------------+   |
|    |                                            |   |
|    |  KVStore 4 - Manage by keeper of Module 3  |   |
|    |                                            |   |
|    +--------------------------------------------+   |
|                                                     |
|    +--------------------------------------------+   |
|    |                                            |   |
|    |  KVStore 5 - Manage by keeper of Module 4  |   |
|    |                                            |   |
|    +--------------------------------------------+   |
|                                                     |
|                    Main Multistore                  |
|                                                     |
+-----------------------------------------------------+

                   Application's State
```

### Store Interface

At its very core, a Cosmos SDK `store` is an object that holds a `CacheWrapper` and implements a `GetStoreType()` method:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L12-L15

The `GetStoreType` is a simple method that returns the type of store, whereas a `CacheWrapper` is a simple interface that specifies cache-wrapping and `Write` methods:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L217-L238

Cache-wrapping is used ubiquitously in the Cosmos SDK and required to be implemented on every store type. A cache-wrapper creates a light snapshot of a store that can be passed around and updated without affecting the main underlying store. This is used to trigger temporary state-transitions that may be reverted later should an error occur. If a state-transition sequence is performed without issue, the cached store can be comitted to the underlying store at the end of the sequence. 

### Commit Store

A commit store is a store that has the ability to commit changes made to the underlying tree or db. The Cosmos SDK differentiates simple stores from commit stores by extending the basic store interfaces with a `Committer`:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L24-L28

The `Committer` is an interface that defines methods to persist changes to disk:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L17-L22

The `CommitID` is a deterministic commit of the state tree. Its hash is returned to the underlying consensus engine and stored in the block header. Note that commit store interfaces exist for various purposes, one of which is to make sure not every object can commit the store. As part of the [object-capabilities model](./ocap.md) of the Cosmos SDK, only `baseapp` should have the ability to commit stores. For example, this is the reason why the `ctx.KVStore()` method by which modules typically access stores returns a `KVStore` and not a `CommitKVStore`. 

The Cosmos SDK comes with many types of stores, the most used being [`CommitMultiStore`](#multistore), [`KVStore`](#kvstore) and [`GasKv` store](#gaskv-store). [Other types of stores](#other-stores) include `Transient` and `TraceKV` stores. 

## Multistore

### Multistore Interface

Each Cosmos SDK application holds a multistore at its root to persist its state. The multistore is a store of `KVStores` that follows the `Multistore` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L83-L112

If tracing is enabled, then cache-wrapping the multistore will wrap all the underlying `KVStore` in [`TraceKv.Store`](#tracekv-store) before caching them. 

### CommitMultiStore

The main type of `Multistore` used in the Cosmos SDK is `CommitMultiStore`, which is an extension of the `Multistore` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L120-L158

As for concrete implementation, the [`rootMulti.Store`] is the go-to implementation of the `CommitMultiStore` interface. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/rootmulti/store.go#L27-L43

The `rootMulti.Store` is a base-layer multistore built around a `db` on top of which multiple `KVStores` can be mounted, and is the default multistore store used in [`baseapp`](./baseapp.md). 

### CacheMultiStore

Whenever the `rootMulti.Store` needs to be cached-wrapped, a [`cachemulti.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/cachemulti/store.go) is used. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/cachemulti/store.go#L17-L28

`cachemulti.Store` cache wraps all substores in its constructor and hold them in `Store.stores`. `Store.GetKVStore()` returns the store from `Store.stores`, and `Store.Write()` recursively calls `CacheWrap.Write()` on all the substores.

## Base-layer KVStores

### `KVStore` and `CommitKVStore` Interfaces

A `KVStore` is a simple key-value store used to store and retrieve data. A `CommitKVStore` is a `KVStore` that also implements a `Committer`. By default, stores mounted in `baseapp`'s main `CommitMultiStore` are `CommitKVStore`s. The `KVStore` interface is primarily used to restrict modules from accessing  the committer. 

Individual `KVStore`s are used by modules to manage a subset of the global state. `KVStores` can be accessed by objects that hold a specific key. This `key` should only be exposed to the [`keeper`](../building-modules/keeper.md) of the module that defines the store. 

`CommitKVStore`s are declared by proxy of their respective `key` and mounted on the application's [multistore](#multistore) in the [main application file](../basics/app-anatomy.md#core-application-file).  In the same file, the `key` is also passed to the module's `keeper` that is responsible for managing the store. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/store.go#L163-L193

Apart from the traditional `Get` and `Set` methods, a `KVStore` is expected to implement an `Iterator()` method which returns an `Iterator` object. The `Iterator()` method is used to iterate over a domain of keys, typically keys that share a common prefix. Here is a common pattern of using an `Iterator` that might be found in a module's `keeper`:

```go
store := ctx.KVStore(keeper.storeKey)
iterator := sdk.KVStorePrefixIterator(store, prefix) // proxy for store.Iterator

defer iterator.Close()
for ; iterator.Valid(); iterator.Next() {
	var object types.Object
	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &object)

	if cb(object) {
        break
    }
}
```

### `IAVL` Store

The default implementation of `KVStore` and `CommitKVStore` used in `baseapp` is the `iavl.Store`.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/iavl/store.go#L32-L47

 `iavl` stores are based around an [IAVL Tree](https://github.com/tendermint/iavl), a self-balancing binary tree which guarantees that:

- `Get` and `Set` operations are O(log n), where n is the number of elements in the tree.
- Iteration efficiently returns the sorted elements within the range.
- Each tree version is immutable and can be retrieved even after a commit (depending on the pruning settings). 

The documentation on the IAVL Tree is located [here](https://github.com/tendermint/iavl/blob/f9d4b446a226948ed19286354f0d433a887cc4a3/docs/overview.md).

### `DbAdapter` Store

`dbadapter.Store` is a adapter for `dbm.DB` making it fulfilling the `KVStore` interface.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/dbadapter/store.go#L13-L16

`dbadapter.Store` embeds `dbm.DB`, meaning most of the `KVStore` interface functions are implemented. The other functions (mostly miscellaneous) are manually implemented. This store is primarily used within [Transient Stores](#transient-stores)

### `Transient` Store

`Transient.Store` is a base-layer `KVStore` which is automatically discarded at the end of the block.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/transient/store.go#L14-L17

`Transient.Store` is a `dbadapter.Store` with a `dbm.NewMemDB()`. All `KVStore` methods are reused. When `Store.Commit()` is called, a new `dbadapter.Store` is assigned, discarding previous reference and making it garbage collected.

This type of store is useful to persist information that is only relevant per-block. One example would be to store parameter changes (i.e. a bool set to `true` if a parameter changed in a block). 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/x/params/subspace/subspace.go#L24-L32

Transient stores are typically accessed via the [`context`](./context.md) via the `TransientStore()` method:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/context.go#L215-L218

## KVStore Wrappers

### CacheKVStore

`cachekv.Store` is a wrapper `KVStore` which provides buffered writing / cached reading functionalities over the underlying `KVStore`.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/cachekv/store.go#L26-L33

This is the type used whenever an IAVL Store needs to be cache-wrapped (typically when setting value that might be reverted later). 

#### `Get`

`Store.Get()` checks `Store.cache` first in order to find if there is any cached value associated with the key. If the value exists, the function returns it. If not, the function calls `Store.parent.Get()`, sets the key-value pair to the `Store.cache`, and returns it.

#### `Set`

`Store.Set()` sets the key-value pair to the `Store.cache`. `cValue` has the field dirty bool which indicates whether the cached value is different from the underlying value. When `Store.Set()` cache new pair, the `cValue.dirty` is set `true` so when `Store.Write()` is called it can be written to the underlying store.

#### `Iterator`

`Store.Iterator()` have to traverse on both caches items and the original items. In `Store.iterator()`, two iterators are generated for each of them, and merged. `memIterator` is essentially a slice of the `KVPairs`, used for cached items. `mergeIterator` is a combination of two iterators, where traverse happens ordered on both iterators.

### `GasKv` Store

Cosmos SDK applications use [`gas`](../basics/gas-fees.md) to track resources usage and prevent spam. [`GasKv.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/gaskv/store.go) is a `KVStore` wrapper that enables automatic gas consumption each time a read or write to the store is made. It is the solution of choice to track storage usage in Cosmos SDK applications.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/gaskv/store.go#L11-L17

When methods of the parent `KVStore` are called, `GasKv.Store` automatically consumes appropriate amount of gas depending on the `Store.gasConfig`:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/gas.go#L141-L150

By default, all `KVStores` are wrapped in `GasKv.Stores` when retrieved. This is done in the `KVStore()` method of the [`context`](./context.md):

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/types/context.go#L210-L213

In this case, the default gas configuration is used:

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/types/gas.go#L152-L163

### `TraceKv` Store

`tracekv.Store` is a wrapper `KVStore` which provides operation tracing functionalities over the underlying `KVStore`. It is applied automatically by the Cosmos SDK on all `KVStore` if tracing is enabled on the parent `MultiStore`. 

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/tracekv/store.go#L20-L43

When each `KVStore` methods are called, `tracekv.Store` automatically logs `traceOperation` to the `Store.writer`. `traceOperation.Metadata` is filled with `Store.context` when it is not nil. `TraceContext` is a `map[string]interface{}`.

### `Prefix` Store

`prefix.Store` is a wrapper `KVStore` which provides automatic key-prefixing functionalities over the underlying `KVStore`.

+++ https://github.com/cosmos/cosmos-sdk/blob/7d7821b9af132b0f6131640195326aa02b6751db/store/prefix/store.go#L17-L20

When `Store.{Get, Set}()` is called, the store forwards the call to its parent, with the key prefixed with the `Store.prefix`.

When `Store.Iterator()` is called, it does not simply prefix the `Store.prefix`, since it does not work as intended. In that case, some of the elements are traversed even they are not starting with the prefix.

## Next {hide}

Learn about [encoding](./encoding.md) {hide}