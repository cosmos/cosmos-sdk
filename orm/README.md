# ORM

The Cosmos SDK ORM is a state management library that provides a rich, but opinionated set of tools for managing a
module's state. It provides support for:
* type safe management of state
* multipart keys
* secondary indexes
* unique indexes
* easy prefix and range queries
* automatic genesis import/export
* automatic query services for clients, including support for light client proof
* indexing state data in external databases

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

We recommend that users of the ORM attempt to follow database design best practices such as [normalization](https://en.wikipedia.org/wiki/Database_normalization).
For instance, defining `repeated` fields in a table is considered an anti-pattern because breaks first normal form (1NF).
Although we support `repeated` fields in tables, they cannot be used as key fields for this reason. This may seem
restrictive but years of best practice (and also experience in the SDK) have shown that following this pattern
leads to easier to maintain schemas.

To illustrate the motivation for these principles with an example from the SDK, historically balances were stored
as a mapping from account -> map of denom to amount. This did not scale well because an account with 100 token balances
but need to be encoded/decoded every time a single coin balance changes. Now balances are stored as account,denom -> amount
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

## Auto-incrementing Primary Keys

A common pattern in SDK modules and in database design is to define tables with a single integer `id` field with an
automatically generated primary key. In the ORM we can do this by setting the `auto_increment` option to `true` on the
primary key, ex:
```protobuf
message Account {
  option (cosmos.orm.v1.table) = {
    id: 1;
    primary_key: { fields: "id", auto_increment: true }
  };

  uint64 id = 1;
  bytes address = 2;
}
```

## Unique Indexes

A unique index can be added by setting the `unique` option to `true` on an index, ex:
```protobuf
message Account {
  option (cosmos.orm.v1.table) = {
    id: 1;
    primary_key: { fields: "id", auto_increment: true }
    index: {id: 1, fields: "address", unique: true}
  };

  uint64 id = 1;
  bytes address = 2;
}
```
