package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestCosmosExtraUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected extraConfig
	}{
		{
			name: "instruments with empty config",
			yaml: `
cosmos_extra:
  instruments:
    host: {}
    runtime: {}
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{
					Instruments: map[string]map[string]any{
						"host":    {},
						"runtime": {},
					},
				},
			},
		},
		{
			name: "instruments with options",
			yaml: `
cosmos_extra:
  instruments:
    host: {}
    diskio:
      disable_virtual_device_filter: true
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{
					Instruments: map[string]map[string]any{
						"host": {},
						"diskio": {
							"disable_virtual_device_filter": true,
						},
					},
				},
			},
		},
		{
			name: "instruments with propagators",
			yaml: `
cosmos_extra:
  instruments:
    host: {}
  propagators:
    - tracecontext
    - baggage
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{
					Instruments: map[string]map[string]any{
						"host": {},
					},
					Propagators: []string{"tracecontext", "baggage"},
				},
			},
		},
		{
			name: "full config",
			yaml: `
cosmos_extra:
  trace_file: /tmp/traces.json
  metrics_file: /tmp/metrics.json
  instruments:
    host: {}
    runtime: {}
    diskio:
      disable_virtual_device_filter: true
  propagators:
    - tracecontext
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{
					TraceFile:   "/tmp/traces.json",
					MetricsFile: "/tmp/metrics.json",
					Instruments: map[string]map[string]any{
						"host":    {},
						"runtime": {},
						"diskio": {
							"disable_virtual_device_filter": true,
						},
					},
					Propagators: []string{"tracecontext"},
				},
			},
		},
		{
			name: "empty cosmos_extra (null)",
			yaml: `
cosmos_extra:
`,
			expected: extraConfig{
				CosmosExtra: nil, // YAML with just "cosmos_extra:" and no value results in nil
			},
		},
		{
			name: "empty cosmos_extra (empty object)",
			yaml: `
cosmos_extra: {}
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{},
			},
		},
		{
			name: "no cosmos_extra",
			yaml: `
some_other_key: value
`,
			expected: extraConfig{
				CosmosExtra: nil,
			},
		},
		{
			name: "realistic otel.yaml with cosmos_extra",
			yaml: `
file_format: "1.0-rc.3"
resource:
  attributes:
    - name: service.name
      value: simapp

tracer_provider:
  processors:
    - batch:
        exporter:
          otlp_grpc:
            endpoint: http://localhost:4317

meter_provider:
  readers:
    - pull:
        exporter:
          prometheus/development:
            host: 0.0.0.0
            port: 9464

cosmos_extra:
  instruments:
    host: {}
    runtime: {}
    diskio:
      disable_virtual_device_filter: true
  propagators:
    - tracecontext
`,
			expected: extraConfig{
				CosmosExtra: &cosmosExtra{
					Instruments: map[string]map[string]any{
						"host":    {},
						"runtime": {},
						"diskio": {
							"disable_virtual_device_filter": true,
						},
					},
					Propagators: []string{"tracecontext"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cfg extraConfig
			err := yaml.Unmarshal([]byte(tc.yaml), &cfg)
			require.NoError(t, err)

			if tc.expected.CosmosExtra == nil {
				require.Nil(t, cfg.CosmosExtra)
				return
			}

			require.NotNil(t, cfg.CosmosExtra)
			require.Equal(t, tc.expected.CosmosExtra.TraceFile, cfg.CosmosExtra.TraceFile)
			require.Equal(t, tc.expected.CosmosExtra.MetricsFile, cfg.CosmosExtra.MetricsFile)
			require.Equal(t, tc.expected.CosmosExtra.Propagators, cfg.CosmosExtra.Propagators)

			require.Equal(t, len(tc.expected.CosmosExtra.Instruments), len(cfg.CosmosExtra.Instruments),
				"instruments count mismatch")

			for name, expectedOpts := range tc.expected.CosmosExtra.Instruments {
				actualOpts, ok := cfg.CosmosExtra.Instruments[name]
				require.True(t, ok, "missing instrument: %s", name)
				require.Equal(t, len(expectedOpts), len(actualOpts),
					"options count mismatch for instrument %s", name)

				for key, expectedVal := range expectedOpts {
					actualVal, ok := actualOpts[key]
					require.True(t, ok, "missing option %s for instrument %s", key, name)
					require.Equal(t, expectedVal, actualVal,
						"option value mismatch for %s.%s", name, key)
				}
			}
		})
	}
}
