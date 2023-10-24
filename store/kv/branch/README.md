# Branch KVStore

The `branch.Store` implementation defines a `BranchedKVStore` that contains a
reference to a `VersionedDatabase`, i.e. an SS backend. The `branch.Store` is
meant to be used as the primary store used in a `RootStore` implementation. It
provides the ability to get the current `ChangeSet`, branching, and writing to
a parent store (if one is defined). Note, all reads first pass through the
staged, i.e. dirty writes. If a key is not found in the staged writes, the read
is then passed to the parent store (if one is defined), finally falling back to
the backing SS engine.
