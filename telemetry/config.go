package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/otelconf/v0.3.0"
)

var sdk otelconf.SDK

var isTelemetryEnabled = true

func init() {
	var err error

	var opts []otelconf.ConfigurationOption

	confFilename := os.Getenv("COSMOS_TELEMETRY")
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
