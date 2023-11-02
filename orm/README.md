# ORM

The Cosmos SDK ORM is a state management library that provides a rich, but opinionated set of tools for managing a
module's state. It provides support for:

* type safe management of state
* multipart keys
* secondary indexes
* unique indexes
* easy prefix and range queries
* automatic genesis import/export
* automatic query services for clients, including support for light client proofs (still in development)
* indexing state data in external databases (still in development)

## Design and Philosophy

The ORM's data model is inspired by the relational data model found in SQL databases. The core abstraction is a table
with a primary key and optional secondary indexes.

Because the Cosmos SDK uses protobuf as its encoding layer, ORM tables are defined directly in .proto files using
protobuf options. Each table is defined by a single protobuf `message` type and a schema of multiple tables is
represented by a single .proto file.

Table structure is specified in the same file where messages are defined in order to make it easy to focus on better
design of the state layer. Because blockchain state layout is part of the public API for clients (TODO: link to docs on
light client proofs), it is important to think about the state layout as being part of the public API of a module.
Changing the state layout actually breaks clients, so it is ideal to think through it carefully up front and to aim for
a design that will eliminate or minimize breaking changes down the road. Also, good design of state enables building
more performant and sophisticated applications. Providing users with a set of tools inspired by relational databases
which have a long history of database design best practices and allowing schema to be specified declaratively in a
single place are design choices the ORM makes to enable better design and more durable APIs.

Also, by only supporting the table abstraction as opposed to key-value pair maps, it is easy to add to new
columns/fields to any data structure without causing a breaking change and the data structures can easily be indexed in
any off-the-shelf SQL database for more sophisticated queries.

The encoding of fields in keys is designed to support ordered iteration for all protobuf primitive field types
except for `bytes` as well as the well-known types `google.protobuf.Timestamp` and `google.protobuf.Duration`. Encodings
are optimized for storage space when it makes sense (see the documentation in `cosmos/orm/v1/orm.proto` for more details)
and table rows do not use extra storage space to store key fields in the value.

We recommend that users of the ORM attempt to follow database design best practices such as
[normalization](https://en.wikipedia.org/wiki/Database_normalization) (at least 1NF).
For instance, defining `repeated` fields in a table is considered an anti-pattern because breaks first normal form (1NF).
Although we support `repeated` fields in tables, they cannot be used as key fields for this reason. This may seem
restrictive but years of best practice (and also experience in the SDK) have shown that following this pattern
leads to easier to maintain schemas.

To illustrate the motivation for these principles with an example from the SDK, historically balances were stored
as a mapping from account -> map of denom to amount. This did not scale well because an account with 100 token balances
needed to be encoded/decoded every time a single coin balance changed. Now balances are stored as account,denom -> amount
as in the example above. With the ORM's data model, if we wanted to add a new field to `Balance` such as
`unlocked_balance` (if vesting accounts were redesigned in this way), it would be easy to add it to this table without
requiring a data migration. Because of the ORM's optimizations, the account and denom are only stored in the key part
of storage and not in the value leading to both a flexible data model and efficient usage of storage.

## Defining Tables

To define a table:

1) create a .proto file to describe the module's state (naming it `state.proto` is recommended for consistency),
and import "cosmos/orm/v1/orm.proto", ex:

```protobuf
syntax = "proto3";
package bank_example;

import "cosmos/orm/v1/orm.proto";
```

2) define a `message` for the table, ex:

```protobuf
message Balance {
  bytes account = 1;
  string denom = 2;
  uint64 balance = 3;
}
```

3) add the `cosmos.orm.v1.table` option to the table and give the table an `id` unique within this .proto file:

```protobuf
message Balance {
  option (cosmos.orm.v1.table) = {
    id: 1
  };
  
  bytes account = 1;
  string denom = 2;
  uint64 balance = 3;
}
```

4) define the primary key field or fields, as a comma-separated list of the fields from the message which should make
up the primary key:

```protobuf
message Balance {
  option (cosmos.orm.v1.table) = {
    id: 1
    primary_key: { fields: "account,denom" }
  };

  bytes account = 1;
  string denom = 2;
  uint64 balance = 3;
}
```

5) add any desired secondary indexes by specifying an `id` unique within the table and a comma-separate list of the
index fields:

```protobuf
message Balance {
  option (cosmos.orm.v1.table) = {
    id: 1;
    primary_key: { fields: "account,denom" }
    index: { id: 1 fields: "denom" } // this allows querying for the accounts which own a denom
  };

  bytes account = 1;
  string denom   = 2;
  uint64 amount  = 3;
}
```

### Auto-incrementing Primary Keys

A common pattern in SDK modules and in database design is to define tables with a single integer `id` field with an
automatically generated primary key. In the ORM we can do this by setting the `auto_increment` option to `true` on the
primary key, ex:

```protobuf
message Account {
  option (cosmos.orm.v1.table) = {
    id: 2;
    primary_key: { fields: "id", auto_increment: true }
  };

  uint64 id = 1;
  bytes address = 2;
}
```

### Unique Indexes

A unique index can be added by setting the `unique` option to `true` on an index, ex:

```protobuf
message Account {
  option (cosmos.orm.v1.table) = {
    id: 2;
    primary_key: { fields: "id", auto_increment: true }
    index: {id: 1, fields: "address", unique: true}
  };

  uint64 id = 1;
  bytes address = 2;
}
```

### Singletons

The ORM also supports a special type of table with only one row called a `singleton`. This can be used for storing
module parameters. Singletons only need to define a unique `id` and that cannot conflict with the id of other
tables or singletons in the same .proto file. Ex:

