package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewMetrics(t *testing.T) {
	testCases := []struct {
		name   string
		labels [][]string
	}{
		{
			name:   "no labels",
			labels: nil,
		},
		{
			name: "single label",
			labels: [][]string{
				{"key1", "value1"},
			},
		},
		{
			name: "multiple labels",
			labels: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
			},
		},
		{
			name: "empty label values",
			labels: [][]string{
				{"key1", ""},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewMetrics(tc.labels)
			require.NotNil(t, m)

			if tc.labels == nil {
				require.Nil(t, m.Labels)
			} else {
				require.NotNil(t, m.Labels)
				require.Len(t, m.Labels, len(tc.labels))
				for i, label := range tc.labels {
					require.Equal(t, label[0], m.Labels[i].Name)
					require.Equal(t, label[1], m.Labels[i].Value)
				}
			}
		})
	}
}

func TestMeasureSince(t *testing.T) {
	// Test with actual metrics implementation
	t.Run("with metrics", func(t *testing.T) {
		m := NewMetrics([][]string{{"test", "value"}})
		time.Sleep(time.Millisecond) // Ensure some time passes
		m.MeasureSince("test", "metric")
		// Note: We can't easily verify the actual measurement since it's using the global metrics sink
		// But we can verify it doesn't panic
	})

	// Test with NoOp implementation
	t.Run("with noop metrics", func(t *testing.T) {
		noOp := NewNoOpMetrics()
		time.Sleep(time.Millisecond)
		noOp.MeasureSince("test", "metric")
		// Should not panic
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty label pairs", func(t *testing.T) {
		m := NewMetrics([][]string{})
		require.NotNil(t, m)
		require.Empty(t, m.Labels)
		m.MeasureSince("test", "metric") // Should not panic
	})

	t.Run("nil labels", func(t *testing.T) {
		m := NewMetrics(nil)
		require.NotNil(t, m)
		require.Empty(t, m.Labels)
		m.MeasureSince("test", "metric") // Should not panic
	})

	t.Run("empty metric name", func(t *testing.T) {
		m := NewMetrics([][]string{{"test", "value"}})
		m.MeasureSince("", "metric") // Should not panic
	})

	t.Run("empty key", func(t *testing.T) {
		m := NewMetrics([][]string{{"", "value"}})
		m.MeasureSince("test", "metric") // Should not panic
	})
}
