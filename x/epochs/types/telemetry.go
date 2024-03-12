package types

// EpochHookFailedMetricName
// epoch_hook_failed
//
// counter that is increased if epoch hook fails
//
// Has the following labels:
// * module_name - the name of the module that errored or panicked
// * err - the error or panic returned
// * is_before_hook - true if this is a before epoch hook. False otherwise.
var EpochHookFailedMetricName = formatEpochMetricName("hook_failed")

// FormatMetricName helper to format a metric name given SDK module name and extension.
func FormatMetricName(moduleName, extension string) string {
	return moduleName + "_" + extension
}

// formatTxFeesMetricName formats the epochs module metric name.
func formatEpochMetricName(metricName string) string {
	return FormatMetricName(ModuleName, metricName)
}
