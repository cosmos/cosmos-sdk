package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hashicorp/go-metrics"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/otelconf/v0.3.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	logglobal "go.opentelemetry.io/otel/log/global"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.yaml.in/yaml/v3"

	"cosmossdk.io/log"
)

var sdk otelconf.SDK
var shutdownFuncs []func(context.Context) error

func init() {
	err := doInit()
	if err != nil {
		panic(err)
	}
}

func doInit() error {
	// if otel is already marked as configured, skip
	if log.IsOpenTelemetryConfigured() {
		return nil
	}

	var err error

	var opts []otelconf.ConfigurationOption

	confFilename := os.Getenv("OTEL_EXPERIMENTAL_CONFIG_FILE")
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

	cfgJson, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry config file: %w", err)
	}
	fmt.Printf("\nInitializing telemetry with config:\n%s\n\n", cfgJson)

	opts = append(opts, otelconf.WithOpenTelemetryConfiguration(*cfg))

	// parse cosmos extra config
	var extraCfg extraConfig
	err = yaml.Unmarshal(bz, &extraCfg)
	if err == nil {
		if extraCfg.CosmosExtra != nil {
			extra := *extraCfg.CosmosExtra
			if extra.TraceFile != "" {
				fmt.Printf("Initializing trace file: %s\n", extra.TraceFile)
				traceFile, err := os.OpenFile(extra.TraceFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
					//stdouttrace.WithPrettyPrint(),
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
				metricsFile, err := os.OpenFile(extra.MetricsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
					//stdoutmetric.WithPrettyPrint(),
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
				logsFile, err := os.OpenFile(extra.LogsFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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
					//stdoutlog.WithPrettyPrint(),
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
		}
	} else {
		fmt.Printf("failed to parse cosmos extra config: %v\n", err)
	}

	sdk, err = otelconf.NewSDK(opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// setup otel global providers
	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	logglobal.SetLoggerProvider(sdk.LoggerProvider())
	// setup slog default provider so that any logs emitted the default slog will be traced
	slog.SetDefault(otelslog.NewLogger("", otelslog.WithSource(true)))
	// mark otel as configured in the log package
	log.SetOpenTelemetryConfigured(true)
	// emit an initialized message which verifies basic telemetry is working
	slog.Info("Telemetry initialized")

	// extract service name from config for go-metrics compatibility layer
	serviceName := "cosmos-sdk" // default fallback
	if cfg.Resource != nil && cfg.Resource.Attributes != nil {
		for _, attr := range cfg.Resource.Attributes {
			if attr.Name == "service.name" {
				if svcNameStr, ok := attr.Value.(string); ok {
					serviceName = svcNameStr
				}
				break
			}
		}
	}

	// setup go-metrics compatibility layer
	_, err = metrics.NewGlobal(metrics.DefaultConfig(serviceName), newOtelGoMetricsSink(
		context.Background(),
		sdk.MeterProvider().Meter("gometrics"),
	))
	if err != nil {
		return fmt.Errorf("failed to initialize go-metrics compatibility layer: %w", err)
	}
	globalTelemetryEnabled = true

	return nil
}

type extraConfig struct {
	CosmosExtra *cosmosExtra `json:"cosmos_extra" yaml:"cosmos_extra" mapstructure:"cosmos_extra"`
}

type cosmosExtra struct {
	TraceFile           string `json:"trace_file" yaml:"trace_file" mapstructure:"trace_file"`
	MetricsFile         string `json:"metrics_file" yaml:"metrics_file" mapstructure:"metrics_file"`
	MetricsFileInterval string `json:"metrics_file_interval" yaml:"metrics_file_interval" mapstructure:"metrics_file_interval"`
	LogsFile            string `json:"logs_file" yaml:"logs_file" mapstructure:"logs_file"`
	InstrumentHost      bool   `json:"instrument_host" yaml:"instrument_host" mapstructure:"instrument_host"`
	InstrumentRuntime   bool   `json:"instrument_runtime" yaml:"instrument_runtime" mapstructure:"instrument_runtime"`
}

func Shutdown(ctx context.Context) error {
	err := sdk.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("failed to shutdown telemetry: %w", err)
	}
	for _, f := range shutdownFuncs {
		if err := f(ctx); err != nil {
			return fmt.Errorf("failed to shutdown telemetry: %w", err)
		}
	}
	return nil
}
