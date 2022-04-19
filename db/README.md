# Key-Value Database

Databases supporting mappings of arbitrary byte sequences.

## Interfaces

The database interface types consist of objects to encapsulate the singular connection to the DB, transactions being made to it, historical version state, and iteration.

### `DBConnection`

This interface represents a connection to a versioned key-value database. All versioning operations are performed using methods on this type.

* The `Versions` method returns a `VersionSet` which represents an immutable view of the version history at the current state.
* Version history is modified via the `{Save,Delete}Version` methods.
* Operations on version history do not modify any database contents.

### `DBReader`, `DBWriter`, and `DBReadWriter`

These types represent transactions on the database contents. Their methods provide CRUD operations as well as iteration.

* Writeable transactions call `Commit` flushes operations to the source DB.
* All open transactions must be closed with `Discard` or `Commit` before a new version can be saved on the source DB.
* The maximum number of safely concurrent transactions is dependent on the backend implementation.
* A single transaction object is not safe for concurrent use.
* Write conflicts on concurrent transactions will cause an error at commit time (optimistic concurrency control).

#### `Iterator`

* An iterator is invalidated by any writes within its `Domain` to the source transaction while it is open.
* An iterator must call `Close` before its source transaction is closed.

### `VersionSet`

This represents a self-contained and immutable view of a database's version history state. It is therefore safe to retain and conccurently access any instance of this object.

## Implementations

### In-memory DB

The in-memory DB in the `db/memdb` package cannot be persisted to disk. It is implemented using the Google [btree](https://pkg.go.dev/github.com/google/btree) library.

* This currently does not perform write conflict detection, so it only supports a single open write-transaction at a time. Multiple and concurrent read-transactions are supported.

### BadgerDB

A [BadgerDB](https://pkg.go.dev/github.com/dgraph-io/badger/v3)-based backend. Internally, this uses BadgerDB's ["managed" mode](https://pkg.go.dev/github.com/dgraph-io/badger/v3#OpenManaged) for version management.
Note that Badger only recognizes write conflicts for rows that are read _after_ a conflicting transaction was opened. In other words, the following will raise an error:

```go
tx1, tx2 := db.Writer(), db.ReadWriter()
key := []byte("key")
tx2.Get(key)
tx1.Set(key, []byte("a"))
tx2.Set(key, []byte("b"))
tx1.Commit()        // ok
err := tx2.Commit() // err is non-nil
```

But this will not:

```go
tx1, tx2 := db.Writer(), db.ReadWriter()
key := []byte("key")
tx1.Set(key, []byte("a"))
tx2.Set(key, []byte("b"))
tx1.Commit() // ok
tx2.Commit() // ok
```

### RocksDB

A [RocksDB](https://github.com/facebook/rocksdb)-based backend. Internally this uses [`OptimisticTransactionDB`](https://github.com/facebook/rocksdb/wiki/Transactions#optimistictransactiondb) to allow concurrent transactions with write conflict detection. Historical versioning is internally implemented with [Checkpoints](https://github.com/facebook/rocksdb/wiki/Checkpoints).
