package telemetry

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/hashicorp/go-metrics"
)

// FileSink writes metrics to a file as JSON lines (JSONL format).
// Each metric emission creates a single JSON line with timestamp, type, key, value, and labels.
//
// This sink is particularly useful for:
//   - Test environments where metrics need to be inspected after execution
//   - CI/CD pipelines where metrics should be logged for analysis
//   - Debugging and local development
//
// The sink is thread-safe and buffers writes for performance.
// Call Close() to flush buffered data and close the underlying writer.
type FileSink struct {
	writer *bufio.Writer
	closer io.Closer
	mu     sync.Mutex
	closed bool
}

// metricLine represents a single metric emission in JSON format.
type metricLine struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Key       []string        `json:"key"`
	Value     float32         `json:"value"`
	Labels    []metrics.Label `json:"labels,omitempty"`
}

// NewFileSink creates a new FileSink that writes to the given io.WriteCloser.
// The sink buffers writes for performance. Call Close() to flush and close.
func NewFileSink(w io.WriteCloser) *FileSink {
	return &FileSink{
		writer: bufio.NewWriter(w),
		closer: w,
		closed: false,
	}
}

// SetGauge implements metrics.MetricSink.
func (f *FileSink) SetGauge(key []string, val float32) {
	f.SetGaugeWithLabels(key, val, nil)
}

// SetGaugeWithLabels implements metrics.MetricSink.
func (f *FileSink) SetGaugeWithLabels(key []string, val float32, labels []metrics.Label) {
	f.writeLine(metricLine{
		Timestamp: time.Now().UTC(),
		Type:      "gauge",
		Key:       key,
		Value:     val,
		Labels:    labels,
	})
}

// EmitKey implements metrics.MetricSink.
func (f *FileSink) EmitKey(key []string, val float32) {
	f.writeLine(metricLine{
		Timestamp: time.Now().UTC(),
		Type:      "kv",
		Key:       key,
		Value:     val,
	})
}

// IncrCounter implements metrics.MetricSink.
func (f *FileSink) IncrCounter(key []string, val float32) {
	f.IncrCounterWithLabels(key, val, nil)
}

// IncrCounterWithLabels implements metrics.MetricSink.
func (f *FileSink) IncrCounterWithLabels(key []string, val float32, labels []metrics.Label) {
	f.writeLine(metricLine{
		Timestamp: time.Now().UTC(),
		Type:      "counter",
		Key:       key,
		Value:     val,
		Labels:    labels,
	})
}

// AddSample implements metrics.MetricSink.
func (f *FileSink) AddSample(key []string, val float32) {
	f.AddSampleWithLabels(key, val, nil)
}

// AddSampleWithLabels implements metrics.MetricSink.
func (f *FileSink) AddSampleWithLabels(key []string, val float32, labels []metrics.Label) {
	f.writeLine(metricLine{
		Timestamp: time.Now().UTC(),
		Type:      "sample",
		Key:       key,
		Value:     val,
		Labels:    labels,
	})
}

// writeLine writes a metric line to the file as JSON.
func (f *FileSink) writeLine(line metricLine) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return
	}

	data, err := json.Marshal(line)
	if err != nil {
		// If JSON marshaling fails, write error to stderr but don't crash
		fmt.Fprintf(io.Discard, "failed to marshal metric: %v\n", err)
		return
	}

	// Write JSON line with newline
	if _, err := f.writer.Write(data); err != nil {
		return
	}
	if err := f.writer.WriteByte('\n'); err != nil {
		return
	}
}

// Close flushes any buffered data and closes the underlying writer.
// It is safe to call Close multiple times.
func (f *FileSink) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true

	// Flush buffered data
	if err := f.writer.Flush(); err != nil {
		f.closer.Close() // Try to close anyway
		return fmt.Errorf("failed to flush metrics file: %w", err)
	}

	// Close the underlying file
	if err := f.closer.Close(); err != nil {
		return fmt.Errorf("failed to close metrics file: %w", err)
	}

	return nil
}
