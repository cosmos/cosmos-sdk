package diskio

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.38.0/systemconv"

	"github.com/cosmos/cosmos-sdk/telemetry/registry"
)

const (
	// Name is the instrument name used in configuration.
	Name = "diskio"
	// OptDisableVirtualDeviceFilter is the option key to disable virtual device filtering.
	OptDisableVirtualDeviceFilter = "disable_virtual_device_filter"
)

func init() {
	registry.Register(instrument{})
}

type instrument struct{}

func (instrument) Name() string { return Name }

func (instrument) Start(cfg map[string]any) error {
	var opts []Option
	if disable, _ := cfg[OptDisableVirtualDeviceFilter].(bool); disable {
		opts = append(opts, WithDisableVirtualDeviceFilter())
	}
	return Start(opts...)
}

// ScopeName is the instrumentation scope name.
const ScopeName = "github.com/cosmos/cosmos-sdk/telemetry/util/diskio"

// Version is the current release version of the disk I/O instrumentation.
func Version() string {
	return "0.1.0"
}

var (
	attrDirectionRead  = systemconv.DiskIO{}.AttrDiskIODirection(systemconv.DiskIODirectionRead)
	attrDirectionWrite = systemconv.DiskIO{}.AttrDiskIODirection(systemconv.DiskIODirectionWrite)
)

// Start initializes reporting of disk I/O metrics using the supplied options.
func Start(opts ...Option) error {
	c := newConfig(opts...)
	meter := c.MeterProvider.Meter(
		ScopeName,
		metric.WithInstrumentationVersion(Version()),
	)

	// systemconv defines all of these counters, but they're not defined
	// as observable counters, so we need to redefine them here reusing
	// the metadata from systemconv.
	diskIO, err := meter.Int64ObservableCounter(
		systemconv.DiskIO{}.Name(),
		metric.WithDescription(systemconv.DiskIO{}.Description()),
		metric.WithUnit(systemconv.DiskIO{}.Unit()),
	)
	if err != nil {
		return err
	}

	diskOps, err := meter.Int64ObservableCounter(
		systemconv.DiskOperations{}.Name(),
		metric.WithDescription(systemconv.DiskOperations{}.Description()),
		metric.WithUnit(systemconv.DiskOperations{}.Unit()),
	)
	if err != nil {
		return err
	}

	diskIOTime, err := meter.Float64ObservableCounter(
		systemconv.DiskIOTime{}.Name(),
		metric.WithDescription(systemconv.DiskIOTime{}.Description()),
		metric.WithUnit(systemconv.DiskIOTime{}.Unit()),
	)
	if err != nil {
		return err
	}

	diskOpTime, err := meter.Float64ObservableCounter(
		systemconv.DiskOperationTime{}.Name(),
		metric.WithDescription(systemconv.DiskOperationTime{}.Description()),
		metric.WithUnit(systemconv.DiskOperationTime{}.Unit()),
	)
	if err != nil {
		return err
	}

	diskMerged, err := meter.Int64ObservableCounter(
		systemconv.DiskMerged{}.Name(),
		metric.WithDescription(systemconv.DiskMerged{}.Description()),
		metric.WithUnit(systemconv.DiskMerged{}.Unit()),
	)
	if err != nil {
		return err
	}

	collector := newCollector(c.MinimumReadInterval)
	var lock sync.Mutex

	_, err = meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			lock.Lock()
			defer lock.Unlock()

			stats := collector.refresh(ctx)
			if stats == nil {
				return nil
			}

			for device, s := range stats {
				if !c.DisableVirtualDeviceFilter && !isPhysicalDevice(device) {
					continue
				}
				attrDevice := systemconv.DiskIO{}.AttrDevice(device)

				// system.disk.io (bytes read/written)
				o.ObserveInt64(diskIO, int64(s.ReadBytes),
					metric.WithAttributes(attrDevice, attrDirectionRead))
				o.ObserveInt64(diskIO, int64(s.WriteBytes),
					metric.WithAttributes(attrDevice, attrDirectionWrite))

				// system.disk.operations (read/write counts)
				o.ObserveInt64(diskOps, int64(s.ReadCount),
					metric.WithAttributes(attrDevice, attrDirectionRead))
				o.ObserveInt64(diskOps, int64(s.WriteCount),
					metric.WithAttributes(attrDevice, attrDirectionWrite))

				// system.disk.io_time (seconds) - gopsutil returns milliseconds
				o.ObserveFloat64(diskIOTime, float64(s.IoTime)/1000.0,
					metric.WithAttributes(attrDevice))

				// system.disk.operation_time (seconds) - gopsutil returns milliseconds
				o.ObserveFloat64(diskOpTime, float64(s.ReadTime)/1000.0,
					metric.WithAttributes(attrDevice, attrDirectionRead))
				o.ObserveFloat64(diskOpTime, float64(s.WriteTime)/1000.0,
					metric.WithAttributes(attrDevice, attrDirectionWrite))

				// system.disk.merged (merged read/write operations)
				o.ObserveInt64(diskMerged, int64(s.MergedReadCount),
					metric.WithAttributes(attrDevice, attrDirectionRead))
				o.ObserveInt64(diskMerged, int64(s.MergedWriteCount),
					metric.WithAttributes(attrDevice, attrDirectionWrite))
			}
			return nil
		},
		diskIO,
		diskOps,
		diskIOTime,
		diskOpTime,
		diskMerged,
	)
	return err
}

