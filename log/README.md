# Log v2

`cosmossdk.io/log/v2` is the Cosmos SDK logging package.

At a high level, there are three pieces to understand:

1. `log.NewLogger(...)` creates the default Cosmos SDK logger. It is backed by `zerolog`.
2. `cosmossdk.io/log/v2/slog` lets you satisfy the same SDK `Logger` interface with a standard library `*slog.Logger`.
3. `log.NewMultiLogger(...)` fans one log call out to multiple SDK loggers. The SDK uses this during server startup when OpenTelemetry log exporting is enabled. To learn more about how we support OpenTelemetry, read the [Telemetry docs](../telemetry/README.md).

If you only need ordinary SDK logging, you usually only need `log.NewLogger`, which is automatically provisioned and set on `sdk.Context`.

## Default Logger

The default implementation is a small wrapper around `zerolog`.

```go
logger := log.NewLogger(os.Stderr)

logger.Info("starting app", "chain_id", chainID)
logger.Error("failed to load state", "err", err)
```

`NewLogger` writes human-readable console output by default. The server command wiring switches options based on CLI configuration, for example:

- `OutputJSONOption()` for JSON logs
- `LevelOption(...)` for a global log level
- `FilterOption(...)` for module-based filtering
- `TraceOption(true)` to include stack traces on error logs
- `VerboseLevelOption(...)` for temporary verbose mode

The SDK also uses the `module` field consistently. The package exposes `log.ModuleKey` for this:

```go
logger = logger.With(log.ModuleKey, "bank")
logger.Info("send coins", "from", from, "to", to)
```

That matters because the log filter implementation keys off the `module` field when parsing values such as `consensus:debug,*:error`.

## Structured Context

`Logger.With(...)` returns a derived logger with additional fields:

```go
keeperLogger := logger.With(log.ModuleKey, "staking", "component", "keeper")
keeperLogger.Info("validator updated", "operator", valAddr)
```

This is the normal way to attach stable metadata to a logger instance.

## Context-Aware Logging

The v2 `Logger` interface adds `*Context` methods:

```go
type Logger interface {
	Info(msg string, keyVals ...any)
	InfoContext(ctx context.Context, msg string, keyVals ...any)
	Warn(msg string, keyVals ...any)
	WarnContext(ctx context.Context, msg string, keyVals ...any)
	Error(msg string, keyVals ...any)
	ErrorContext(ctx context.Context, msg string, keyVals ...any)
	Debug(msg string, keyVals ...any)
	DebugContext(ctx context.Context, msg string, keyVals ...any)
	With(keyVals ...any) Logger
	Impl() any
}
```

The important distinction is:

- `Info`, `Warn`, `Error`, and `Debug` log without inspecting a `context.Context`
- `InfoContext`, `WarnContext`, `ErrorContext`, and `DebugContext` use the provided context for trace correlation

For the default `zerolog` implementation, the `*Context` methods extract the active OpenTelemetry span from `ctx` and add:

- `trace_id`
- `span_id`
- `trace_flags` when present

If there is no valid span in the context, they behave like normal log calls.

## Trace Correlation

When you want logs to line up with spans, use the context-aware methods.

```go
func (k Keeper) UpdateBalance(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) error {
	ctx, span := ctx.StartSpan(tracer, "UpdateBalance")
	defer span.End()

	logger := ctx.Logger().With(log.ModuleKey, "bank")
	logger.InfoContext(ctx, "updating balance", "address", addr.String())

	return nil
}
```

Two details matter here:

1. `sdk.Context.StartSpan(...)` returns a new `sdk.Context` with the Go `context.Context` updated to include the span.
2. The logger only sees trace information when you call one of the logger's `*Context` methods with that updated context.

Without the `*Context` call, the default logger will not add trace fields to the log record.

## `log/slog`

`cosmossdk.io/log/v2/slog` is an adapter for code that already has a standard library `*slog.Logger`.

```go
base := slog.New(handler)
logger := sdklogSlog.NewCustomLogger(base)
```

