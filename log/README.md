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

## Log Correlation with Tempo and Loki

To get log correlation with Tempo and Loki, simply scrape the logs in JSON format from stdout/stderr, and forward them to Loki. Then, configure Tempo with the following settings:

In the `Trace to logs` section:
- Set data source to Loki if not already
- Set `Span start time shift` to an reasonable amount of time (+/-1m)
- Set the custom query to: `{${__tags}} | trace_id = "${__trace.traceId}"`

## Future Considerations

Adding contextual methods was the most painfree way to offer log-trace correlation with OpenTelemetry while minimizing performance impacts and breakages for users.

Other considerations included changing the backend to [slog](https://pkg.go.dev/log/slog) from the Go standard lib, as well as using [otelslog](https://pkg.go.dev/go.opentelemetry.io/contrib/bridges/otelslog) directly.

It should also be noted that a zerolog bridge is also possible, however it is blocked by an open PR in the zerolog repo: https://github.com/rs/zerolog/pull/682
If this bridge is added to the core otel stack, this should be the preferred solution to log exporting. More info: https://github.com/open-telemetry/opentelemetry-go-contrib/issues/5969