type ioCollector struct {
	// now is used to replace the implementation of time.Now for testing
	now func() time.Time
	// lastCollect tracks the last time metrics were refreshed
	lastCollect time.Time
	// minimumInterval is the minimum amount of time between calls to metrics.Read
	minimumInterval time.Duration
	// lastStats holds the last collected disk I/O stats
	lastStats map[string]disk.IOCountersStat
}

func newCollector(minimumInterval time.Duration) *ioCollector {
	return &ioCollector{
		minimumInterval: minimumInterval,
		now:             time.Now,
	}
}

func (c *ioCollector) refresh(ctx context.Context) map[string]disk.IOCountersStat {
	now := c.now()
	if now.Sub(c.lastCollect) < c.minimumInterval && c.lastStats != nil {
		return c.lastStats
	}

	stats, err := disk.IOCountersWithContext(ctx)
	if err != nil {
		// on error, return last stats if available
		return c.lastStats
	}

	c.lastStats = stats
	c.lastCollect = now
	return stats
}

// isPhysicalDevice determines which devices to report based on the OS.
// On Linux, it strictly filters out partitions, loopbacks, and software RAID
// to prevent double counting. On other OS's, it reports everything.
func isPhysicalDevice(deviceName string) bool {
	if runtime.GOOS == "linux" {
		return isPhysicalDeviceLinux(deviceName)
	}
	// TODO: maybe add filters for macos? windows?

	return true
}

// isPhysicalDeviceLinux out loopback and RAID devices, and anything that reports as partition.
// We filter out loopback and RAID devices as these are virtual storage blocks, and writes to these will
// incur writes to a physical storage device. By filtering we avoid a double count of write pressure to physical disks.
func isPhysicalDeviceLinux(deviceName string) bool {
	// filter out virtual devices. (i.e. loop, md)
	virtualPath := filepath.Join("/sys/devices/virtual/block", deviceName)
	_, err := os.Stat(virtualPath)
	// file exists, its virtual. not physical.
	if err == nil {
		return false
	}

	// check sysfs to distinguish Physical Disks from Partitions.
	// /sys/class/block/<device>/partition exists ONLY for partitions.
	sysPath := filepath.Join("/sys/class/block", deviceName, "partition")
	_, err = os.Stat(sysPath)
	// file exists, its a partition. not physical.
	if err == nil {
		return false
	}

	// probably a physical disk by this point.
	return true
}
