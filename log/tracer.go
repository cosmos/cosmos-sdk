package log

import "context"

// Tracer is an interface for creating and managing spans.
// It may be backed by open telemetry or other tracing libraries,
// Spans may also be used for collecting timing metrics.
// It embeds the Logger interface. Log events may be associated with spans.
type Tracer interface {
	Logger

	// StartSpan starts a new span with the given operation name and key-value pair attributes.
	// If there is a parent span, the new span will be a child of that span.
	// It is recommended to use a defer statement to end the span like this:
	// 	 span := tracer.StartSpan("my-span")
	// 	 defer span.End()
	StartSpan(operation string, kvs ...any) Span

	// StartSpanContext attempts to retrieve an existing tracer from the context and then starts a new span
	// as a child of that span.
	// If no tracer is found, it returns a new span that is a child of this tracer instance.
	// This is useful if a span may have been set in the context, but we are not sure.
	// The function also returns a context with the span added to it.
	StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span)

	// RootTracer returns a root-level tracer that is not part of any existing span.
	// Use this when starting async work to ensure its spans are not timed as part of the current span.
	// Example usage:
	//   rootTracer := tracer.RootTracer()
	// 	 go func() {
	//		span := rootTracer.StartSpan("my-go-routine")
	// 	 	defer span.End()
	// 	 	doSomething()
	//	 }()
	RootTracer() Tracer
}

// Span is an interface for managing spans and creating nested spans via the embedded Tracer interface.
type Span interface {
	// Tracer is embedded to allow for the creation of nested spans.
	Tracer

	// SetAttrs sets additional key-value attributes on the span.
	SetAttrs(kvs ...any)

	// SetErr records an optional error on the span and optionally adds additional key-value pair attributes.
	// It returns the error value unchanged, allowing use in return statements.
	// If err is nil, the span is marked as successful.
	// If err is not nil, the span is marked as failed.
	// This does NOT end the span, you must still call End.
	// Example usage:
	// 	 span := tracer.StartSpan("my-span")
	// 	 defer span.End()
	//   err := doSomething()
	// 	 return span.SetErr(err, "additional", "info") // okay to call with a nil error
	SetErr(err error, kvs ...any) error

	// End marks the end of a span and is designed to be used in a defer statement right after the span is created.
	// Calling End on a span that has already ended is a no-op.
	// Example usage:
	// 	 span := tracer.StartSpan("my-span")
	// 	 defer span.End()
	End()
}

type traceContextKeyType string

const traceContextKey traceContextKeyType = "trace-context"

func ContextWithTracer(ctx context.Context, tracer Tracer) context.Context {
	return context.WithValue(ctx, traceContextKey, tracer)
}

// TracerFromContext returns the Tracer from the context.
// If no Tracer is found, it returns a no-op Tracer that is safe to use.
func TracerFromContext(ctx context.Context) Tracer {
	if tracer, ok := ctx.Value(traceContextKey).(Tracer); ok {
		return tracer
	}
	return NewNopTracer()
}

// NewNopTracer returns a Tracer that does nothing.
func NewNopTracer() Tracer {
	return nopTracer{}
}

type nopTracer struct {
	nopLogger
}

func (n nopTracer) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span) {
	if tracer, ok := ctx.Value(traceContextKey).(Tracer); ok {
		span := tracer.StartSpan(operation, kvs...)
		return ContextWithTracer(ctx, span), span
	}
	// no tracer found in context, create a new span, no need to add to context
	return ctx, nopSpan{}
}

func (n nopTracer) StartSpan(string, ...any) Span {
	return nopSpan{}
}

func (n nopTracer) RootTracer() Tracer {
	return n
}

type nopSpan struct {
	nopTracer
}

func (n nopSpan) SetAttrs(...any) {}

func (n nopSpan) SetErr(err error, _ ...any) error { return err }

func (n nopSpan) End() {}

var _ Tracer = nopTracer{}
var _ Span = nopSpan{}
