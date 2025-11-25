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
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.yaml.in/yaml/v3"
)

const (
	OtelConfigEnvVar = "OTEL_EXPERIMENTAL_CONFIG_FILE"
)

var (
	sdk           *otelconf.SDK
	shutdownFuncs []func(context.Context) error
)

func init() {
	err := initOpenTelemetry()
	if err != nil {
		panic(err)
	}
}

func initOpenTelemetry() error {
	var err error

	var opts []otelconf.ConfigurationOption

	confFilename := os.Getenv(OtelConfigEnvVar)
	if confFilename == "" {
		return nil
	}

	bz, err := os.ReadFile(confFilename)
	if err != nil {
		return fmt.Errorf("failed to read telemetry config file: %w", err)
	}

	cfg, err := otelconf.ParseYAML(bz)
	if err != nil {
		return fmt.Errorf("failed to parse telemetry config file: %w", err)
	}

	fmt.Printf("\nInitializing OpenTelemetry\n")

	opts = append(opts, otelconf.WithOpenTelemetryConfiguration(*cfg))

	// parse cosmos extra config
	var extraCfg extraConfig
	err = yaml.Unmarshal(bz, &extraCfg)
	if err == nil {
		if extraCfg.CosmosExtra != nil {
			extra := *extraCfg.CosmosExtra
			if extra.TraceFile != "" {
				fmt.Printf("Initializing trace file: %s\n", extra.TraceFile)
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
				fmt.Printf("Initializing metrics file: %s\n", extra.MetricsFile)
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
					fmt.Printf("Configuring metrics export interval: %v\n", interval)
					readerOpts = append(readerOpts, metricsdk.WithInterval(interval))
				}

				opts = append(opts, otelconf.WithMeterProviderOptions(
					metricsdk.WithReader(metricsdk.NewPeriodicReader(exporter, readerOpts...)),
				))
			}
			if extra.LogsFile != "" {
				fmt.Printf("Initializing logs file: %s\n", extra.LogsFile)
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
	} else {
		fmt.Printf("failed to parse cosmos extra config: %v\n", err)
	}

	otelSDK, err := otelconf.NewSDK(opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}
	sdk = &otelSDK

	// setup otel global providers
	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	logglobal.SetLoggerProvider(sdk.LoggerProvider())
	fmt.Printf("\nOpenTelemetry initialized successfully\n")

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

type extraConfig struct {
	CosmosExtra *cosmosExtra `json:"cosmos_extra" yaml:"cosmos_extra" mapstructure:"cosmos_extra"`
}

type cosmosExtra struct {
	TraceFile           string   `json:"trace_file" yaml:"trace_file" mapstructure:"trace_file"`
	MetricsFile         string   `json:"metrics_file" yaml:"metrics_file" mapstructure:"metrics_file"`
	MetricsFileInterval string   `json:"metrics_file_interval" yaml:"metrics_file_interval" mapstructure:"metrics_file_interval"`
	LogsFile            string   `json:"logs_file" yaml:"logs_file" mapstructure:"logs_file"`
	InstrumentHost      bool     `json:"instrument_host" yaml:"instrument_host" mapstructure:"instrument_host"`
	InstrumentRuntime   bool     `json:"instrument_runtime" yaml:"instrument_runtime" mapstructure:"instrument_runtime"`
	Propagators         []string `json:"propagators" yaml:"propagators" mapstructure:"propagators"`
}

func Shutdown(ctx context.Context) error {
	if sdk != nil {
		err := sdk.Shutdown(ctx)
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
