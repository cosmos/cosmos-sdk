# Iterator and Pagination

Both [tables](01_table.md) and [secondary indexes](02_secondary_index.md) support iterating over a domain of keys, through `PrefixScan` or `ReversePrefixScan`, as well pagination.

## Iterator

An `Iterator` allows iteration through a sequence of key value pairs.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/types.go#L77-L85

Tables rely on a `typeSafeIterator` that is used by `PrefixScan` and `ReversePrefixScan` `table` methods to iterate through a range of `RowID`s.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/table.go#L285-L290

Secondary indexes rely on an `indexIterator` that can strip the `RowID` from the full index key in order to get the underlying value in the table prefix store.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/index.go#L232-L238

Under the hood, both use a prefix store `Iterator` (alias for tm-db `Iterator`).

## Pagination

The `Paginate` function does pagination given an [`Iterator`](#iterator) and a `query.PageRequest`, and returns a `query.PageResponse`.
It unmarshals the results into the provided dest interface that should be a pointer to a slice of models.

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-rc1/x/group/internal/orm/iterator.go#L102-L220

Secondary indexes have a `GetPaginated` method that returns an `Iterator` for the given searched secondary index key, starting from the `query.PageRequest` key if provided. It's important to note that this `query.PageRequest` key should be a `RowID` (that could have been returned by a previous paginated request). The returned `Iterator` can then be used with the `Paginate` function and the same `query.PageRequest`.
