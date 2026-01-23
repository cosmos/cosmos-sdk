// Package diskio provides an implementation of the disk I/O metrics
// following the OpenTelemetry semantic conventions for system metrics
// specified here: https://opentelemetry.io/docs/specs/semconv/system/system-metrics/.
// Under the hood it uses the gopsutil library (github.com/shirou/gopsutil/v4/disk) to
// gather disk I/O statistics.
//
// The metric events produced are listed here with attribute dimensions.
//
//	Name                       Attribute
//
// ----------------------------------------------------------------------
//
//	system.disk.io             system.device, disk.io.direction=read|write
//	system.disk.operations     system.device, disk.io.direction=read|write
//	system.disk.io_time        system.device
//	system.disk.operation_time system.device, disk.io.direction=read|write
//	system.disk.merged         system.device, disk.io.direction=read|write
//
// This package attempts to closely follow the conventions in the
// go.opentelemetry.io/contrib/instrumentation packages.
package diskio
