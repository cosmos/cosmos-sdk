// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

// globalLabels defines the set of global labels that will be applied to all
// metrics emitted using the telemetry package function wrappers.
// var globalLabels = []gometrics.Label{}

// Metrics supported format types.
const (
	FormatDefault    = ""
	FormatPrometheus = "prometheus"
	FormatText       = "text"
)

// Common metric key constants
const (
	MetricKeyBeginBlocker = "begin_blocker"
	MetricKeyEndBlocker   = "end_blocker"
	MetricLabelNameModule = "module"
)
