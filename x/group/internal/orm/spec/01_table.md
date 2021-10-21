# Table

```go
type table struct {
	model       reflect.Type
	prefix      [2]byte
	afterSet    []AfterSetInterceptor
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

## AutoUInt64Table

`AutoUInt64Table` is a table type with an auto incrementing `uint64` ID.

```go
type AutoUInt64Table struct {
	*table
	seq Sequence
}
```

It's based on the `Sequence` struct which is a persistent unique key generator based on a counter encoded using 8 byte big endian.

## PrimaryKeyTable

`PrimaryKeyTable` provides simpler object style orm methods where are persisted and loaded with a reference to their unique primary key.

The model provided for creating a `PrimaryKeyTable` should implement the `PrimaryKeyed` interface:

```go
type PrimaryKeyed interface {
	// PrimaryKeyFields returns the fields of the object that will make up
	// the primary key. The PrimaryKey function will encode and concatenate
	// the fields to build the primary key.
	//
	// PrimaryKey parts can be []byte, string, and integer types. []byte is
	// encoded with a length prefix, strings are null-terminated, and
	// integers are encoded using 8 byte big endian.
	//
	// IMPORTANT: []byte parts are encoded with a single byte length prefix,
	// so cannot be longer than 255 bytes.
	PrimaryKeyFields() []interface{}
	codec.ProtoMarshaler
}
```

`PrimaryKeyFields()` method returns the list of key parts for a given object.
The primary key parts can be []byte, string, and `uint64` types. 
 Key parts, except the last part, follow these rules:
  - []byte is encoded with a single byte length prefix
  - strings are null-terminated
  - `uint64` are encoded using 8 byte big endian.