This does not add extra SDK behavior by itself. It simply makes a `*slog.Logger` satisfy the Cosmos SDK `Logger` interface. Filtering, formatting, sinks, and handler behavior are whatever the underlying `slog.Logger` is configured to do.

## `MultiLogger`

`log.NewMultiLogger(loggers...)` returns a logger that dispatches each log call to every wrapped logger.

That includes:

- ordinary log methods such as `Info(...)`
- context-aware methods such as `InfoContext(...)`
- `With(...)`, which derives a child logger for each wrapped logger

If an underlying logger implements `VerboseModeLogger`, `SetVerboseMode(...)` is also forwarded.

In other words, `MultiLogger` is just fanout. It does not merge records or add new fields on its own.

## When The SDK Configures `MultiLogger`

`MultiLogger` is not created for every app automatically.

During the node's server start, the SDK first builds the normal server logger from CLI/config flags. That logger is the usual `zerolog`-backed logger.

Then the SDK initializes OpenTelemetry from `config/otel.yaml`. If `telemetry.IsOtelLoggerEnabled()` reports that the global OpenTelemetry logger provider has active log processors/exporters, the SDK wraps the existing server logger like this:

```go
otelLogger := sdkSlog.NewCustomLogger(otelslog.NewLogger(""))
svrCtx.Logger = log.NewMultiLogger(svrCtx.Logger, otelLogger)
```

So when OpenTelemetry log exporting is enabled, one log call is sent to:

- the existing console/stdout logger
- an OpenTelemetry-backed logger for export

If OpenTelemetry logging is not enabled, the server continues using only the normal logger.

## What `otelslog` Is

`otelslog` is an OpenTelemetry bridge for Go's `log/slog` package.

More specifically, it provides a `slog.Handler` and `slog.Logger` that convert `slog.Record` values into OpenTelemetry log records and sends them to the configured OpenTelemetry logger provider.

In the Cosmos SDK startup path:

- `otelslog.NewLogger("")` creates an `*slog.Logger` backed by that bridge
- `cosmossdk.io/log/v2/slog.NewCustomLogger(...)` wraps it so it satisfies the SDK `Logger` interface
- `log.NewMultiLogger(...)` fans logs out to both the normal `zerolog` logger and the OpenTelemetry bridge

Because `slog` has native `InfoContext`/`WarnContext`/`ErrorContext`/`DebugContext` methods, the `otelslog` side receives the context directly. That means trace/span correlation is handled by the OpenTelemetry logging pipeline without the SDK needing to manually inject `trace_id` fields into that branch.

## Two Common Setups

### 1. Stdout only

If you do not configure an OpenTelemetry logger provider, logs only go to the normal SDK logger output. This does not restrict you from log correlation, however.

For trace correlation in tools such as Grafana Tempo and Loki, you can:

1. Emit JSON logs to stdout/stderr.
2. Scrape those logs with an agent such as the OpenTelemetry Collector filelog receiver.
3. Forward them to Loki.
4. Query by the `trace_id` field in the logs.

Remember, `trace_id` is only injected into the log if a contextual method was called with a context that contains an active span.

### 2. OpenTelemetry log exporter enabled

If `otel.yaml` enables an OpenTelemetry log pipeline with real log processors/exporters, the SDK configures a `MultiLogger`.

In that setup:

- console logging still works as before
- logs are also exported through OpenTelemetry
- context-aware log calls carry trace context into the OpenTelemetry branch as well

This is the path to use when you want the SDK to write logs directly into an OpenTelemetry logging backend, which eliminates the need to setup scraping infrastructure.

## Future Direction

Today the SDK uses a `MultiLogger` because the default logger is `zerolog`, while OpenTelemetry currently offers a bridge for `slog` rather than `zerolog`.

If a first-class `zerolog` bridge becomes available and suitable, that would likely be a simpler export path than maintaining a separate fanout logger. Relevant discussion:

- https://github.com/rs/zerolog/pull/682
- https://github.com/open-telemetry/opentelemetry-go-contrib/issues/5969
