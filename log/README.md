# Log v2

The `cosmossdk.io/log/v2` provides a zerolog logging implementation for the Cosmos SDK and Cosmos SDK modules.

To use a logger wrapping an instance of the standard library's `log/slog` package, use `cosmossdk.io/log/slog`.

## OpenTelemetry

The v2 release of log adds contextual methods to the `Logger` interface. This allows logs to be correlated with traces.

```go
// Logger is the Cosmos SDK logger interface.
type Logger interface {
	// InfoContext is like Info but extracts trace context from ctx for correlation.
	InfoContext(ctx context.Context, msg string, keyVals ...any)

	// WarnContext is like Warn but extracts trace context from ctx for correlation.
	WarnContext(ctx context.Context, msg string, keyVals ...any)

	// ErrorContext is like Error but extracts trace context from ctx for correlation.
	ErrorContext(ctx context.Context, msg string, keyVals ...any)

	// DebugContext is like Debug but extracts trace context from ctx for correlation.
	DebugContext(ctx context.Context, msg string, keyVals ...any)
}
```

If your OpenTelemetry configuration contained a [logger provider](https://opentelemetry.io/docs/concepts/signals/logs/#logger-provider), the Cosmos SDK will automatically provision a MultiLogger,
logging to both stdout via the existing `zerolog` wrapper, and to the configured logger provider via `otelslog`.

## Log Correlation with Tempo and Loki

To get log correlation with Tempo and Loki, simply scrape the logs in JSON format from stdout/stderr and forward them to Loki. Then, configure Tempo with the following settings:

In the `Trace to logs` section:
- Set data source to Loki if not already
- Set `Span start time shift` to an reasonable amount of time (+/-1m)
- Set the custom query to: `{${__tags}} | trace_id = "${__trace.traceId}"`

## Future Considerations

OpenTelemetry has built log bridges, which automatically forward logs to the configured logger provider. For example, OpenTelemetry has bridges for zap and slog, but currently, not zerolog.
A zerolog bridge is possible, however it is blocked by an open PR in the zerolog repo: https://github.com/rs/zerolog/pull/682
If this bridge is added to the core otel stack, this should be the preferred solution to log exporting instead of a MultiLogger. More info: https://github.com/open-telemetry/opentelemetry-go-contrib/issues/5969