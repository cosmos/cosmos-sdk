# MultiStore

TODO: reconcile this with everything ... would be nice to have this explanation
somewhere but where does it belong ? So far we've already showed how to use it
all by creating KVStore keys and calling app.MountStoresIAVL !


The Cosmos-SDK provides a special Merkle database called a `MultiStore` to be used for all application
storage. The MultiStore consists of multiple Stores that must be mounted to the
MultiStore during application setup. Stores are mounted to the MultiStore using a capabilities key, 
ensuring that only parts of the program with access to the key can access the store.

The goals of the MultiStore are as follows:

- Enforce separation of concerns at the storage level
- Restrict access to storage using capabilities
- Support multiple Store implementations in a single MultiStore, for instance the Tendermint IAVL tree and
  the Ethereum Patricia Trie
- Merkle proofs for various queries (existence, absence, range, etc.) on current and retained historical state
- Allow for iteration within Stores
- Provide caching for intermediate state during execution of blocks and transactions (including for iteration)

- Support historical state pruning and snapshotting

Currently, all Stores in the MultiStore must satisfy the `KVStore` interface,
which defines a simple key-value store. In the future, 
we may support more kinds of stores, such as a HeapStore
or a NDStore for multidimensional storage.

## Mounting Stores

Stores are mounted during application setup. To mount some stores, first create
their capability-keys:

```
fooKey := sdk.NewKVStoreKey("foo")
barKey := sdk.NewKVStoreKey("bar")
catKey := sdk.NewKVStoreKey("cat")
```

Stores are mounted directly on the BaseApp.
They can either specify their own database, or share the primary one already
passed to the BaseApp.

In this example, `foo` and `bar` will share the primary database, while `cat` will
specify its own:

```
catDB := dbm.NewMemDB()
app.MountStore(fooKey, sdk.StoreTypeIAVL)
app.MountStore(barKey, sdk.StoreTypeIAVL)
app.MountStoreWithDB(catKey, sdk.StoreTypeIAVL, catDB)
```

## Accessing Stores

In the Cosmos-SDK, the only way to access a store is with a capability-key.
Only modules given explicit access to the capability-key will 
be able to access the corresponding store. Access to the MultiStore is mediated
through the `Context`.

## Notes 

TODO: move this to the spec

In the example above, all IAVL nodes (inner and leaf) will be stored
in mainDB with the prefix of "s/k:foo/" and "s/k:bar/" respectively,
thus sharing the mainDB.  All IAVL nodes (inner and leaf) for the
cat KVStore are stored separately in catDB with the prefix of
"s/\_/".  The "s/k:KEY/" and "s/\_/" prefixes are there to
disambiguate store items from other items of non-storage concern.

