package telemetry

import "context"

type Tracer interface {
	// StartSpan starts a new span with the given operation name and key-value pairs.
	// If there is a parent span in the context, the new span will be a child of that span.
	StartSpan(operation string, kvs ...any) Span

	// StartSpanContext starts a new span with the given operation name and key-value pairs
	// and attaches it to the given context.
	// If there is a parent span in the context, the new span will be a child of that span.
	StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span)

	// StartAsyncSpan starts a new asynchronous span with the given operation name and key-value pairs.
	// Asynchronous spans are not tied to the current context and can outlive it.
	StartAsyncSpan(operation string, kvs ...any) Span
}

type Span interface {
	// Tracer is embedded to allow for the creation of nested spans.
	Tracer

	// End marks the end of a span and automatically closes any open child spans.
	// Calling End on a span that has already ended is a no-op.
	End(kvs ...any)
}
