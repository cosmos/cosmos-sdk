package telemetry

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func TestMetrics_Disabled(t *testing.T) {
	m, err := New(Config{Enabled: false}, "")
	require.Nil(t, m)
	require.Nil(t, err)
}

func TestMetrics_InMem(t *testing.T) {
	m, err := New(Config{
		Enabled:        true,
		EnableHostname: false,
		ServiceName:    "test",
	}, "")
	require.NoError(t, err)
	require.NotNil(t, m)

	emitMetrics()

	gr, err := m.Gather(FormatText)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, "application/json")

	jsonMetrics := make(map[string]interface{})
	require.NoError(t, json.Unmarshal(gr.Metrics, &jsonMetrics))

	counters := jsonMetrics["Counters"].([]interface{})
	require.Equal(t, counters[0].(map[string]interface{})["Count"].(float64), 10.0)
	require.Equal(t, counters[0].(map[string]interface{})["Name"].(string), "test.dummy_counter")
}

func TestMetrics_Prom(t *testing.T) {
	m, err := New(Config{
		Enabled:                 true,
		EnableHostname:          false,
		ServiceName:             "test",
		PrometheusRetentionTime: 60,
		EnableHostnameLabel:     false,
	}, "")
	require.NoError(t, err)
	require.NotNil(t, m)
	require.True(t, m.prometheusEnabled)

	emitMetrics()

	gr, err := m.Gather(FormatPrometheus)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, string(expfmt.FmtText))

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

func TestMetrics_getDefaultGlobalLabels(t *testing.T) {
	type args struct {
		rootDir string
	}
	tests := []struct {
		name        string
		args        args
		versionInfo version.Info
		want        []metrics.Label
	}{
		{
			name: "happy path",
			args: args{
				rootDir: "./testdata/valid",
			},
			versionInfo: version.Info{
				GoVersion:        "go version go1.14 linux/amd64",
				CosmosSdkVersion: "0.42.5",
			},
			want: []metrics.Label{
				NewLabel("go", "go version go1.14 linux/amd64"),
				NewLabel("version", "0.42.5"),
				NewLabel("upgrade_height", "123"),
			},
		},
		{
			name: "empty versionInfo",
			args: args{
				rootDir: "./testdata/valid",
			},
			want: []metrics.Label{
				NewLabel("upgrade_height", "123"),
			},
		},
		{
			name: "upgrade-info.json is not available",
			args: args{
				rootDir: "./home",
			},
			versionInfo: version.Info{
				GoVersion:        "go version go1.14 linux/amd64",
				CosmosSdkVersion: "0.42.5",
			},
			want: []metrics.Label{
				NewLabel("go", "go version go1.14 linux/amd64"),
				NewLabel("version", "0.42.5"),
			},
		},
		{
			name: "upgrade-info.json unmarshal failed",
			args: args{
				rootDir: "./testdata/invalid",
			},
			versionInfo: version.Info{
				GoVersion:        "go version go1.14 linux/amd64",
				CosmosSdkVersion: "0.42.5",
			},
			want: []metrics.Label{
				NewLabel("go", "go version go1.14 linux/amd64"),
				NewLabel("version", "0.42.5"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVersionNewInfo(t, tt.versionInfo)
			require.ElementsMatch(t, tt.want, getDefaultGlobalLabels(tt.args.rootDir))
		})
	}
}

func mockVersionNewInfo(t *testing.T, newInfo version.Info) {
	oriNewInfo := version.NewInfo
	version.NewInfo = func() version.Info {
		return newInfo
	}
	t.Cleanup(func() {
		version.NewInfo = oriNewInfo
	})
}
