package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestMetrics_Disabled(t *testing.T) {
	m, err := New(Config{Enabled: false})
	require.Nil(t, m)
	require.Nil(t, err)
}

func TestMetrics_InMem(t *testing.T) {
	m, err := New(Config{
		MetricsSink:    MetricSinkInMem,
		Enabled:        true,
		EnableHostname: false,
		ServiceName:    "test",
	})
	require.NoError(t, err)
	require.NotNil(t, m)

	emitMetrics()

	gr, err := m.Gather(FormatText)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, "application/json")

	jsonMetrics := make(map[string]any)
	require.NoError(t, json.Unmarshal(gr.Metrics, &jsonMetrics))

	counters := jsonMetrics["Counters"].([]any)
	require.Equal(t, counters[0].(map[string]any)["Count"].(float64), 10.0)
	require.Equal(t, counters[0].(map[string]any)["Name"].(string), "test.dummy_counter")
}

func TestMetrics_Prom(t *testing.T) {
	m, err := New(Config{
		MetricsSink:             MetricSinkInMem,
		Enabled:                 true,
		EnableHostname:          false,
		ServiceName:             "test",
		PrometheusRetentionTime: 60,
		EnableHostnameLabel:     false,
	})
	require.NoError(t, err)
	require.NotNil(t, m)
	require.True(t, m.prometheusEnabled)

	emitMetrics()

	gr, err := m.Gather(FormatPrometheus)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, string(ContentTypeText))

	require.True(t, strings.Contains(string(gr.Metrics), "test_dummy_counter 30"))
}

func emitMetrics() {
	ticker := time.NewTicker(time.Second)
	timeout := time.After(30 * time.Second)

	for {
		select {
		case <-ticker.C:
			metrics.IncrCounter([]string{"dummy_counter"}, 1.0)
		case <-timeout:
			return
		}
	}
}

func TestOtlpConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     OtlpConfig
		wantErr bool
	}{
		{
			name: "exporter disabled → no error",
			cfg: OtlpConfig{
				ExporterEnabled: false,
				// other fields may be empty
			},
			wantErr: false,
		},
		{
			name: "exporter enabled but missing endpoint → error",
			cfg: OtlpConfig{
				ExporterEnabled: true,
				// CollectorEndpoint is empty
				User:         "user",
				Token:        "token",
				PushInterval: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "exporter enabled but missing user/token → error",
			cfg: OtlpConfig{
				ExporterEnabled:   true,
				CollectorEndpoint: "http://example.com:4318",
				// User or OtlpToken is empty
				User:         "",
				Token:        "",
				PushInterval: 10 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "exporter enabled but zero interval → error",
			cfg: OtlpConfig{
				ExporterEnabled:   true,
				CollectorEndpoint: "http://example.com:4318",
				User:              "user",
				Token:             "token",
				PushInterval:      0, // invalid
			},
			wantErr: true,
		},
		{
			name: "exporter enabled with negative interval → error",
			cfg: OtlpConfig{
				ExporterEnabled:   true,
				CollectorEndpoint: "http://example.com:4318",
				User:              "user",
				Token:             "token",
				PushInterval:      -5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "exporter enabled with all fields set correctly → no error",
			cfg: OtlpConfig{
				ExporterEnabled:   true,
				CollectorEndpoint: "https://collector.example.com:4318",
				User:              "user123",
				Token:             "tokenABC",
				ServiceName:       "my-service",
				PushInterval:      15 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
