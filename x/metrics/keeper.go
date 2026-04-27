package metrics

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Keeper manages metric collection and storage.
type Keeper struct {
	config   MetricsConfig
	registry *MetricRegistry
	mu       sync.RWMutex

	txCounts  map[string]float64
	gasUsed   map[string]float64
	blockTime float64
	moduleCalls map[string]float64
	moduleDurations map[string]float64
}

// NewKeeper creates a new Keeper with the provided configuration.
func NewKeeper(config MetricsConfig) *Keeper {
	return &Keeper{
		config:          config,
		registry:        NewMetricRegistry(),
		txCounts:        make(map[string]float64),
		gasUsed:         make(map[string]float64),
		moduleCalls:     make(map[string]float64),
		moduleDurations: make(map[string]float64),
	}
}

// RecordTransaction tracks transaction counts and gas usage by type.
func (k *Keeper) RecordTransaction(txType string, gasUsed int64) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.txCounts[txType]++
	k.gasUsed[txType] += float64(gasUsed)
}

// RecordBlockTime tracks the duration of block processing.
func (k *Keeper) RecordBlockTime(duration time.Duration) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.blockTime = duration.Seconds()
}

// RecordModuleCall tracks module method call count and duration.
func (k *Keeper) RecordModuleCall(module string, method string, duration time.Duration) {
	k.mu.Lock()
	defer k.mu.Unlock()
	key := module + "." + method
	k.moduleCalls[key]++
	k.moduleDurations[key] += duration.Seconds()
}

// GetMetrics returns all collected metrics as a slice.
func (k *Keeper) GetMetrics() []Metric {
	k.mu.RLock()
	defer k.mu.RUnlock()

	var metrics []Metric
	now := time.Now()

	for txType, count := range k.txCounts {
		metrics = append(metrics, k.newMetric("tx_count", Counter, count, map[string]string{"type": txType}, now))
	}
	for txType, gas := range k.gasUsed {
		metrics = append(metrics, k.newMetric("tx_gas_used", Counter, gas, map[string]string{"type": txType}, now))
	}
	if k.blockTime > 0 {
		metrics = append(metrics, k.newMetric("block_time_seconds", Gauge, k.blockTime, nil, now))
	}
	for key, count := range k.moduleCalls {
		parts := strings.SplitN(key, ".", 2)
		labels := map[string]string{"module": parts[0], "method": parts[1]}
		metrics = append(metrics, k.newMetric("module_call_count", Counter, count, labels, now))
	}
	for key, dur := range k.moduleDurations {
		parts := strings.SplitN(key, ".", 2)
		labels := map[string]string{"module": parts[0], "method": parts[1]}
		metrics = append(metrics, k.newMetric("module_call_duration_seconds", Counter, dur, labels, now))
	}

	// Include metrics from registered collectors.
	metrics = append(metrics, k.registry.CollectAll()...)
	return metrics
}

// FormatPrometheus returns all metrics in Prometheus text exposition format.
func (k *Keeper) FormatPrometheus() string {
	metrics := k.GetMetrics()
	var sb strings.Builder
	for _, m := range metrics {
		fqName := fmt.Sprintf("%s_%s_%s", k.config.Namespace, k.config.Subsystem, m.Name)
		sb.WriteString(fmt.Sprintf("# TYPE %s %s\n", fqName, m.Type.String()))
		sb.WriteString(fqName)
		if len(m.Labels) > 0 {
			sb.WriteString("{")
			first := true
			for lk, lv := range m.Labels {
				if !first {
					sb.WriteString(",")
				}
				sb.WriteString(fmt.Sprintf(`%s="%s"`, lk, lv))
				first = false
			}
			sb.WriteString("}")
		}
		sb.WriteString(fmt.Sprintf(" %g\n", m.Value))
	}
	return sb.String()
}

// Reset clears all collected metrics.
func (k *Keeper) Reset() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.txCounts = make(map[string]float64)
	k.gasUsed = make(map[string]float64)
	k.blockTime = 0
	k.moduleCalls = make(map[string]float64)
	k.moduleDurations = make(map[string]float64)
}

// newMetric is a helper to create a Metric with the keeper's config applied.
func (k *Keeper) newMetric(name string, typ MetricType, value float64, labels map[string]string, ts time.Time) Metric {
	return Metric{
		Name:      name,
		Type:      typ,
		Value:     value,
		Labels:    labels,
		Timestamp: ts,
	}
}
