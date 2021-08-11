## Abstract

The orm package provides a framework for creating relational database tables with primary and secondary keys.

```go
type Table struct {
	model       reflect.Type
	prefix      byte
	afterSave   []AfterSaveInterceptor
	afterDelete []AfterDeleteInterceptor
	cdc         codec.Codec
}
```

Such table can be built given a `codec.ProtoMarshaler` model type, a prefix to access the underlying prefix store used to store table data as well as a `Codec` for marshalling/unmarshalling.
In the prefix store, entities are stored by an unique identifier called `RowID` which can be based either on an `uint64` auto-increment counter or dynamic size bytes.
Regular CRUD operations can be performed on a table, these methods take a `sdk.KVStore` as parameter to get the table prefix store.