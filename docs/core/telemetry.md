<!--
order: 8
-->

# Telemetry

The Cosmos SDK enables operators and developers gain insight into the performance and behavior of
their application through the use of the `telemetry` package. The Cosmos SDK currently supports
enabling in-memory and prometheus as telemetry sinks. This allows the ability to query for and scrape
metrics from a single exposed API endpoint -- `/metrics?format={text|prometheus}`, the default being
`text`.

If telemetry is enabled via configuration, a single global metrics collector is registered via the
[go-metrics](https://github.com/armon/go-metrics) library. This allows emitting and collecting
metrics through simple API calls.

Example:

```go
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
  defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.MetricKeyEndBlocker)

  // ...
}
```

Developers may use the `telemetry` package directly, which provides wrappers around metric APIs
that include adding useful labels, or they must the `go-metrics` library directly. It is preferable
to add as much context and adequate dimensionality to metrics as possible, so the `telemetry` package
is advised. Regardless of the package or method used, the Cosmos SDK supports the following metrics
types:

* gauges
* summaries
* counters

## Cardinality

Cardinality is key, specifically label and key cardinality. Cardinality is how many unique values of
something there are. So there is naturally a tradeoff between granularity and how much stress is put
on the telemetry sink in terms of indexing, scrape, and query performance.

Developers should take care to support metrics with enough dimensionality and granularity to be
useful, but not increase the cardinality beyond the sink's limits. A general rule of thumb is to not
exceed a cardinality of 10.

Consider the following examples with enough granularity and adequate cardinality:

* begin/end blocker time
* tx gas used
* block gas used
* amount of tokens minted

The following examples expose too much cardinality and may not even prove to be useful:

* transfers between accounts with amount
* voting/deposit amount from unique addresses
