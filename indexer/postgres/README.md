# PostgreSQL Indexer

The PostgreSQL indexer can fully index the current state for all modules that implement `cosmossdk.io/schema.HasModuleCodec`.
implement `cosmossdk.io/schema.HasModuleCodec`.

## Table, Column and Enum Naming

`ObjectType`s names are converted to table names prefixed with the module name and an underscore. i.e. the `ObjectType` `foo` in module `bar` will be stored in a table named `bar_foo`.

Column names are identical to field names. All identifiers are quoted with double quotes so that they are case-sensitive and won't clash with any reserved names. 

Like, table names, enum types are prefixed with the module name and an underscore.

## Schema Type Mapping

The mapping of `cosmossdk.io/schema` `Kind`s to PostgreSQL types is as follows:

| Kind | PostgreSQL Type            | Notes                                                                                                                                                                           |
|---------------------|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `StringKind`        | `TEXT`                     |                                                                                                                                                                                 |
| `BoolKind`          | `BOOLEAN`                  |                                                                                                                                                                                 |
| `BytesKind`         | `BYTEA`                    |                                                                                                                                                                                 |
| `Int8Kind`          | `SMALLINT`                 |                                                                                                                                                                                 |
| `Int16Kind`         | `SMALLINT`                 |                                                                                                                                                                                 |
| `Int32Kind`         | `INTEGER`                  |                                                                                                                                                                                 |
| `Int64Kind`         | `BIGINT`                   |                                                                                                                                                                                 |
| `Uint8Kind`         | `SMALLINT`                 |                                                                                                                                                                                 |
| `Uint16Kind`        | `INTEGER`                  |                                                                                                                                                                                 |
| `Uint32Kind`        | `BIGINT`                   |                                                                                                                                                                                 |
| `Uint64Kind`        | `NUMERIC`                  |                                                                                                                                                                                 |
| `Float32Kind`       | `REAL`                     |                                                                                                                                                                                 |
| `Float64Kind`       | `DOUBLE PRECISION`         |                                                                                                                                                                                 |
| `IntegerStringKind` | `NUMERIC`                  |                                                                                                                                                                                 |
| `DecimalStringKind` | `NUMERIC`                  |                                                                                                                                                                                 |
| `JSONKind`          | `JSONB`                    |                                                                                                                                                                                 |
| `Bech32AddressKind` | `TEXT`                     | addresses are converted to strings with the specified address prefix                                                                                                            |
| `TimeKind`          | `BIGINT` and `TIMESTAMPTZ` | time types are stored as two columns, one with the `_nanos` suffix with full nanoseconds precision, and another as a `TIMESTAMPTZ` generated column with microsecond precision |
| `DurationKind`      | `BIGINT`                   | durations are stored as a single column in nanoseconds                                                                                                                          |
| `EnumKind` | `<module_name>_<enum_name>` | a custom enum type is created for each module prefixed with the module name it pertains to                                                                                     |


