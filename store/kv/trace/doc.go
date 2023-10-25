/*
Package trace provides a KVStore implementation that wraps a parent KVStore
and allows all operations to be traced to an io.Writer. This can be useful to
serve use cases such as tracing and digesting all read operations for a specific
store key and key or value.
*/
package trace
