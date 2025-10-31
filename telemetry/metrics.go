package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/go-metrics/datadog"
	metricsprom "github.com/hashicorp/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"cosmossdk.io/log"
)

// Metrics provides access to the application's metrics collection system.
// It wraps the go-metrics global registry and configured sinks.
//
// When using the in-memory sink, sending SIGUSR1 to the process (kill -USR1 <pid>)
// will dump current metrics to stderr for debugging.
//
// The Metrics object maintains references to configured sinks and provides
// a Gather() method for pull-based metric retrieval (useful for testing and monitoring).
//
// When using the file sink, call Close() to flush buffered data and close the file.
//
// Note: go-metrics uses a singleton global registry. Only one Metrics instance
// should be created per process.
type Metrics struct {
	sink              metrics.MetricSink
	prometheusEnabled bool
	startFuncs        []func(ctx context.Context) error
	shutdownFuncs     []func(context.Context) error
	logger            log.Logger
	tracer            log.Tracer
	metricsTracer     *MetricsTracer

	// OpenTelemetry metrics (when OtelMetricsExporters is configured)
	meterProvider *sdkmetric.MeterProvider
	meter         metric.Meter

	// Instrument cache for wrapper functions
	counters   map[string]metric.Int64Counter
	gauges     map[string]metric.Float64Gauge
	histograms map[string]metric.Float64Histogram
	mu         sync.RWMutex
}

// GatherResponse contains collected metrics in the requested format.
// The Metrics field holds the serialized metric data, and ContentType
// indicates how it's encoded ("application/json" or prometheus text format).
type GatherResponse struct {
	Metrics     []byte
	ContentType string
}

