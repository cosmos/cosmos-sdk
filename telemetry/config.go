package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/otelconf/v0.3.0"
	"go.opentelemetry.io/otel"
	logglobal "go.opentelemetry.io/otel/log/global"
)

var sdk otelconf.SDK

var isTelemetryEnabled = true

func init() {
	var err error

	var opts []otelconf.ConfigurationOption

	confFilename := os.Getenv("OTEL_EXPERIMENTAL_CONFIG_FILE")
	if confFilename != "" {
		bz, err := os.ReadFile(confFilename)
		if err != nil {
			panic(fmt.Sprintf("failed to read telemetry config file: %v", err))
		}

		cfg, err := otelconf.ParseYAML(bz)
		if err != nil {
			panic(fmt.Sprintf("failed to parse telemetry config file: %v", err))
		}

		cfgJson, err := json.Marshal(cfg)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal telemetry config file: %v", err))
		}
		fmt.Printf("\nInitializing telemetry with config:\n%s\n\n", cfgJson)

		opts = append(opts, otelconf.WithOpenTelemetryConfiguration(*cfg))
	}

	sdk, err = otelconf.NewSDK(opts...)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize telemetry: %v", err))
	}

	// setup otel global providers
	otel.SetTracerProvider(sdk.TracerProvider())
	otel.SetMeterProvider(sdk.MeterProvider())
	logglobal.SetLoggerProvider(sdk.LoggerProvider())
	// setup slog default provider so that any logs emitted the default slog will be traced
	slog.SetDefault(otelslog.NewLogger("", otelslog.WithSource(true)))
	// emit an initialized message which verifies basic telemetry is working
	slog.Info("Telemetry initialized")
}

func Shutdown(ctx context.Context) error {
	return sdk.Shutdown(ctx)
}

func IsTelemetryEnabled() bool {
	return isTelemetryEnabled
}

func SetTelemetryEnabled(v bool) {
	isTelemetryEnabled = v
}
