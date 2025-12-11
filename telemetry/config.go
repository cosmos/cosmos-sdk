package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	otelconf "go.opentelemetry.io/contrib/otelconf/v0.3.0"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	logglobal "go.opentelemetry.io/otel/log/global"
	lognoop "go.opentelemetry.io/otel/log/noop"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
	"go.yaml.in/yaml/v3"
)

const (
	OtelFileName = "otel.yaml"

	otelConfigEnvVar = "OTEL_EXPERIMENTAL_CONFIG_FILE"
)

var (
	openTelemetrySDK *otelconf.SDK
	shutdownFuncs    []func(context.Context) error
)

func init() {
	if openTelemetrySDK == nil {
		if otelFilePath := os.Getenv(otelConfigEnvVar); otelFilePath != "" {
			if err := InitializeOpenTelemetry(otelFilePath); err != nil {
				panic(err)
			}
		}
	}
}

// InitializeOpenTelemetry initializes the OpenTelemetry SDK.
// We assume that the otel configuration file is in `~/.<your_node_home>/config/otel.yaml`.
// An empty otel.yaml is automatically placed in the directory above in the `appd init` command.
//
// Note that a late initialization of the open telemetry SDK causes meters/tracers to utilize a delegate, which incurs
// an atomic load.
// In our benchmarks, we saw only a few nanoseconds incurred from this atomic operation.
// If you wish to avoid this overhead entirely, you may set the OTEL_EXPERIMENTAL_CONFIG_FILE environment variable,'
// and the OpenTelemetry SDK will be instantiated via init.
// This will eliminate the atomic operation overhead.
func InitializeOpenTelemetry(filePath string) error {
	if openTelemetrySDK != nil {
		return nil
	}
	var err error

	var opts []otelconf.ConfigurationOption

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			setNoop()
			return nil
		}
		return err // return other errors (permission issues, etc.)
	}

	bz, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read telemetry config file: %w", err)
	}
	if len(bz) == 0 {
		setNoop()
		return nil
	}

	cfg, err := otelconf.ParseYAML(bz)
	if err != nil {
		return fmt.Errorf("failed to parse telemetry config file: %w", err)
	}

	opts = append(opts, otelconf.WithOpenTelemetryConfiguration(*cfg))

	// parse cosmos extra config
	var extraCfg extraConfig
	err = yaml.Unmarshal(bz, &extraCfg)
	if err == nil {
		if extraCfg.CosmosExtra != nil {
			extra := *extraCfg.CosmosExtra
			if extra.TraceFile != "" {
				traceFile, err := os.OpenFile(extra.TraceFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					return fmt.Errorf("failed to open trace file: %w", err)
				}
				shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
					if err := traceFile.Close(); err != nil {
						return fmt.Errorf("failed to close trace file: %w", err)
					}
					return nil
				})
				exporter, err := stdouttrace.New(
					stdouttrace.WithWriter(traceFile),
					// stdouttrace.WithPrettyPrint(),
				)
				if err != nil {
					return fmt.Errorf("failed to create stdout trace exporter: %w", err)
				}
				opts = append(opts, otelconf.WithTracerProviderOptions(
					tracesdk.WithBatcher(exporter),
				))
			}
			if extra.MetricsFile != "" {
				metricsFile, err := os.OpenFile(extra.MetricsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					return fmt.Errorf("failed to open metrics file: %w", err)
				}
				shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
					if err := metricsFile.Close(); err != nil {
						return fmt.Errorf("failed to close metrics file: %w", err)
					}
					return nil
				})
				exporter, err := stdoutmetric.New(
					stdoutmetric.WithWriter(metricsFile),
					// stdoutmetric.WithPrettyPrint(),
				)
				if err != nil {
					return fmt.Errorf("failed to create stdout metric exporter: %w", err)
				}

				// Configure periodic reader with custom interval if specified
				readerOpts := []metricsdk.PeriodicReaderOption{}
				if extra.MetricsFileInterval != "" {
					interval, err := time.ParseDuration(extra.MetricsFileInterval)
					if err != nil {
						return fmt.Errorf("failed to parse metrics_file_interval: %w", err)
					}
					readerOpts = append(readerOpts, metricsdk.WithInterval(interval))
				}

				opts = append(opts, otelconf.WithMeterProviderOptions(
					metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter, readerOpts...)),
				))
			}
			if extra.LogsFile != "" {
				logsFile, err := os.OpenFile(extra.LogsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
				if err != nil {
					return fmt.Errorf("failed to open logs file: %w", err)
				}
				shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
					if err := logsFile.Close(); err != nil {
						return fmt.Errorf("failed to close logs file: %w", err)
					}
					return nil
				})
				exporter, err := stdoutlog.New(
					stdoutlog.WithWriter(logsFile),
					// stdoutlog.WithPrettyPrint(),
				)
				if err != nil {
					return fmt.Errorf("failed to create stdout log exporter: %w", err)
				}
				opts = append(opts, otelconf.WithLoggerProviderOptions(
					logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter)),
				))
			}
			if extra.InstrumentHost {
				fmt.Println("Initializing host instrumentation")
				if err := host.Start(); err != nil {
					return fmt.Errorf("failed to start host instrumentation: %w", err)
				}
			}
			if extra.InstrumentRuntime {
				fmt.Println("Initializing runtime instrumentation")
				if err := runtime.Start(); err != nil {
					return fmt.Errorf("failed to start runtime instrumentation: %w", err)
				}
			}

			// TODO: this code should be removed once propagation is properly supported by otelconf.
			if len(extra.Propagators) > 0 {
				propagator := initPropagator(extra.Propagators)
				otel.SetTextMapPropagator(propagator)
			}
		}
	}

	otelSDK, err := otelconf.NewSDK(opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	openTelemetrySDK = &otelSDK

	// setup otel global providers
	otel.SetTracerProvider(openTelemetrySDK.TracerProvider())
	otel.SetMeterProvider(openTelemetrySDK.MeterProvider())
	logglobal.SetLoggerProvider(openTelemetrySDK.LoggerProvider())

	return nil
}