// New creates and initializes the metrics system with the given configuration.
//
// Returns nil if telemetry is disabled (cfg.Enabled == false), which allows
// callers to safely ignore the Metrics object.
//
// The function:
//   - Initializes the go-metrics global registry
//   - Configures the specified sink(s) (mem, prometheus, statsd, dogstatsd, file)
//   - Sets up global labels to be applied to all metrics
//   - Enables SIGUSR1 signal handling for in-memory sink dumps
//   - Creates a FanoutSink if multiple sinks are needed (e.g., mem + prometheus)
//
// Example:
//
//	m, err := telemetry.New(telemetry.Config{
//		Enabled:                 true,
//		ServiceName:             "cosmos-app",
//		MetricsSink:             telemetry.MetricSinkInMem,
//		PrometheusRetentionTime: 60,
//	})
//	if err != nil {
//		return err
//	}
//	defer m.Close()
func New(cfg Config, opts ...Option) (_ *Metrics, rerr error) {
	globalTelemetryEnabled = cfg.Enabled
	if !cfg.Enabled {
		return nil, nil
	}

	if numGlobalLabels := len(cfg.GlobalLabels); numGlobalLabels > 0 {
		parsedGlobalLabels := make([]metrics.Label, numGlobalLabels)
		for i, gl := range cfg.GlobalLabels {
			parsedGlobalLabels[i] = NewLabel(gl[0], gl[1])
		}
		globalLabels = parsedGlobalLabels
	}

	metricsConf := metrics.DefaultConfig(cfg.ServiceName)
	metricsConf.EnableHostname = cfg.EnableHostname
	metricsConf.EnableHostnameLabel = cfg.EnableHostnameLabel

	// Initialize Metrics struct early with default logger
	m := &Metrics{
		logger: log.NewNopLogger(), // Default to nop logger
	}
	// Apply functional options to set logger
	for _, opt := range opts {
		opt(m)
	}

	var err error
	switch cfg.MetricsSink {
	case MetricSinkStatsd:
		m.sink, err = metrics.NewStatsdSink(cfg.StatsdAddr)
	case MetricSinkDogsStatsd:
		m.sink, err = datadog.NewDogStatsdSink(cfg.StatsdAddr, cfg.DatadogHostname)
	case MetricSinkFile:
		if cfg.MetricsFile == "" {
			return nil, errors.New("metrics-file must be set when metrics-sink is 'file'")
		}
		file, err := os.OpenFile(cfg.MetricsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open metrics file: %w", err)
		}
		fileSink := NewFileSink(file)
		m.shutdownFuncs = append(m.shutdownFuncs, func(ctx context.Context) error {
			return file.Close()
		})
		m.sink = fileSink
	default:
		memSink := metrics.NewInmemSink(10*time.Second, time.Minute)
		m.sink = memSink
		inMemSig := metrics.DefaultInmemSignal(memSink)
		defer func() {
			if rerr != nil {
				inMemSig.Stop()
			}
		}()
	}
	if err != nil {
		return nil, err
	}

	m.tracer = log.NewNopTracer(m.logger)
	m.metricsTracer = NewMetricsTracer(metrics.Default(), nil, globalLabels, m.logger)

	switch cfg.TraceSink {
	case TraceSinkOtel:
		var tracerProviderOpts []otelsdktrace.TracerProviderOption
		for _, exporterOpts := range cfg.OtelTraceExporters {
			switch exporterOpts.Type {
			case "otlp":
				endpoint := exporterOpts.Endpoint
				if endpoint == "" {
					return nil, fmt.Errorf("otlp endpoint must be set")
				}
				var client otlptrace.Client
				switch exporterOpts.OTLPTransport {
				case "grpc":
					opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
					if exporterOpts.Insecure {
						opts = append(opts, otlptracegrpc.WithInsecure())
					}
					if len(exporterOpts.Headers) > 0 {
						opts = append(opts, otlptracegrpc.WithHeaders(exporterOpts.Headers))
					}
					client = otlptracegrpc.NewClient(opts...)
				default:
					opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
					if exporterOpts.Insecure {
						opts = append(opts, otlptracehttp.WithInsecure())
					}
					if len(exporterOpts.Headers) > 0 {
						opts = append(opts, otlptracehttp.WithHeaders(exporterOpts.Headers))
					}
					client = otlptracehttp.NewClient(opts...)
				}
				exporter := otlptrace.NewUnstarted(client)
				m.startFuncs = append(m.startFuncs, exporter.Start)
				batcherOpt := otelsdktrace.WithBatcher(exporter)
				tracerProviderOpts = append(tracerProviderOpts, batcherOpt)
			case "stdout":
				var opts []stdouttrace.Option
				if exporterOpts.File != "" {
					file, err := os.OpenFile(exporterOpts.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
					if err != nil {
						return nil, fmt.Errorf("failed to open stdout trace file %s: %w", exporterOpts.File, err)
					}
					opts = append(opts, stdouttrace.WithWriter(file))
					m.shutdownFuncs = append(m.shutdownFuncs, func(ctx context.Context) error {
						return file.Close()
					})
				}
				if exporterOpts.PrettyPrint {
					opts = append(opts, stdouttrace.WithPrettyPrint())
				}
				exporter, err := stdouttrace.New(opts...)
				if err != nil {
					return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
				}
				batcher := otelsdktrace.WithBatcher(exporter)
				tracerProviderOpts = append(tracerProviderOpts, batcher)
			default:
				return nil, fmt.Errorf("unknown trace exporter type: %s", exporterOpts.Type)
			}
		}

		// Determine service name for both resource and tracer
		serviceName := cfg.ServiceName
		if serviceName == "" {
			serviceName = "cosmos-sdk"
		}

		// Create OpenTelemetry resource with service name
		res, err := resource.New(context.Background(),
			resource.WithAttributes(
				semconv.ServiceName(serviceName),
			),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace resource: %w", err)
		}

		// Add resource to tracer provider options
		tracerProviderOpts = append(tracerProviderOpts, otelsdktrace.WithResource(res))
		tracerProvider := otelsdktrace.NewTracerProvider(tracerProviderOpts...)
		m.shutdownFuncs = append(m.shutdownFuncs, tracerProvider.Shutdown)
		m.tracer = NewOtelTracer(tracerProvider.Tracer(serviceName), m.logger)
	case "metrics":
		m.tracer = m.metricsTracer
	default:
	}

	fanout := metrics.FanoutSink{m.sink}

	if cfg.PrometheusRetentionTime > 0 {
		m.prometheusEnabled = true
		prometheusOpts := metricsprom.PrometheusOpts{
			Expiration: time.Duration(cfg.PrometheusRetentionTime) * time.Second,
		}

		promSink, err := metricsprom.NewPrometheusSinkFrom(prometheusOpts)
		if err != nil {
			return nil, err
		}

		fanout = append(fanout, promSink)
	}

	if _, err := metrics.NewGlobal(metricsConf, fanout); err != nil {
		return nil, err
	}

	return m, nil
}

// Gather collects all registered metrics and returns a GatherResponse where the
// metrics are encoded depending on the type. Metrics are either encoded via
// Prometheus or JSON if in-memory.
func (m *Metrics) Gather(format string) (GatherResponse, error) {
	switch format {
	case FormatPrometheus:
		return m.gatherPrometheus()

	case FormatText:
		return m.gatherGeneric()

	case FormatDefault:
		return m.gatherGeneric()

	default:
		return GatherResponse{}, fmt.Errorf("unsupported metrics format: %s", format)
	}
}

// gatherPrometheus collects Prometheus metrics and returns a GatherResponse.
// If Prometheus metrics are not enabled, it returns an error.
func (m *Metrics) gatherPrometheus() (GatherResponse, error) {
	if !m.prometheusEnabled {
		return GatherResponse{}, errors.New("prometheus metrics are not enabled")
	}

	metricsFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather prometheus metrics: %w", err)
	}

	buf := &bytes.Buffer{}
	defer buf.Reset()

	e := expfmt.NewEncoder(buf, expfmt.NewFormat(expfmt.TypeTextPlain))

	for _, mf := range metricsFamilies {
		if err := e.Encode(mf); err != nil {
			return GatherResponse{}, fmt.Errorf("failed to encode prometheus metrics: %w", err)
		}
	}

	return GatherResponse{ContentType: ContentTypeText, Metrics: buf.Bytes()}, nil
}

