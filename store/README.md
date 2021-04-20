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

`Store.Iterator()` have to traverse on both caches items and the original items. In `Store.iterator()`, two iterators are generated for each of them, and merged. `memIterator` is essentially a slice of the `KVPair`s, used for cached items. `mergeIterator` is a combination of two iterators, where traverse happens ordered on both iterators.

## CacheMulti

`cachemulti.Store` is a wrapper `MultiStore` which provides buffered writing / cached reading functionalities over the underlying `MutliStore`

```go
type Store struct {
    db types.CacheKVStore
    stores map[types.StoreKey] types.CacheWrap
}
```

`cachemulti.Store` branches all substores in its constructor and hold them in `Store.stores`. `Store.GetKVStore()` returns the store from `Store.stores`, and `Store.Write()` recursively calls `CacheWrap.Write()` on the substores.

## DBAdapter

`dbadapter.Store` is a adapter for `dbm.DB` making it fulfilling the `KVStore` interface.

```go
type Store struct {
    dbm.DB
}
```

`dbadapter.Store` embeds `dbm.DB`, so most of the `KVStore` interface functions are implemented. The other functions(mostly miscellaneous) are manually implemented.

## IAVL

`iavl.Store` is a base-layer self-balancing merkle tree. It is guaranteed that 

1. Get & set operations are `O(log n)`, where `n` is the number of elements in the tree
2. Iteration efficiently returns the sorted elements within the range
3. Each tree version is immutable and can be retrieved even after a commit(depending on the pruning settings)

Specification and implementation of IAVL tree can be found in [https://github.com/tendermint/iavl].

## GasKV

`gaskv.Store` is a wrapper `KVStore` which provides gas consuming functionalities over the underlying `KVStore`.

```go
type Store struct {
    gasMeter types.GasMeter
    gasConfig types.GasConfig
    parent types.KVStore
}
```

When each `KVStore` methods are called, `gaskv.Store` automatically consumes appropriate amount of gas depending on the `Store.gasConfig`.


## Prefix

`prefix.Store` is a wrapper `KVStore` which provides automatic key-prefixing functionalities over the underlying `KVStore`.

```go
type Store struct {
    parent types.KVStore
    prefix []byte
}
```

When `Store.{Get, Set}()` is called, the store forwards the call to its parent, with the key prefixed with the `Store.prefix`.

When `Store.Iterator()` is called, it does not simply prefix the `Store.prefix`, since it does not work as intended. In that case, some of the elements are traversed even they are not starting with the prefix.

## RootMulti

`rootmulti.Store` is a base-layer `MultiStore` where multiple `KVStore` can be mounted on it and retrieved via object-capability keys. The keys are memory addresses, so it is impossible to forge the key unless an object is a valid owner(or a receiver) of the key, according to the object capability principles.

## TraceKV

`tracekv.Store` is a wrapper `KVStore` which provides operation tracing functionalities over the underlying `KVStore`.

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

## Transient

`transient.Store` is a base-layer `KVStore` which is automatically discarded at the end of the block.

```go
type Store struct {
    dbadapter.Store
}
```

`Store.Store` is a `dbadapter.Store` with a `dbm.NewMemDB()`. All `KVStore` methods are reused. When `Store.Commit()` is called, new `dbadapter.Store` is assigned, discarding previous reference and making it garbage collected.
