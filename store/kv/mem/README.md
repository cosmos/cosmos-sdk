# Memory KVStore

The `mem.Store` implementation defines an in-memory `KVStore`, which is internally
backed by a thread-safe BTree. The `mem.Store` does not provide any branching
functionality and should be used as an ephemeral store, typically reset between
blocks. A `mem.Store` contains no reference to a parent store, but can be used
as a parent store for other stores. The `mem.Store` is can be useful for testing
purposes and where state persistence is not required or should be ephemeral.
