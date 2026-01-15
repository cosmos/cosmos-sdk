# Log v2

The `cosmossdk.io/log/v2` provides a zerolog logging implementation for the Cosmos SDK and Cosmos SDK modules.

To use a logger wrapping an instance of the standard library's `log/slog` package, use `cosmossdk.io/log/slog`.

## OpenTelemetry

Log v2 offers includes contextual methods, allowing logs to be correlated with traces.

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

Logs will be emitted with the trace_id, span_id, and trace_flags set as fields. Logs can then be scraped and forwarded to a backend such as Loki.
You may need to configure your Loki instance to scrape these fields from the logs so they can be correlated with traces.