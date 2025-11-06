// Package telemetry initializes OpenTelemetry global using the OpenTelemetry declarative configuration API.
// It also provides some deprecated legacy metrics wrapper functions and metrics configuration using
// github.com/hashicorp/go-metrics.
// By default, this package configures the github.com/hashicorp/go-metrics default instance to
// send all metrics to OpenTelemetry.
package telemetry
