# Store

## CacheKV

`cachekv.Store` is a wrapper `KVStore` which provides buffered writing / cached reading functionalities over the underlying `KVStore`. 

```go
type Store struct {
    cache map[string]cValue
    parent types.KVStore
}
```

### Get

`Store.Get()` checks `Store.cache` first in order to find if there is any cached value associated with the key. If the value exists, the function returns it. If not, the function calls `Store.parent.Get()`, sets the key-value pair to the `Store.cache`, and returns it.

### Set

`Store.Set()` sets the key-value pair to the `Store.cache`. `cValue` has the field `dirty bool` which indicates whether the cached value is different from the underlying value. When `Store.Set()` cache new pair, the `cValue.dirty` is set true so when `Store.Write()` is called it can be written to the underlying store.

### Iterator

`Store.Iterator()` have to traverse on both caches items and the original items. In `Store.iterator()`, two iterators are generated for each of them, and merged. `memIterator` is essentially a slice of the `KVPair`s, used for cached items. `mergeIterator` is a combination of two iterators, 

## CacheMulti

`cachemulti.Store` is a wrapper `MultiStore` which provides buffered writing / cached reading functionalities over the underlying `MutliStore`

```go
type Store struct {
    db types.CacheKVStore
    stores map[types.StoreKey] types.CacheWrap
}
```

`cachemulti.Store` cache wraps all substores in its constructor and hold them in `Store.stores`. `Store.GetKVStore()` returns the store from `Store.stores`, and `Store.Write()` recursively calls `CacheWrap.Write()` on the substores.

## DBAdapter

`dbadapter.Store` is a adapter for `dbm.DB` making it fulfilling the `KVStore` interface.

```go
type Store struct {
    dbm.DB
}
```

`dbadapter.Store` embeds `dbm.DB`, so most of the `KVStore` interface functions are implemented. The other functions(mostly miscellaneous) are manually implemented.

## IAVL

`iavl.Store` is a base-layer self-balancing merkle tree.

### Get

### Set

### Iterator

## GasKV

`gaskv.Store` is a wrapper `KVStore` which provides gas consuming functionalities over the underlying `KVStore`.

## Prefix

`prefix.Store` is a wrapper `KVStore` which provides automatic key-prefixing functionalities over the underlying `KVStore`.

```go
type Store struct {
    parent types.KVStore
    prefix []byte
}
```

When `Store.{Get, Set}()` is called, the store forwards the call to its parent, with the key prefixed with the `Store.prefix`.

When `Store.Iterator()` is called, it does not simply prefix the `Store.prefix`, since it does not work as intended. 

## RootMulti

`rootmulti.Store` is a base-layer `MultiStore` where multiple `KVStore` can be mounted on it and retrieved via object-capability keys

## TraceKV

`tracekv.Store` is a wrapper `KVStore` which provides operation tracing functionalities over the underlying `KVStore`.

## Transient

`transient.Store` is a base-layer `KVStore` which is automatically discarded at the end of the block.

```go
type Store struct {
    dbadapter.Store
}
```

`Store.Store` is a `dbadapter.Store` with a `dbm.NewMemDB()`. All `KVStore` methods are reused. When `Store.Commit()` is called, new `dbadapter.Store` is assigned, discarding previous reference and making it garbage collected.
