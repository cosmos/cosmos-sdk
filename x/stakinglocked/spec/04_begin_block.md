<!--
order: 4
-->

# Begin-Block

Each abci begin block call, the historical info will get stored and pruned
according to the `HistoricalEntries` parameter.

## Historical Info Tracking

If the `HistoricalEntries` parameter is 0, then the `BeginBlock` performs a no-op.

Otherwise, the latest historical info is stored under the key `historicalInfoKey|height`, while any entries older than `height - HistoricalEntries` is deleted.
In most cases, this results in a single entry being pruned per block.
However, if the parameter `HistoricalEntries` has changed to a lower value there will be multiple entries in the store that must be pruned.