```protobuf
message Params {
  option (cosmos.orm.v1.singleton) = {
    id: 3;
  };
  
  google.protobuf.Duration voting_period = 1;
  uint64 min_threshold = 2;
}
```

## Running Codegen

NOTE: the ORM will only work with protobuf code that implements the [google.golang.org/protobuf](https://pkg.go.dev/google.golang.org/protobuf)
API. That means it will not work with code generated using gogo-proto.

To install the ORM's code generator, run:

```shell
go install cosmossdk.io/orm/cmd/protoc-gen-go-cosmos-orm@latest
```

The recommended way to run the code generator is to use [buf build](https://docs.buf.build/build/usage).
This is an example `buf.gen.yaml` that runs `protoc-gen-go`, `protoc-gen-go-grpc` and `protoc-gen-go-cosmos-orm`
using buf managed mode:

```yaml
version: v1
managed:
  enabled: true
  go_package_prefix:
    default: foo.bar/api # the go package prefix of your package
    override:
      buf.build/cosmos/cosmos-sdk: cosmossdk.io/api # required to import the Cosmos SDK api module
plugins:
  - name: go
    out: .
    opt: paths=source_relative
  - name: go-grpc
    out: .
    opt: paths=source_relative
  - name: go-cosmos-orm
    out: .
    opt: paths=source_relative
```

## Using the ORM in a module

### Initialization

To use the ORM in a module, first create a `ModuleSchemaDescriptor`. This tells the ORM which .proto files have defined
an ORM schema and assigns them all a unique non-zero id. Ex:

```go
var MyModuleSchema = &ormv1alpha1.ModuleSchemaDescriptor{
    SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
        {
            Id:            1,
            ProtoFileName: mymodule.File_my_module_state_proto.Path(),
        },
    },
}
```

In the ORM generated code for a file named `state.proto`, there should be an interface `StateStore` that got generated
with a constructor `NewStateStore` that takes a parameter of type `ormdb.ModuleDB`. Add a reference to `StateStore`
to your module's keeper struct. Ex:

```go
type Keeper struct {
    db StateStore
}
```

Then instantiate the `StateStore` instance via an `ormdb.ModuleDB` that is instantiated from the `SchemaDescriptor`
above and one or more store services from `cosmossdk.io/core/store`. Ex:

```go
func NewKeeper(storeService store.KVStoreService) (*Keeper, error) {
    modDb, err := ormdb.NewModuleDB(MyModuleSchema, ormdb.ModuleDBOptions{KVStoreService: storeService})
    if err != nil {
        return nil, err
    }
    db, err := NewStateStore(modDb)
    if err != nil {
        return nil, err
    }
    return Keeper{db: db}, nil
}
```

### Using the generated code

The generated code for the ORM contains methods for inserting, updating, deleting and querying table entries.
For each table in a .proto file, there is a type-safe table interface implemented in generated code. For instance,
for a table named `Balance` there should be a `BalanceTable` interface that looks like this:

```go
type BalanceTable interface {
    Insert(ctx context.Context, balance *Balance) error
    Update(ctx context.Context, balance *Balance) error
    Save(ctx context.Context, balance *Balance) error
    Delete(ctx context.Context, balance *Balance) error
    Has(ctx context.Context, acocunt []byte, denom string) (found bool, err error)
    // Get returns nil and an error which responds true to ormerrors.IsNotFound() if the record was not found.
    Get(ctx context.Context, acocunt []byte, denom string) (*Balance, error)
    List(ctx context.Context, prefixKey BalanceIndexKey, opts ...ormlist.Option) (BalanceIterator, error)
    ListRange(ctx context.Context, from, to BalanceIndexKey, opts ...ormlist.Option) (BalanceIterator, error)
    DeleteBy(ctx context.Context, prefixKey BalanceIndexKey) error
    DeleteRange(ctx context.Context, from, to BalanceIndexKey) error

    doNotImplement()
}
```

This `BalanceTable` should be accessible from the `StateStore` interface (assuming our file is named `state.proto`)
via a `BalanceTable()` accessor method. If all the above example tables/singletons were in the same `state.proto`,
then `StateStore` would get generated like this:

```go
type BankStore interface {
    BalanceTable() BalanceTable
    AccountTable() AccountTable
    ParamsTable() ParamsTable

    doNotImplement()
}
```

So to work with the `BalanceTable` in a keeper method we could use code like this:

```go
func (k keeper) AddBalance(ctx context.Context, acct []byte, denom string, amount uint64) error {
    balance, err := k.db.BalanceTable().Get(ctx, acct, denom)
    if err != nil && !ormerrors.IsNotFound(err) {
        return err
    }

    if balance == nil {
        balance = &Balance{
            Account: acct,
            Denom:   denom,
            Amount:  amount,
        }
    } else {
        balance.Amount = balance.Amount + amount
    }

    return k.db.BalanceTable().Save(ctx, balance)
}
```

`List` methods take `IndexKey` parameters. For instance, `BalanceTable.List` takes `BalanceIndexKey`. `BalanceIndexKey`
let's represent index keys for the different indexes (primary and secondary) on the `Balance` table. The primary key
in the `Balance` table gets a struct `BalanceAccountDenomIndexKey` and the first index gets an index key `BalanceDenomIndexKey`.
If we wanted to list all the denoms and amounts that an account holds, we would use `BalanceAccountDenomIndexKey`
with a `List` query just on the account prefix. Ex:

```go
it, err := keeper.db.BalanceTable().List(ctx, BalanceAccountDenomIndexKey{}.WithAccount(acct))
```
