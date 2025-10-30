package telemetry

import (
	"context"
	"fmt"

	otelattr "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"

	"cosmossdk.io/log"
)

// OtelTracer is a log.Tracer implementation that uses OpenTelemetry for distributed tracing.
// It wraps a logger and forwards all log calls to it, while creating OtelSpan instances
// that emit trace spans.
type OtelTracer struct {
	log.Logger
	tracer oteltrace.Tracer
}

func NewOtelTracer(tracer oteltrace.Tracer, logger log.Logger) *OtelTracer {
	return &OtelTracer{
		Logger: logger,
		tracer: tracer,
	}
}

func (o *OtelTracer) StartSpan(operation string, kvs ...any) log.Span {
	ctx, span := o.tracer.Start(context.Background(), operation, oteltrace.WithAttributes(toKVs(kvs)...))
	return &OtelSpan{
		tracer: o.tracer,
		ctx:    ctx,
		span:   span,
	}
}

func (o *OtelTracer) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	ctx, span := o.tracer.Start(ctx, operation, oteltrace.WithAttributes(toKVs(kvs)...))
	return ctx, &OtelSpan{
		tracer: o.tracer,
		ctx:    ctx,
		span:   span,
	}
}

func (o *OtelTracer) StartRootSpan(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	ctx, span := o.tracer.Start(ctx, operation, oteltrace.WithAttributes(toKVs(kvs)...), oteltrace.WithNewRoot())
	return ctx, &OtelSpan{
		tracer: o.tracer,
		ctx:    ctx,
		span:   span,
	}
}

var _ log.Tracer = (*OtelTracer)(nil)

type OtelSpan struct {
	tracer          oteltrace.Tracer
	ctx             context.Context
	span            oteltrace.Span
	persistentAttrs []otelattr.KeyValue
}

func (o *OtelSpan) addEvent(level, msg string, keyVals ...any) {
	o.span.AddEvent(msg,
		oteltrace.WithAttributes(o.persistentAttrs...),
		oteltrace.WithAttributes(toKVs(keyVals...)...),
		oteltrace.WithAttributes(otelattr.String("level", level)),
	)
}

func (o *OtelSpan) Info(msg string, keyVals ...any) {
	o.addEvent("info", msg, keyVals...)
}

func (o *OtelSpan) Warn(msg string, keyVals ...any) {
	o.addEvent("warn", msg, keyVals...)
}

func (o *OtelSpan) Error(msg string, keyVals ...any) {
	o.addEvent("error", msg, keyVals...)
}

func (o *OtelSpan) Debug(msg string, keyVals ...any) {
	o.addEvent("debug", msg, keyVals...)
}

func (o *OtelSpan) With(keyVals ...any) log.Logger {
	attrs := toKVs(keyVals...)
	persistentAttrs := make([]otelattr.KeyValue, 0, len(o.persistentAttrs)+len(attrs))
	persistentAttrs = append(persistentAttrs, o.persistentAttrs...)
	persistentAttrs = append(persistentAttrs, attrs...)
	return &OtelSpan{
		tracer:          o.tracer,
		ctx:             o.ctx,
		span:            o.span,
		persistentAttrs: persistentAttrs,
	}
}

func (o *OtelSpan) Impl() any {
	return o.span
}

func (o *OtelSpan) startSpan(ctx context.Context, operation string, kvs []any, opts ...oteltrace.SpanStartOption) *OtelSpan {
	if len(o.persistentAttrs) > 0 {
		opts = append(opts, oteltrace.WithAttributes(o.persistentAttrs...))
	}
	opts = append(opts, oteltrace.WithAttributes(toKVs(kvs...)...))
	ctx, span := o.tracer.Start(ctx, operation, opts...)
	return &OtelSpan{
		tracer:          o.tracer,
		ctx:             ctx,
		span:            span,
		persistentAttrs: o.persistentAttrs,
	}
}

func (o *OtelSpan) StartSpan(operation string, kvs ...any) log.Span {
	return o.startSpan(o.ctx, operation, kvs)
}

func (o *OtelSpan) StartSpanContext(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	if !oteltrace.SpanContextFromContext(ctx).IsValid() {
		// if we don't have a valid span in the context, use the one from the tracer
		ctx = o.ctx
	}
	span := o.startSpan(ctx, operation, kvs)
	return span.ctx, span
}

func (o *OtelSpan) StartRootSpan(ctx context.Context, operation string, kvs ...any) (context.Context, log.Span) {
	span := o.startSpan(ctx, operation, kvs, oteltrace.WithNewRoot())
	return span.ctx, span
}

func (o *OtelSpan) SetAttrs(kvs ...any) {
	o.span.SetAttributes(toKVs(kvs...)...)
}

func (o *OtelSpan) SetErr(err error, kvs ...any) error {
	if err == nil {
		o.span.SetStatus(otelcodes.Ok, "OK")
	} else {
		o.span.RecordError(err)
		o.span.SetStatus(otelcodes.Error, err.Error())
	}
	if len(kvs) > 0 {
		o.span.SetAttributes(toKVs(kvs...)...)
	}
	return err
}

func (o *OtelSpan) End() {
	o.span.End()
}

var _ log.Span = (*OtelSpan)(nil)

func toKVs(kvs ...any) []otelattr.KeyValue {
	if len(kvs)%2 != 0 {
		panic(fmt.Sprintf("kvs must have even length, got %d", len(kvs)))
	}
	res := make([]otelattr.KeyValue, 0, len(kvs)/2)
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			panic("key must be string")
		}
		res = append(res, otelattr.KeyValue{
			Key:   otelattr.Key(key),
			Value: toValue(kvs[i+1]),
		})
	}
	return res
}

func toValue(value any) otelattr.Value {
	switch v := value.(type) {
	case bool:
		return otelattr.BoolValue(v)
	case string:
		return otelattr.StringValue(v)
	case int64:
		return otelattr.Int64Value(v)
	case int:
		return otelattr.IntValue(v)
	case float64:
		return otelattr.Float64Value(v)
	case []string:
		return otelattr.StringSliceValue(v)
	case []int64:
		return otelattr.Int64SliceValue(v)
	case []int:
		return otelattr.IntSliceValue(v)
	case []float64:
		return otelattr.Float64SliceValue(v)
	default:
		return otelattr.StringValue(fmt.Sprintf("%+v", value))
	}

}
