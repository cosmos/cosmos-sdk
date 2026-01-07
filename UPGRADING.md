# Upgrade Reference

This document provides a quick reference for the upgrades from `v0.53.x` to `v0.54.x` of Cosmos SDK.

Note, always read the **App Wiring Changes** section for more information on application wiring updates.

## TLDR

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/CHANGELOG.md).

## x/gov

### Keeper Initialization

The `x/gov` module has been decoupled from `x/staking`. The `keeper.NewKeeper` constructor now requires a `CalculateVoteResultsAndVotingPowerFn` parameter instead of a `StakingKeeper`.

**Before:**
```go
govKeeper := keeper.NewKeeper(
    cdc,
    storeService,
    authKeeper,
    bankKeeper,
    stakingKeeper,  // StakingKeeper parameter
    distrKeeper,
    router,
    config,
    authority,
)
```

**After:**
```go
govKeeper := keeper.NewKeeper(
    cdc,
    storeService,
    authKeeper,
    bankKeeper,
    keeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper),  // Function parameter
    distrKeeper,
    router,
    config,
    authority,
)
```

For applications using depinject, the governance module now accepts an optional `CalculateVoteResultsAndVotingPowerFn`. If not provided, it will use the `StakingKeeper` (also optional) to create the default function.

### GovHooks Interface

The `AfterProposalSubmission` hook now includes the proposer address as a parameter.

**Before:**
```go
func (h MyGovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64) error {
    // implementation
}
```

**After:**
```go
func (h MyGovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64, proposerAddr sdk.AccAddress) error {
    // implementation
}
```

## Adoption of OpenTelemetry and Deprecation of `github.com/hashicorp/go-metrics`

Existing Cosmos SDK telemetry support is provided by `github.com/hashicorp/go-metrics` which is undermaintained and only supported metrics instrumentation.
OpenTelemetry provides an integrated solution for metrics, traces, and logging which is widely adopted and actively maintained.
The existing wrapper functions in the `telemetry` package required acquiring mutex locks and map lookups for every metric operation which is sub-optimal. OpenTelemetry's API uses atomic concurrency wherever possible and should introduce less performance overhead during metric collection.

The [README.md](telemetry/README.md) in the `telemetry` package provides more details on usage, but below is a quick summary:
1. application developers should follow the official [go OpenTelemetry](https://pkg.go.dev/go.opentelemetry.io/otel) guidelines when instrumenting their applications.
2. node operators who want to configure OpenTelemetry exporters should set the `OTEL_EXPERIMENTAL_CONFIG_FILE` environment variable to the path of a yaml file which follows the OpenTelemetry declarative configuration format specified here: https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf. As long as the `telemetry` package has been imported somewhere (it should already be imported if you are using the SDK), OpenTelemetry will be initialized automatically based on the configuration file.

NOTE: the go implementation of [otelconf](https://pkg.go.dev/go.opentelemetry.io/contrib/otelconf) is still under development and we will update our usage of it as it matures.

## Log v2

The log package has been refactored to support OpenTelemetry tracing correlation. If you have basic usage of the logger, such as getting it from sdk.Context and logging to the console; nothing changes. You are unaffected.

If you implement your own logger, or setup the logger manually, a few things have been changed.


### Logger Interface

The interface has 1 new method:

```go
type Logger interface {
	Ctx(context.Context) Logger
}
```

This method attaches a context to the logger, allowing logs to be correlated with OpenTelemetry traces.
Note that if OpenTelemetry is not used in your application, ZeroLogger will be used, and the Ctx method will be a noop.

### Constructor

If you construct your own logger, the constructor function arguments and option names have changed.

Old constructor: 
```go
func NewLogger(dst io.Writer, options ...Option) Logger
```

New constructor:
```go
func NewLogger(name string, opts ...Option) Logger
```

### Options

All option functions are now prefixed with `With` to make option discovery easier and more clear.

Example:

Before:
```go
func FilterOption(filter FilterFunc) Option
```

After:
```go
func WithFilter(filter FilterFunc) Option
```

Additionally, new options have been added. See log/options.go for all usable options.

### Default Behavior

By default, the logger will log to the console AND to OpenTelemetry if a logger provider has been set in the optional OpenTelemetry configuration.

For production nodes, we have introduced a new flag, `--log_no_console` to disable the overhead of logging to the console. This is useful in situations where you ONLY want OpenTelemetry log forwarding.

### Log Levels

Levels have been changed to achieve parity with slog. The supported levels are:

- info
- debug
- warn
- error

The following levels are no longer supported:

- trace
- fatal
- panic