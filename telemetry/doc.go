// Package telemetry provides observability through metrics and distributed tracing.
//
// # Metrics Collection
//
// Metrics collection uses hashicorp/go-metrics with support for multiple sink backends:
//   - mem: In-memory aggregation with SIGUSR1 signal dumping to stderr
//   - prometheus: Prometheus registry for pull-based scraping via /metrics endpoint
//   - statsd: Push-based metrics to StatsD daemon
//   - dogstatsd: Push-based metrics to Datadog StatsD daemon with tagging
//   - file: Write metrics to a file as JSON lines (useful for tests and debugging)
//
// Multiple sinks can be active simultaneously via FanoutSink (e.g., both in-memory and Prometheus).
//
// # Distributed Tracing
//
// Tracing support is provided via OtelSpan, which wraps OpenTelemetry for hierarchical span tracking.
// See otel.go for the log.Tracer implementation.
//
// # Usage
//
// Initialize metrics at application startup:
//
//	m, err := telemetry.New(telemetry.Config{
//		Enabled:                 true,
//		ServiceName:             "cosmos-app",
//		PrometheusRetentionTime: 60,
//		GlobalLabels:            [][]string{{"chain_id", "cosmoshub-1"}},
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer m.Close()
//
// Emit metrics from anywhere in the application:
//
//	telemetry.IncrCounter(1, "tx", "processed")
//	telemetry.SetGauge(1024, "mempool", "size")
//	defer telemetry.MeasureSince(telemetry.Now(), "block", "execution")
package telemetry
