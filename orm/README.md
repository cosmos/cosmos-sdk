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
with a primary key and optional secondary indexes. Because the Cosmos SDK uses protobuf as its encoding layer, ORM
tables are defined directly in .proto files using protobuf options. Each table is defined by a single protobuf `message`
type and a schema of multiple tables is represented by a single .proto file. Table structure is specified in the same
file where messages are defined in order to make it easy to focus on better design of the state layer. Because
blockchain state layout is part of the public API for clients (TODO: link to docs on light client proofs), it is
important to think about the state layout as being part of the public API of a module. Changing the state layout
actually breaks clients, so it is ideal to think through it carefully up front and to aim for a design that will
eliminate or minimize breaking changes down the road. Also, good design of state enables building more performant
and sophisticated applications. Providing users with a set of tools inspired by relational databases which have a
long history of database design best practices and allowing schema to be specified declaratively in a single place
are design choices the ORM makes to enable better design and more durable APIs. Also, by only supporting the table
abstraction as opposed to key-value pair maps, it is easy to add to new columns/fields to any data structure without
causing a breaking change and the data structures can easily be indexed in any off-the-shelf SQL database for more
sophisticated queries.

## Defining Tables



