# State Storage (SS)

The `storage` package contains the state storage (SS) implementation. Specifically,
it contains RocksDB, PebbleDB, and SQLite (Btree) backend implementations of the
`VersionedDatabase` interface.

The goal of SS is to provide a modular storage backend, i.e. multiple implementations,
to facilitate storing versioned raw key/value pairs in a fast embedded database,
although an embedded database is not required, i.e. you could use a replicated
RDBMS system.

The responsibility and functions of SS include the following:

* Provided fast and efficient queries for versioned raw key/value pairs
* Provide versioned CRUD operations
* Provide versioned batching functionality
* Provide versioned iteration (forward and reverse) functionality
* Provide pruning functionality

All of the functionality provided by an SS backend should work under a versioned
scheme, i.e. a user should be able to get, store, and iterate over keys for the
latest and historical versions efficiently.

## Backends

### RocksDB

The RocksDB implementation is a CGO-based SS implementation. It fully supports
the `VersionedDatabase` API and is arguably the most efficient implementation. It
also supports versioning out-of-the-box using User-defined Timestamps in
ColumnFamilies (CF). However, it requires the CGO dependency which can complicate
an app’s build process.

### PebbleDB

The PebbleDB implementation is a native Go SS implementation that is primarily an
alternative to RocksDB. Since it does not support CF, results in the fact that we
need to implement versioning (MVCC) ourselves. This comes with added implementation
complexity and potential performance overhead. However, it is a pure Go implementation
and does not require CGO.

### SQLite (Btree)

The SQLite implementation is a native Go SS implementation. It fully supports the
`VersionedDatabase` API and the implementation is relatively straightforward and
easy to understand as it’s SQL-based.

## Benchmarks

Benchmarks for basic operations on all supported native SS implementations can
be found in `store/storage/storage_bench_test.go`.

At the time of writing, the following benchmarks were performed:

```shell
name                                              time/op
Get/backend_rocksdb_versiondb_opts-10             7.41µs ± 0%
Get/backend_pebbledb_default_opts-10              6.17µs ± 0%
Get/backend_btree_sqlite-10                       29.1µs ± 0%
ApplyChangeset/backend_pebbledb_default_opts-10   5.73ms ± 0%
ApplyChangeset/backend_btree_sqlite-10            56.9ms ± 0%
ApplyChangeset/backend_rocksdb_versiondb_opts-10  4.07ms ± 0%
Iterate/backend_pebbledb_default_opts-10           1.04s ± 0%
Iterate/backend_btree_sqlite-10                    1.59s ± 0%
Iterate/backend_rocksdb_versiondb_opts-10          778ms ± 0%
```

## Usage

An SS backend is meant to be used within a broader store implementation, as it
only stores data for direct and historical query purposes. We define a `Database`
interface in the `storage` package which is mean to be represent a `VersionedDatabase`
with only the necessary methods. The `StorageStore` interface is meant to wrap or
accept this `Database` type, e.g. RocksDB.

The `StorageStore` interface is an abstraction or wrapper around the backing SS
engine can be seen as the the main entry point to using SS.

Higher up the stack, there should exist a `RootStore` implementation. The `RootStore`
is meant to encapsulate both an SS backend and an SC backend. The SS backend is
defined by this `StorageStore` implementation.
