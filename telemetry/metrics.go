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
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"

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
	sink                 metrics.MetricSink
	prometheusEnabled    bool
	startFuncs           []func(ctx context.Context) error
	shutdownFuncs        []func(context.Context) error
	traceProvider        log.TraceProvider
	metricsTraceProvider *MetricsTraceProvider
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
func New(cfg Config) (_ *Metrics, rerr error) {
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

	var startFuncs []func(context.Context) error
	var shutdownFuncs []func(context.Context) error

	var (
		sink metrics.MetricSink
		err  error
	)
	switch cfg.MetricsSink {
	case MetricSinkStatsd:
		sink, err = metrics.NewStatsdSink(cfg.StatsdAddr)
	case MetricSinkDogsStatsd:
		sink, err = datadog.NewDogStatsdSink(cfg.StatsdAddr, cfg.DatadogHostname)
	case MetricSinkFile:
		if cfg.MetricsFile == "" {
			return nil, errors.New("metrics-file must be set when metrics-sink is 'file'")
		}
		file, err := os.OpenFile(cfg.MetricsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open metrics file: %w", err)
		}
		fileSink := NewFileSink(file)
		sink = fileSink
		shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
			return fileSink.Close()
		})
	default:
		memSink := metrics.NewInmemSink(10*time.Second, time.Minute)
		sink = memSink
		inMemSig := metrics.DefaultInmemSignal(memSink)
		defer func() {
			if rerr != nil {
				inMemSig.Stop()
			}
		}()
	}

	metricsTraceProvider := NewMetricsTraceProvider(nil, globalLabels, metrics.Default())

	var tracerBase log.TraceProvider
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
					client = otlptracegrpc.NewClient(opts...)
				default:
					opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
					if exporterOpts.Insecure {
						opts = append(opts, otlptracehttp.WithInsecure())
					}
					client = otlptracehttp.NewClient(opts...)
				}
				exporter := otlptrace.NewUnstarted(client)
				startFuncs = append(startFuncs, exporter.Start)
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
					shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
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

		tracerProvider := otelsdktrace.NewTracerProvider(tracerProviderOpts...)
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		tracerName := cfg.ServiceName
		if tracerName == "" {
			tracerName = "cosmos-sdk"
		}
		tracerBase = NewOtelTraceProvider(tracerProvider.Tracer(tracerName))

	case "metrics":
		tracerBase = metricsTraceProvider
	default:
		tracerBase = log.NewNopTracer()
	}

	if err != nil {
		return nil, err
	}

	m := &Metrics{
		sink:                 sink,
		startFuncs:           startFuncs,
		shutdownFuncs:        shutdownFuncs,
		traceProvider:        tracerBase,
		metricsTraceProvider: metricsTraceProvider,
	}
	fanout := metrics.FanoutSink{sink}

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

// TraceProvider returns the base trace provider for creating spans.
func (m *Metrics) TraceProvider() log.TraceProvider {
	return m.traceProvider
}

// MetricsTraceProvider returns a trace provider that only emits metrics for spans.
// Use this when you specifically want to configure a code path to only emit metrics
// and not actual logging spans (useful for benchmarking small operations such as store operations).
func (m *Metrics) MetricsTraceProvider() log.TraceProvider {
	return m.metricsTraceProvider
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
	for _, f := range m.shutdownFuncs {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}
