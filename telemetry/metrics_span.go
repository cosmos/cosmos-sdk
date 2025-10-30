package telemetry

import (
	"context"
	"time"

	"github.com/hashicorp/go-metrics"

	"cosmossdk.io/log"
)

type MetricsTraceProvider struct {
	rootPath   []string        // Base path set at tracer creation, preserved across all spans
	rootLabels []metrics.Label // Labels applied to all metrics emitted by this tracer
	metrics    *metrics.Metrics
}

func NewMetricsTraceProvider(rootPath []string, rootLabels []metrics.Label, metrics *metrics.Metrics) *MetricsTraceProvider {
	return &MetricsTraceProvider{rootPath: rootPath, rootLabels: rootLabels, metrics: metrics}
}

func (m *MetricsTraceProvider) startSpan(existingPath []string, operation string, _ ...any) log.Span {
	path := make([]string, len(existingPath)+1)
	copy(path, existingPath)
	path[len(path)-1] = operation
	return &MetricsSpan{
		MetricsTraceProvider: m,
		path:                 []string{operation},
		start:                time.Now(),
	}
}

func (m *MetricsTraceProvider) StartSpan(operation string, _ ...any) log.Span {
	return m.startSpan(m.rootPath, operation)
}

func (m *MetricsTraceProvider) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	return ctx, m.StartSpan(operation, kvs...)
}

func (m *MetricsTraceProvider) StartRootSpan(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	return ctx, m.StartSpan(operation, kvs...)
}

var _ log.TraceProvider = (*MetricsTraceProvider)(nil)

// MetricsSpan is a log.Span implementation that emits timing and count metrics
// to go-metrics when the span ends.
//
// Unlike distributed tracing spans, MetricsSpan:
//   - Does not support logging (Info/Warn/Error/Debug are no-ops)
//   - Ignores span attributes (kvs parameters)
//   - Emits aggregated metrics rather than individual trace events
//
// When End() is called, two metrics are emitted:
//   - A timer metric with ".time" suffix (e.g., "query.get.time")
//   - A counter metric with ".count" suffix (e.g., "query.get.count")
//
// Root path and labels are preserved across all spans created from this tracer,
// ensuring consistent metric namespacing and labeling throughout the span hierarchy.
type MetricsSpan struct {
	*MetricsTraceProvider
	start time.Time
	path  []string
}

// NewMetricsTracer creates a new MetricsSpan that acts as a root tracer.
// The rootPath defines the base metric name. If empty, metrics start from the root.
// The rootLabels are applied to all metrics emitted by this tracer and its children.
//
// Example:
//
//	labels := []metrics.Label{{Name: "module", Value: "staking"}}
//	tracer := NewMetricsTracer(metrics, []string{"app", "tx"}, labels)
//	span := tracer.StartSpan("validate")
//	defer span.End()
//	// Emits: "app.tx.validate.time" and "app.tx.validate.count" with module=staking label
func NewMetricsTracer(m *metrics.Metrics, rootPath []string, rootLabels []metrics.Label) *MetricsSpan {
	return &MetricsSpan{
		MetricsTraceProvider: &MetricsTraceProvider{
			metrics:    m,
			rootPath:   rootPath,
			rootLabels: rootLabels,
		},
		path:  rootPath,
		start: time.Now(),
	}
}

// Logger methods are no-ops - MetricsSpan does not support logging.
func (m *MetricsSpan) Info(msg string, keyVals ...any)  {}
func (m *MetricsSpan) Warn(msg string, keyVals ...any)  {}
func (m *MetricsSpan) Error(msg string, keyVals ...any) {}
func (m *MetricsSpan) Debug(msg string, keyVals ...any) {}
func (m *MetricsSpan) With(keyVals ...any) log.Logger   { return m }
func (m *MetricsSpan) Impl() any                        { return nil }

// StartSpan creates a child span by appending the operation name to the current path.
// Root path and labels are preserved in the child span.
// The kvs parameters are ignored (metrics don't support dynamic attributes).
func (m *MetricsSpan) StartSpan(operation string, kvs ...any) log.Span {
	return m.startSpan(m.path, operation)
}

// StartSpanContext creates a child span and returns the context unchanged.
// The span is not stored in the context.
func (m *MetricsSpan) StartSpanContext(ctx context.Context, operation string, _ ...any) (context.Context, log.Span) {
	return ctx, m.startSpan(m.path, operation)
}

// StartRootSpan creates a new root span by combining the tracer's root path with the operation.
// Unlike StartSpan, this does not extend the current span's path - it starts fresh from the root path.
// Root labels are preserved in the new span.
//
// Example:
//
//	tracer := NewMetricsTracer(metrics, []string{"app"}, nil)
//	_, span := tracer.StartRootSpan(ctx, "process")
//	defer span.End()
//	// Emits: "app.process.time" and "app.process.count"
func (m *MetricsSpan) StartRootSpan(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	return ctx, m.startSpan(m.rootPath, operation)
}

// SetAttrs is a no-op - metrics don't support dynamic attributes.
func (m *MetricsSpan) SetAttrs(...any) {}

// SetErr is a no-op but returns the error unchanged for convenience.
func (m *MetricsSpan) SetErr(err error, _ ...any) error { return err }

// End emits timing and count metrics with ".time" and ".count" suffixes.
// Root labels are applied to both metrics.
//
// For a span with path ["query", "get"], this emits:
//   - Timer: "query.get.time" with duration since start
//   - Counter: "query.get.count" incremented by 1
//
// If root labels were set (e.g., module=staking), they are included in both metrics.
func (m *MetricsSpan) End() {
	// Create paths with suffixes to avoid metric type conflicts
	timePath := make([]string, len(m.path)+1)
	copy(timePath, m.path)
	timePath[len(timePath)-1] = "time"

	countPath := make([]string, len(m.path)+1)
	copy(countPath, m.path)
	countPath[len(countPath)-1] = "count"

	// Apply root labels to metrics
	if len(m.rootLabels) > 0 {
		m.metrics.MeasureSinceWithLabels(timePath, m.start, m.rootLabels)
		m.metrics.IncrCounterWithLabels(countPath, 1, m.rootLabels)
	} else {
		m.metrics.MeasureSince(timePath, m.start)
		m.metrics.IncrCounter(countPath, 1)
	}
}

var _ log.Span = (*MetricsSpan)(nil)
var _ log.Tracer = (*MetricsSpan)(nil)
