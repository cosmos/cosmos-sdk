# Store

## Pre-requisite Reading

- [Anatomy of an SDK application](../basics/app-anatomy.md)

## Synopsis

A store is a data structure that holds the state of the application. 

- [Introduction to SDK Stores](#introduction-to-sdk-stores)
    + [Store Interface](#store-interface)
    + [Commit Store](#commit-store)
- [Multistore](#multistore)
    + [Multistore Interface](#multistore-interface)
    + [CommitMultiStore](#commitmultistore)
    + [CacheMultiStore](#cachemultistore)
- [Base Layer KvStores](#base-layer-kvstores)
    + [`KVStore` and `CommitKVStore` Interfaces](#kvstore-and-commitkvstore-interfaces)
    + [`IAVL` Store](#iavl-store)
    + [`DBAdapter` Store](#dbadapter-store)
    + [`Transient` Store](#transient-store)
- [KVStore Wrappers](#kvstore-wrappers)
    + [CacheKVStore](#cachekvstore)
    + [`GasKv` Store](#gaskv-store)
    + [`TraceKV` Store](#tracekv-store)
    + [`Prefix` Store](#prefix-store)

## Introduction to SDK Stores

The Cosmos SDK comes with a large set of stores to persist the state of applications. By default, the main store of SDK applications is a multistore, i.e. a store of stores. Developers can add any number of key-value stores to the multistore, depending on their application needs. The multistore exists to support the modularity of the Cosmos SDK, as it lets each module declare and manage their own subset of the state. Key-value stores in the multistore can be accessed with a specific `key`, which is typically held in the [`keeper`](../building-modules/keeper.md) of the module that declared the store. 

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

```go
type Store interface { 
	GetStoreType() StoreType
	CacheWrapper
}
```

The `GetStoreType` is a simple method that returns the type of store, whereas a `CacheWrapper` is a simple interface that specifies cache-wrapping and `Write` methods:

```go
type CacheWrap interface {
	// Write syncs with the underlying store.
	Write()

	// CacheWrap recursively wraps again.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace recursively wraps again with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}

type CacheWrapper interface { //nolint
	// CacheWrap cache wraps.
	CacheWrap() CacheWrap

	// CacheWrapWithTrace cache wraps with tracing enabled.
	CacheWrapWithTrace(w io.Writer, tc TraceContext) CacheWrap
}
```

Cache-wrapping is used ubiquitously in the Cosmos SDK and required to be implemented on every store type. A cache-wrapper creates a light snapshot of a store that can be passed around and updated without affecting the main underlying store. This is used to trigger temporary state-transitions that may be reverted later should an error occur. If a state-transition sequence is performed without issue, the cached store can be comitted to the main store at the end of the sequence. 

### Commit Store

A commit store is a store that has the ability to commit changes made to the underlying tree or db. The Cosmos SDK differentiates simple stores from commit stores by extending the basic store interfaces with a `Committer`:

```go
// Stores of MultiStore must implement CommitStore.
type CommitStore interface {
	Committer
	Store
}
```

The `Committer` is an interface that defines methods to persist changes to disk:

```go
// something that can persist to disk
type Committer interface {
	Commit() CommitID
	LastCommitID() CommitID
	SetPruning(PruningOptions)
}
```

The `CommitID` is a deterministic commit of the state tree. Its hash is returned to the underlying consensus engine and stored in the block header. Note that commit store interfaces exist for various purposes, one of which is to make sure not every object can commit the store. As part of the [object-capabilities model](./ocap.md) of the Cosmos SDK, only `baseapp` should have the ability to commit stores. For example, this is the reason why the `ctx.KVStore()` method by which modules typically access stores returns a `KVStore` and not a `CommitKVStore`. 

The Cosmos SDK comes with many types of stores, the most used being [`CommitMultiStore`](#multistore), [`KVStore`](#kvstore) and [`GasKv` store](#gaskv-store). [Other types of stores](#other-stores) include `Transient` and `TraceKV` stores. 

## Multistore

### Multistore Interface

Each Cosmos SDK application holds a multistore at its root to persist its state. The multistore is a store of `KVStores` that follows the `Multistore` interface:

```go
type MultiStore interface { //nolint
	Store

	// Cache wrap MultiStore.
	// NOTE: Caller should probably not call .Write() on each, but
	// call CacheMultiStore.Write().
	CacheMultiStore() CacheMultiStore

	// CacheMultiStoreWithVersion cache-wraps the underlying MultiStore where
	// each stored is loaded at a specific version (height).
	CacheMultiStoreWithVersion(version int64) (CacheMultiStore, error)

	// Convenience for fetching substores.
	// If the store does not exist, panics.
	GetStore(StoreKey) Store
	GetKVStore(StoreKey) KVStore

	// TracingEnabled returns if tracing is enabled for the MultiStore.
	TracingEnabled() bool

	// SetTracer sets the tracer for the MultiStore that the underlying
	// stores will utilize to trace operations. The modified MultiStore is
	// returned.
	SetTracer(w io.Writer) MultiStore

	// SetTracingContext sets the tracing context for a MultiStore. It is
	// implied that the caller should update the context when necessary between
	// tracing operations. The modified MultiStore is returned.
	SetTracingContext(TraceContext) MultiStore
}
```

If tracing is enabled, then cache-wrapping the multistore will wrap all the underlying `KVStore` in [`TraceKv.Store`](#tracekv-store) before caching them. 

### CommitMultiStore

The main type of `Multistore` used in the Cosmos SDK is `CommitMultiStore`, which is an extension of the `Multistore` interface:

```go
// A non-cache MultiStore.
type CommitMultiStore interface {
	Committer
	MultiStore

	// Mount a store of type using the given db.
	// If db == nil, the new store will use the CommitMultiStore db.
	MountStoreWithDB(key StoreKey, typ StoreType, db dbm.DB)

	// Panics on a nil key.
	GetCommitStore(key StoreKey) CommitStore

	// Panics on a nil key.
	GetCommitKVStore(key StoreKey) CommitKVStore

	// Load the latest persisted version. Called once after all calls to
	// Mount*Store() are complete.
	LoadLatestVersion() error

	// LoadLatestVersionAndUpgrade will load the latest version, but also
	// rename/delete/create sub-store keys, before registering all the keys
	// in order to handle breaking formats in migrations
	LoadLatestVersionAndUpgrade(upgrades *StoreUpgrades) error

	// LoadVersionAndUpgrade will load the named version, but also
	// rename/delete/create sub-store keys, before registering all the keys
	// in order to handle breaking formats in migrations
	LoadVersionAndUpgrade(ver int64, upgrades *StoreUpgrades) error

	// Load a specific persisted version. When you load an old version, or when
	// the last commit attempt didn't complete, the next commit after loading
	// must be idempotent (return the same commit id). Otherwise the behavior is
	// undefined.
	LoadVersion(ver int64) error

	// Set an inter-block (persistent) cache that maintains a mapping from
	// StoreKeys to CommitKVStores.
	SetInterBlockCache(MultiStorePersistentCache)
}
```

As for concrete implementation, the [`rootMulti.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/rootmulti/store.go) is the go-to implementation of the `CommitMultiStore` interface. The `rootMulti.Store` is a base-layer multistore built around a `db` on top of which multiple `KVStores` can be mounted, and is the default multistore store used in [`baseapp`](./baseapp.md). 

### CacheMultiStore

Whenever the `rootMulti.Store` needs to be cached-wrapped, a [`cachemulti.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/cachemulti/store.go) is used. 

```go
type Store struct {
    db types.CacheKVStore
    stores map[types.StoreKey] types.CacheWrap
}
```

`cachemulti.Store` cache wraps all substores in its constructor and hold them in `Store.stores`. `Store.GetKVStore()` returns the store from `Store.stores`, and `Store.Write()` recursively calls `CacheWrap.Write()` on all the substores.

## Base-layer KVStores

### `KVStore` and `CommitKVStore` Interfaces

A `KVStore` is a simple key-value store used to store and retrieve data. A `CommitKVStore` is a `KVStore` that also implements a `Committer`. By default, stores mounted in `baseapp`'s main `CommitMultiStore` are `CommitKVStore`s. The `KVStore` interface is primarily used to restrict modules from accessing  the committer . 

Individual `KVStore`s are used by modules to manage a subset of the global state. `KVStores` can be accessed by objects that hold a specific key. This `key` should only be exposed to the [`keeper`](../building-modules/keeper.md) of the module that defines the store. 

`CommitKVStore`s are declared by proxy of their respective `key` and mounted on the application's [multistore](#multistore) in the [main application file](../basics/app-anatomy.md#core-application-file).  In the same file, the `key` is also passed to the module's `keeper` that is responsible for managing the store. 

```go
type KVStore interface {
	Store

	// Get returns nil iff key doesn't exist. Panics on nil key.
	Get(key []byte) []byte

	// Has checks if a key exists. Panics on nil key.
	Has(key []byte) bool

	// Set sets the key. Panics on nil key or value.
	Set(key, value []byte)

	// Delete deletes the key. Panics on nil key.
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	Iterator(start, end []byte) Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// CONTRACT: No writes may happen within a domain while an iterator exists over it.
	// Exceptionally allowed for cachekv.Store, safe to write in the modules.
	ReverseIterator(start, end []byte) Iterator
}

// Stores of MultiStore must implement CommitStore.
type CommitKVStore interface {
	Committer
	KVStore
}
```

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

The default implementation of `KVStore` and `CommitKVStore` used in `baseapp` is the [`iavl.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/iavl/store.go). `iavl` stores are based around an [IAVL Tree](https://github.com/tendermint/iavl), a self-balancing binary tree which guarantees that:

- `Get` and `Set` operations are O(log n), where n is the number of elements in the tree.
- Iteration efficiently returns the sorted elements within the range.
- Each tree version is immutable and can be retrieved even after a commit (depending on the pruning settings). 

### `DbAdapter` Store

`dbadapter.Store` is a adapter for `dbm.DB` making it fulfilling the `KVStore` interface.

```go
type Store struct {
    dbm.DB
}
```

`dbadapter.Store` embeds `dbm.DB`, meaning most of the `KVStore` interface functions are implemented. The other functions (mostly miscellaneous) are manually implemented. This store is primarily used within [Transient Stores](#transient-stores)

### `Transient` Store

`Transient.Store` is a base-layer `KVStore` which is automatically discarded at the end of the block.

```go
type Store struct {
    dbadapter.Store
}
```

`Transient.Store` is a `dbadapter.Store` with a `dbm.NewMemDB()`. All `KVStore` methods are reused. When `Store.Commit()` is called, a new `dbadapter.Store` is assigned, discarding previous reference and making it garbage collected.

This type of store is useful to persist information that is only relevant per-block. One example would be to store [parameter changes](https://github.com/cosmos/cosmos-sdk/blob/master/x/params/subspace/subspace.go#L27) (i.e. a bool set to `true` if a parameter changed in a block). 

Transient stores are typically accessed via the [`context`](./context.md) via the `TransientStore()` method:

```go
// TransientStore fetches a TransientStore from the MultiStore.

func (c Context) TransientStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.TransientGasConfig())
}
```

## KVStore Wrappers

### CacheKVStore

`cachekv.Store` is a wrapper `KVStore` which provides buffered writing / cached reading functionalities over the underlying `KVStore`.

```go
type Store struct {
    cache map[string]cValue
    parent types.KVStore
}
```

This is the type used whenever an IAVL Store needs to be cache-wrapped (typically when setting value that might be reverted later). 

#### `Get`

`Store.Get()` checks `Store.cache` first in order to find if there is any cached value associated with the key. If the value exists, the function returns it. If not, the function calls `Store.parent.Get()`, sets the key-value pair to the `Store.cache`, and returns it.

#### `Set`

`Store.Set()` sets the key-value pair to the `Store.cache`. `cValue` has the field dirty bool which indicates whether the cached value is different from the underlying value. When `Store.Set()` cache new pair, the `cValue.dirty` is set `true` so when `Store.Write()` is called it can be written to the underlying store.

#### `Iterator`

`Store.Iterator()` have to traverse on both caches items and the original items. In `Store.iterator()`, two iterators are generated for each of them, and merged. `memIterator` is essentially a slice of the `KVPairs`, used for cached items. `mergeIterator` is a combination of two iterators, where traverse happens ordered on both iterators.

### `GasKv` Store

Cosmos SDK applications use [`gas`](../basics/accounts-fees-gas.md#fees-and-gas) to track resources usage and prevent spam. [`GasKv.Store`](https://github.com/cosmos/cosmos-sdk/blob/master/store/gaskv/store.go) is a `KVStore` wrapper that enables automatic gas consumption each time a read or write to the store is made. It is the solution of choice to track storage usage in Cosmos SDK applications.

```go
type Store struct {
    gasMeter types.GasMeter
    gasConfig types.GasConfig
    parent types.KVStore
}
```

When methods of the parent `KVStore` are called, `GasKv.Store` automatically consumes appropriate amount of gas depending on the `Store.gasConfig`:

```go
type GasConfig struct {
	HasCost          Gas
	DeleteCost       Gas
	ReadCostFlat     Gas
	ReadCostPerByte  Gas
	WriteCostFlat    Gas
	WriteCostPerByte Gas
	IterNextCostFlat Gas
}
```

By default, all `KVStores` are wrapped in `GasKv.Stores` when retrieved. This is done in the `KVStore()` method of the [`context`](./context.md):

```go
// KVStore fetches a KVStore from the MultiStore.
func (c Context) KVStore(key StoreKey) KVStore {
	return gaskv.NewStore(c.MultiStore().GetKVStore(key), c.GasMeter(), stypes.KVGasConfig())
}
```

In this case, the default gas configuration is used:

```go
func KVGasConfig() GasConfig {
	return GasConfig{
		HasCost:          1000,
		DeleteCost:       1000,
		ReadCostFlat:     1000,
		ReadCostPerByte:  3,
		WriteCostFlat:    2000,
		WriteCostPerByte: 30,
		IterNextCostFlat: 30,
	}
}
```

### `TraceKv` Store

`tracekv.Store` is a wrapper `KVStore` which provides operation tracing functionalities over the underlying `KVStore`. It is applied automatically by the Cosmos SDK on all `KVStore` if tracing is enabled on the parent `MultiStore`. 

```go
type Store struct {
    parent types.KVStore
    writer io.Writer
    context types.TraceContext
}
```

When each `KVStore` methods are called, `tracekv.Store` automatically logs `traceOperation` to the `Store.writer`.

```go
type traceOperation struct {
    Operation operation
    Key string
    Value string
    Metadata map[string]interface{}
} 
```

`traceOperation.Metadata` is filled with `Store.context` when it is not nil. `TraceContext` is a `map[string]interface{}`.

### `Prefix` Store

`prefix.Store` is a wrapper `KVStore` which provides automatic key-prefixing functionalities over the underlying `KVStore`.

```go
type Store struct {
    parent types.KVStore
    prefix []byte
}
```

When `Store.{Get, Set}()` is called, the store forwards the call to its parent, with the key prefixed with the `Store.prefix`.

When `Store.Iterator()` is called, it does not simply prefix the `Store.prefix`, since it does not work as intended. In that case, some of the elements are traversed even they are not starting with the prefix.

## Next

Learn about [encoding](./encoding.md).