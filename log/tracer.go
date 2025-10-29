package log

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
