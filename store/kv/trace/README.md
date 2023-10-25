# Trace KVStore

The `trace.Store` implementation defines a store which wraps a parent `KVStore`
and traces all operations performed on it. Each trace operation is written to a
provided `io.Writer` object. Specifically, a `TraceOperation` object is JSON
encoded and written to the writer. The `TraceOperation` object contains the exact
operation, e.g. a read or write, and the corresponding key and value pair.

A `trace.Store` can also be instantiated with a `store.TraceContext` which
can allow each traced operation to include additional metadata, e.g. a block height
or hash.

Note, `trace.Store` is not meant to be branched or written to. The parent `KVStore`
is responsible for all branching and writing operations, while a `trace.Store`
wraps such a store and traces all relevant operations on it.
