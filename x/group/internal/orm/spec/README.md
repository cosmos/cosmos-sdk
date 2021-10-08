## Abstract

The orm package provides a framework for creating relational database tables with primary and secondary keys.

### Tables

```go
type table struct {
	model       reflect.Type
	prefix      [2]byte
	afterSet   []AfterSetInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}
```

A table can be built given a `codec.ProtoMarshaler` model type, a prefix to access the underlying prefix store used to store table data as well as a `Codec` for marshalling/unmarshalling.
In the prefix store, entities should be stored by an unique identifier called `RowID` which can be based either on an `uint64` auto-increment counter, string or dynamic size bytes.
Regular CRUD operations can be performed on a table, these methods take a `sdk.KVStore` as parameter to get the table prefix store.

The `table` struct does not:
 - enforce uniqueness of the `RowID`
 - enforce prefix uniqueness of keys, i.e. not allowing one key to be a prefix
 of another
 - optimize Gas usage conditions
The `table` struct is private, so that we only have custom tables built on top of it, that do satisfy these requirements.