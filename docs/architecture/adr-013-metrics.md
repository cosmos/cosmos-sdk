# ADR 013: Metrics

## Changelog

- 14-10-2019: Initial Draft

## Status

Proposed

## Context

### A

To bring more visibility into applications, I would like to propose reporting of metrics and possibly tracing support. There are a few solutions to be considered:

1. [Prometheus](https://prometheus.io/docs/introduction/overview/)
2. [Open Telemetry](https://opentelemetry.io/)
   - Open telemetry is still not stable, and if the decision is to use it then the implementation should be postponed.
3. [OpenCensus](https://opencensus.io/)

#### 1. Prometheus

Prometheus is the most popular at this current time and is being maintained by a few companies. It has a Go client library, powerful queries and alerts.

a) Prometheus API

We can commit to using Prometheus, but there has already been some talks that some people would like to use other forms of metrics and not be tied down to prometheus.

b) [go-kit metrics package](https://godoc.org/github.com/go-kit/kit/metrics#pkg-subdirectories) as an interface

metrics package provides a set of uniform interfaces for service instrumentation and offers adapters to popular metrics packages:

https://godoc.org/github.com/go-kit/kit/metrics#pkg-subdirectories

Comparing to Prometheus API, we're losing customizability and control, but gaining freedom in choosing any instrument from the above list given we will extract metrics creation into a separate function.

#### 2. Open Telemetry

Open telemetry is a new tool that is still under active development, it provides metrics and tracing. The api has stabilized but has not been fully implemented and they will not support backwards compatibility. Open Telemetry also has a [Go client library](https://github.com/open-telemetry/opentelemetry-go).

a) [Open Telemetry API Specification](https://github.com/open-telemetry/opentelemetry-specification#table-of-contents)

We can commit to using the API and accept the breaking changes when they come in the future.

b) Wait to see if integration into a library like [go-kit metrics](https://godoc.org/github.com/go-kit/kit/metrics#pkg-subdirectories) will be made. This would allow the user to decide what metric library they would like to use.

#### 3. OpenCensus

OpenCensus and Open tracing are being merged into Open telemetry.

#### List Of Metrics

(TBD)

| Name                  |   type    | Description |
| --------------------- | :-------: | ----------: |
| iavl_io               |  Gauage   |             |
| db_io                 |  Gauage   |             |
| txs                   | histogram |             |
| delegations (staking) | histogram |             |
| Rewards (distr)       | histogram |             |
| Sends (bank)          | histogram |             |
| Supply (supply)       |  Counter  |             |

### B

Part B of this ADR is geared more towards modules. Extending `AppModuleBasic` to support registering of metrics could be the best path forward.

```go
type AppModuleBasic interface {
	Name() string
  RegisterCodec(*codec.Codec)
  + RegisterMetrics(namespace string, labelsAndValues ...string) *Metrics +

	// genesis
	DefaultGenesis() json.RawMessage
	ValidateGenesis(json.RawMessage) error

	// client functionality
	RegisterRESTRoutes(context.CLIContext, *mux.Router)
	GetTxCmd(*codec.Codec) *cobra.Command
	GetQueryCmd(*codec.Codec) *cobra.Command
}

```

## Decision

## Consequences

### Positive

- Add more visibility into SDK based application and modules

### Negative

### Neutral

## References