func initPropagator(propagatorTypes []string) propagation.TextMapPropagator {
	var propagators []propagation.TextMapPropagator

	for _, name := range propagatorTypes {
		switch name {
		case "tracecontext":
			propagators = append(propagators, propagation.TraceContext{})
		case "baggage":
			propagators = append(propagators, propagation.Baggage{})
		case "b3":
			propagators = append(propagators, b3.New())
		case "b3multi":
			propagators = append(propagators, b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))
		case "jaeger":
			propagators = append(propagators, jaeger.Jaeger{})
			// Add others as needed
		}
	}

	if len(propagators) == 0 {
		// Default to W3C TraceContext + Baggage
		return propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
	}

	return propagation.NewCompositeTextMapPropagator(propagators...)
}

func setNoop() {
	otel.SetTracerProvider(tracenoop.NewTracerProvider())
	otel.SetMeterProvider(metricnoop.NewMeterProvider())
	logglobal.SetLoggerProvider(lognoop.NewLoggerProvider())
}

type extraConfig struct {
	CosmosExtra *cosmosExtra `json:"cosmos_extra" yaml:"cosmos_extra" mapstructure:"cosmos_extra"`
}

// cosmosExtra provides extensions to the OpenTelemetry declarative configuration.
// These options allow features not yet supported by otelconf, such as writing traces/metrics/logs to local
// files, enabling additional host/runtime instrumentation, and configuring custom propagators.
//
// When present in otel.yaml under the `cosmos_extra` key, these fields
// augment/override portions of the OpenTelemetry SDK initialization.
//
// For an example configuration, see the README in this package.
type cosmosExtra struct {
	// TraceFile is an optional path to a file where spans should be exported
	// using the stdouttrace exporter. If empty, no file-based trace export is
	// configured.
	TraceFile string `json:"trace_file" yaml:"trace_file" mapstructure:"trace_file"`

	// MetricsFile is an optional path to a file where metrics should be written
	// using the stdoutmetric exporter. If unset, no file-based metrics export
	// is performed.
	MetricsFile string `json:"metrics_file" yaml:"metrics_file" mapstructure:"metrics_file"`

	// MetricsFileInterval defines how frequently metric data should be flushed
	// to MetricsFile. It must be a valid Go duration string (e.g. "10s",
	// "1m"). If empty, the default PeriodicReader interval is used.
	MetricsFileInterval string `json:"metrics_file_interval" yaml:"metrics_file_interval" mapstructure:"metrics_file_interval"`

	// LogsFile is an optional output file for structured logs exported through
	// the stdoutlog exporter. If unset, log exporting to file is disabled.
	LogsFile string `json:"logs_file" yaml:"logs_file" mapstructure:"logs_file"`

	// InstrumentHost enables collection of host-level metrics such as CPU,
	// memory, and network statistics using the otel host instrumentation.
	InstrumentHost bool `json:"instrument_host" yaml:"instrument_host" mapstructure:"instrument_host"`

	// InstrumentRuntime enables runtime instrumentation that reports Go runtime
	// metrics such as GC activity, heap usage, and goroutine count.
	InstrumentRuntime bool `json:"instrument_runtime" yaml:"instrument_runtime" mapstructure:"instrument_runtime"`

	// Propagators configures additional or alternative TextMapPropagators
	// (e.g. "tracecontext", "baggage", "b3", "b3multi", "jaeger").
	Propagators []string `json:"propagators" yaml:"propagators" mapstructure:"propagators"`
}

func Shutdown(ctx context.Context) error {
	if openTelemetrySDK != nil {
		err := openTelemetrySDK.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown telemetry: %w", err)
		}
		for _, f := range shutdownFuncs {
			if err := f(ctx); err != nil {
				return fmt.Errorf("failed to shutdown telemetry: %w", err)
			}
		}
	}
	return nil
}
