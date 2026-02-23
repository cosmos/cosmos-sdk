# CacheKVStore

The cache key-value store is a wrapper around a `KVStore` that buffers all writes to the parent store,
so that they can be dropped if they need to be reverted.

## No Caching

Previous iterations of this code did complex caching of KV-store reads.
However, this is completely unnecessary because caching of reads against the underlying KV-store
should be a first order concern of the underlying KV-store implementation, not
delegated to this wrapper which might actually introduce multiple nested cache layers during execution.

## No Mutex

Also, the previous implementation had a mutex lock which prevented any concurrent reads.
In this implementation, there is no lock.
Concurrent reads should be totally safe as long as there are no concurrent writes, and as long
as the underlying KV-store implementation is safe for concurrent reads.
If it isn't safe to do concurrent reads against the underlying store, that store should have a mutex.
It is and has never been safe to do concurrent writes against any state store in this project,
and no code path should ever do that (there would be other problems like non-determinism if it did).