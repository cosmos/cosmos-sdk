package server

import (
	"testing"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

func Test_emitServerInfoMetrics(t *testing.T) {
	oriNewInfo := version.NewInfo
	version.NewInfo = func() version.Info {
		return version.Info{
			GoVersion:        "go version go1.14 linux/amd64",
			CosmosSdkVersion: "0.42.5",
		}
	}
	t.Cleanup(func() {
		version.NewInfo = oriNewInfo
	})

	m, err := telemetry.New(telemetry.Config{
		Enabled:                 true,
		EnableHostname:          false,
		ServiceName:             "test",
		PrometheusRetentionTime: 60,
		EnableHostnameLabel:     false,
	})
	require.NoError(t, err)
	require.NotNil(t, m)

	emitServerInfoMetrics("./testdata")

	gr, err := m.Gather(telemetry.FormatPrometheus)
	require.NoError(t, err)
	require.Equal(t, gr.ContentType, string(expfmt.FmtText))

	require.Contains(t, string(gr.Metrics), `test_server_info{go="go version go1.14 linux/amd64",update_height="123",version="0.42.5"`)

	metrics.Shutdown()
}
