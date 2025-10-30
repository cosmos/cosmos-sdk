package telemetry

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hashicorp/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestFileSink_BasicOperations(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "metrics.jsonl")
	file, err := os.Create(tmpfile)
	require.NoError(t, err)

	sink := NewFileSink(file)

	// Emit various metric types
	sink.IncrCounter([]string{"test", "counter"}, 1.5)
	sink.SetGauge([]string{"test", "gauge"}, 42.0)
	sink.AddSample([]string{"test", "sample"}, 100.0)
	sink.EmitKey([]string{"test", "kv"}, 3.14)

	// Emit metrics with labels
	labels := []metrics.Label{
		{Name: "module", Value: "bank"},
		{Name: "operation", Value: "send"},
	}
	sink.IncrCounterWithLabels([]string{"test", "counter_labeled"}, 2.0, labels)
	sink.SetGaugeWithLabels([]string{"test", "gauge_labeled"}, 99.0, labels)
	sink.AddSampleWithLabels([]string{"test", "sample_labeled"}, 50.0, labels)

	// Close to flush
	require.NoError(t, sink.Close())

	// Read and verify file contents
	data, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	file2, err := os.Open(tmpfile)
	require.NoError(t, err)
	defer file2.Close()

	scanner := bufio.NewScanner(file2)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
		var metric metricLine
		err := json.Unmarshal(scanner.Bytes(), &metric)
		require.NoError(t, err, "line %d should be valid JSON", lineCount)
		require.NotZero(t, metric.Timestamp, "line %d should have timestamp", lineCount)
		require.NotEmpty(t, metric.Type, "line %d should have type", lineCount)
		require.NotEmpty(t, metric.Key, "line %d should have key", lineCount)
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, 7, lineCount, "should have 7 metrics")

	// Verify specific metric formats
	require.Contains(t, string(data), `"type":"counter"`)
	require.Contains(t, string(data), `"type":"gauge"`)
	require.Contains(t, string(data), `"type":"sample"`)
	require.Contains(t, string(data), `"type":"kv"`)
	require.Contains(t, string(data), `"key":["test","counter"]`)
	require.Contains(t, string(data), `"value":1.5`)
	require.Contains(t, string(data), `"labels":[{"Name":"module","Value":"bank"}`)
}

func TestFileSink_ConcurrentWrites(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "metrics_concurrent.jsonl")
	file, err := os.Create(tmpfile)
	require.NoError(t, err)

	sink := NewFileSink(file)

	// Spawn multiple goroutines writing metrics concurrently
	const numGoroutines = 10
	const metricsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < metricsPerGoroutine; j++ {
				sink.IncrCounter([]string{"concurrent", "test"}, float32(id))
			}
		}(i)
	}

	wg.Wait()
	require.NoError(t, sink.Close())

	// Count lines in file
	file, err = os.Open(tmpfile)
	require.NoError(t, err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, numGoroutines*metricsPerGoroutine, lineCount, "all metrics should be written")
}

func TestFileSink_CloseIdempotent(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "metrics_close.jsonl")
	file, err := os.Create(tmpfile)
	require.NoError(t, err)

	sink := NewFileSink(file)
	sink.IncrCounter([]string{"test"}, 1.0)

	// Close multiple times should not error
	require.NoError(t, sink.Close())
	require.NoError(t, sink.Close())
	require.NoError(t, sink.Close())
}

func TestFileSink_WritesAfterClose(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "metrics_after_close.jsonl")
	file, err := os.Create(tmpfile)
	require.NoError(t, err)

	sink := NewFileSink(file)
	sink.IncrCounter([]string{"before"}, 1.0)
	require.NoError(t, sink.Close())

	// Writes after close should be silently ignored (no panic)
	sink.IncrCounter([]string{"after"}, 1.0)

	// File should only contain one metric
	data, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	file3, err := os.Open(tmpfile)
	require.NoError(t, err)
	defer file3.Close()

	scanner := bufio.NewScanner(file3)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}
	require.NoError(t, scanner.Err())
	require.Equal(t, 1, lineCount, "only metric before close should be written")
	require.Contains(t, string(data), `"key":["before"]`)
	require.NotContains(t, string(data), `"key":["after"]`)
}

func TestFileSink_JSONFormat(t *testing.T) {
	tmpfile := filepath.Join(t.TempDir(), "metrics_json.jsonl")
	file, err := os.Create(tmpfile)
	require.NoError(t, err)

	sink := NewFileSink(file)
	labels := []metrics.Label{{Name: "env", Value: "test"}}
	sink.IncrCounterWithLabels([]string{"api", "requests"}, 5.0, labels)
	require.NoError(t, sink.Close())

	// Parse JSON and verify structure
	data, err := os.ReadFile(tmpfile)
	require.NoError(t, err)

	var metric metricLine
	err = json.Unmarshal(data[:len(data)-1], &metric) // Remove trailing newline
	require.NoError(t, err)

	require.Equal(t, "counter", metric.Type)
	require.Equal(t, []string{"api", "requests"}, metric.Key)
	require.Equal(t, float32(5.0), metric.Value)
	require.Len(t, metric.Labels, 1)
	require.Equal(t, "env", metric.Labels[0].Name)
	require.Equal(t, "test", metric.Labels[0].Value)
	require.NotZero(t, metric.Timestamp)
}
