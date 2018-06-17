
## Storage

### MultiStore

MultiStore is like a root filesystem of an operating system, except
all the entries are fully Merkleized.  You mount onto a MultiStore
any number of Stores.  Currently only KVStores are supported, but in
the future we may support more kinds of stores, such as a HeapStore
or a NDStore for multidimensional storage.

The MultiStore as well as all mounted stores provide caching (aka
cache-wrapping) for intermediate state (aka software transactional
memory) during the execution of transactions.  In the case of the
KVStore, this also works for iterators.  For example, after running
the app's AnteHandler, the MultiStore is cache-wrapped (and each
store is also cache-wrapped) so that should processing of the
transaction fail, at least the transaction fees are paid and
sequence incremented.

The MultiStore as well as all stores support (or will support)
historical state pruning and snapshotting and various kinds of
queries with proofs.

### KVStore

Here we'll focus on the IAVLStore, which is a kind of KVStore.

IAVLStore is a fast balanced dynamic Merkle store that also supports
iteration, and of course cache-wrapping, state pruning, and various
queries with proofs, such as proofs of existence, absence, range,
and so on.

Here's how you mount them to a MultiStore.

```go
mainDB, catDB := dbm.NewMemDB(), dbm.NewMemDB()
fooKey := sdk.NewKVStoreKey("foo")
barKey := sdk.NewKVStoreKey("bar")
catKey := sdk.NewKVStoreKey("cat")
ms := NewCommitMultiStore(mainDB)
ms.MountStoreWithDB(fooKey, sdk.StoreTypeIAVL, nil)
ms.MountStoreWithDB(barKey, sdk.StoreTypeIAVL, nil)
ms.MountStoreWithDB(catKey, sdk.StoreTypeIAVL, catDB)
```

In the example above, all IAVL nodes (inner and leaf) will be stored
in mainDB with the prefix of "s/k:foo/" and "s/k:bar/" respectively,
thus sharing the mainDB.  All IAVL nodes (inner and leaf) for the
cat KVStore are stored separately in catDB with the prefix of
"s/\_/".  The "s/k:KEY/" and "s/\_/" prefixes are there to
disambiguate store items from other items of non-storage concern.



## 

Mounting an IAVLStore
TODO:

IAVLStore: Fast balanced dynamic Merkle store.
supports iteration.
MultiStore: multiple Merkle tree backends in a single store
allows using Ethereum Patricia Trie and Tendermint IAVL in same app
Provide caching for intermediate state during execution of blocks and
transactions (including for iteration)
Historical state pruning and snapshotting.
Query proofs (existence, absence, range, etc.) on current and retained
historical state.