// gatherGeneric collects generic metrics and returns a GatherResponse.
func (m *Metrics) gatherGeneric() (GatherResponse, error) {
	gm, ok := m.sink.(DisplayableSink)
	if !ok {
		return GatherResponse{}, errors.New("non in-memory metrics sink does not support generic format")
	}

	summary, err := gm.DisplayMetrics(nil, nil)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to gather in-memory metrics: %w", err)
	}

	content, err := json.Marshal(summary)
	if err != nil {
		return GatherResponse{}, fmt.Errorf("failed to encode in-memory metrics: %w", err)
	}

	return GatherResponse{ContentType: "application/json", Metrics: content}, nil
}

// Tracer returns the base tracer for creating spans and logging.
func (m *Metrics) Tracer() log.Tracer {
	return m.tracer
}

// MetricsTracer returns a tracer that only emits metrics for spans.
// Use this when you specifically want to configure a code path to only emit metrics
// and not actual logging spans (useful for benchmarking small operations such as store operations).
func (m *Metrics) MetricsTracer() log.Tracer {
	return m.metricsTracer
}

// Start starts all configured exporters.
// Start should be called after New() in order to ensure that all configured
// exporters are started.
func (m *Metrics) Start(ctx context.Context) error {
	for _, f := range m.startFuncs {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown must be called before the application exits to shutdown any
// exporters and close any open files.
func (m *Metrics) Shutdown(ctx context.Context) error {
	n := len(m.shutdownFuncs)
	// shutdown in reverse order because we can't close files until after the exporter is stopped
	for i := n - 1; i >= 0; i-- {
		if err := m.shutdownFuncs[i](ctx); err != nil {
			return err
		}
	}
	return nil
}
