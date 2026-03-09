package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
)

func TestExtensionOptionsUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected *ExtensionOptions
	}{
		{
			name: "instruments with empty config",
			yaml: `
extensions:
  instruments:
    host: {}
    runtime: {}
`,
			expected: &ExtensionOptions{
				Instruments: map[string]map[string]any{
					"host":    {},
					"runtime": {},
				},
			},
		},
		{
			name: "instruments with options",
			yaml: `
extensions:
  instruments:
    host: {}
    diskio:
      disable_virtual_device_filter: true
`,
			expected: &ExtensionOptions{
				Instruments: map[string]map[string]any{
					"host": {},
					"diskio": {
						"disable_virtual_device_filter": true,
					},
				},
			},
		},
		{
			name: "instruments with propagators",
			yaml: `
extensions:
  instruments:
    host: {}
  propagators:
    - tracecontext
    - baggage
`,
			expected: &ExtensionOptions{
				Instruments: map[string]map[string]any{
					"host": {},
				},
				Propagators: []string{"tracecontext", "baggage"},
			},
		},
		{
			name: "full config",
			yaml: `
extensions:
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
			expected: &ExtensionOptions{
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
		{
			name: "empty extensions (null)",
			yaml: `
extensions:
`,
			expected: nil,
		},
		{
			name: "empty extensions (empty object)",
			yaml: `
extensions: {}
`,
			expected: &ExtensionOptions{},
		},
		{
			name: "no extensions",
			yaml: `
some_other_key: value
`,
			expected: nil,
		},
		{
			name: "realistic otel.yaml with extensions",
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

extensions:
  instruments:
    host: {}
    runtime: {}
    diskio:
      disable_virtual_device_filter: true
  propagators:
    - tracecontext
`,
			expected: &ExtensionOptions{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cfg struct {
				Extensions *ExtensionOptions `yaml:"extensions"`
			}
			err := yaml.Unmarshal([]byte(tc.yaml), &cfg)
			require.NoError(t, err)

			if tc.expected == nil {
				require.Nil(t, cfg.Extensions)
				return
			}

			require.NotNil(t, cfg.Extensions)
			require.Equal(t, tc.expected.TraceFile, cfg.Extensions.TraceFile)
			require.Equal(t, tc.expected.MetricsFile, cfg.Extensions.MetricsFile)
			require.Equal(t, tc.expected.Propagators, cfg.Extensions.Propagators)

			require.Equal(t, len(tc.expected.Instruments), len(cfg.Extensions.Instruments),
				"instruments count mismatch")

			for name, expectedOpts := range tc.expected.Instruments {
				actualOpts, ok := cfg.Extensions.Instruments[name]
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
