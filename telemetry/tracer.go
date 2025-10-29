package telemetry

import (
	"context"
	"time"

	"cosmossdk.io/log"
	"github.com/hashicorp/go-metrics"
	//oteltrace "go.opentelemetry.io/otel/trace"
)

type TracerBase interface {
	log.Logger
	// StartSpan starts a new span with the given operation name and key-value pairs.
	// If there is a parent span in the context, the new span will be a child of that span.
	StartSpan(operation string, kvs ...any) Span
	//
	//// NewRootSpan starts a new asynchronous span with the given operation name and key-value pairs.
	//// Asynchronous spans are not tied to the current context and can outlive it.
	//NewRootSpan(operation string, kvs ...any) Span

	IncrCounter(name string, x uint64)
	SetGauge(name string, x int64)
	AddSample(name string, x float64)
}

type Tracer interface {
	TracerBase

	// StartSpanContext starts a new span with the given operation name and key-value pairs
	// and attaches it to the given context.
	// If there is a parent span in the context, the new span will be a child of that span.
	StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span)
}

type Span interface {
	// TracerBase is embedded to allow for the creation of nested spans.
	TracerBase

	// End marks the end of a span.
	// Calling End on a span that has already ended is a no-op.
	End(kvs ...any)

	// EndStatus marks the end of a span with an possibly nil error status.
	// If the error is non-nil, the span is marked as failed,
	// on the other hand, if the error is nil, the span is marked as successful.
	EndStatus(err error, kvs ...any)
}

type MetricsTracer struct {
	metrics *metrics.Metrics
	labels  []metrics.Label
}

func (m *MetricsTracer) StartSpan(operation string, kvs ...any) Span {
	return &MetricsSpan{
		tracer:    m,
		operation: []string{operation},
		start:     time.Now(),
	}
}

type spanContextKeyType string

const spanContextKey = spanContextKeyType("spanContextKey")

func (m *MetricsTracer) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span) {
	var parent TracerBase = ctx.Value(spanContextKey).(TracerBase)
	if parent == nil {
		parent = m
	}
	span := parent.StartSpan(operation, kvs...)
	ctx = context.WithValue(ctx, spanContextKey, span)
	return ctx, span
}

func (m *MetricsTracer) StartAsyncSpan(operation string, kvs ...any) Span {
	//TODO implement me
	panic("implement me")
}

type MetricsSpan struct {
	tracer    *MetricsTracer
	operation []string
	start     time.Time
}

func (m *MetricsSpan) StartSpan(operation string, kvs ...any) Span {
	op := make([]string, len(m.operation)+1)
	copy(op, m.operation)
	op[len(m.operation)] = operation
	return &MetricsSpan{
		tracer:    m.tracer,
		operation: op,
		start:     time.Now(),
	}
}

func (m *MetricsSpan) End(kvs ...any) {
	if m.operation == nil {
		// already closed
	}
	m.tracer.metrics.MeasureSinceWithLabels(m.operation, m.start, m.tracer.labels)
	m.tracer.metrics.IncrCounterWithLabels(m.operation, 1, m.tracer.labels)
	m.operation = nil
}

func (m *MetricsSpan) EndStatus(err error, kvs ...any) {
	m.End(kvs...)
}

//var _ Tracer = (*MetricsTracer)(nil)
//var _ Span = (*MetricsSpan)(nil)
//
//type OtelTracer struct {
//	tracer oteltrace.Tracer
//}
//
//func (o *OtelTracer) StartSpan(operation string, kvs ...any) Span {
//	_, span := o.tracer.Start(context.Background(), operation)
//	return &OtelSpan{span: span}
//}
//
//func (o *OtelTracer) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, Span) {
//	//TODO implement me
//	panic("implement me")
//}
//
//type OtelSpan struct {
//	span oteltrace.Span
//}
//
//var _ Tracer = (*OtelTracer)(nil)
//var _ Span = (*OtelSpan)(nil)
