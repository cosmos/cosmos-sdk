# ADR 013: Observability

## Changelog

- 20-01-2020: Initial Draft

## Status

Proposed

## Context

### A

There has been discussion around exposing more metrics to users and node operators about the application. Currently there is only a way to expose metrics from Tendermint and not the application itself.To bring more visibility into applications, I would like to propose reporting of metrics through [Prometheus](https://prometheus.io/).

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

Part B of this ADR is geared more towards modules. Extending `AppModuleBasic` to support registering of metrics would enable developers to see more information about individual modules.

```go
type AppModuleBasic interface {
  Name() string
  RegisterCodec(*codec.Codec)
  + RegisterMetrics(moduleName string, labelsAndValues ...string) *Metrics

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

- Use Prometheus for metric gathering.
- Add a method to register metrics to the `AppModuleBasic` interface

## Consequences

### Positive

- Add more visibility into SDK based application and modules

### Negative

### Neutral

## References
