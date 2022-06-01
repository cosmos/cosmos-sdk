# Table

A table can be built given a `codec.ProtoMarshaler` model type, a prefix to access the underlying prefix store used to store table data as well as a `Codec` for marshalling/unmarshalling.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/table.go#L30-L36

In the prefix store, entities should be stored by an unique identifier called `RowID` which can be based either on an `uint64` auto-increment counter, string or dynamic size bytes.
Regular CRUD operations can be performed on a table, these methods take a `sdk.KVStore` as parameter to get the table prefix store.

The `table` struct does not:

* enforce uniqueness of the `RowID`
* enforce prefix uniqueness of keys, i.e. not allowing one key to be a prefix of another
* optimize Gas usage conditions

The `table` struct is private, so that we only have custom tables built on top of it, that do satisfy these requirements.

`table` provides methods for exporting (using a [`PrefixScan` `Iterator`](03_iterator_pagination.md#iterator)) and importing genesis data. For the import to be successful, objects have to be aware of their primary key by implementing the [`PrimaryKeyed`](#primarykeyed) interface.

## AutoUInt64Table

`AutoUInt64Table` is a table type with an auto incrementing `uint64` ID.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/auto_uint64.go#L15-L18

It's based on the `Sequence` struct which is a persistent unique key generator based on a counter encoded using 8 byte big endian.

## PrimaryKeyTable

`PrimaryKeyTable` provides simpler object style orm methods where are persisted and loaded with a reference to their unique primary key.

### PrimaryKeyed

The model provided for creating a `PrimaryKeyTable` should implement the `PrimaryKeyed` interface:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/primary_key.go#L30-L44

`PrimaryKeyFields()` method returns the list of key parts for a given object.
The primary key parts can be []byte, string, and `uint64` types.

### Key codec

Key parts, except the last part, follow these rules:

* []byte is encoded with a single byte length prefix (which means the max []byte length is 255)
* strings are null-terminated
* `uint64` are encoded using 8 byte big endian.
