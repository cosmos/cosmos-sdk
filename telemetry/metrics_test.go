// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"testing"
	"time"

	gometrics "github.com/armon/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		M := initMetric(t)
		require.NotNil(t, M)
	})

}

func TestInit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		M := initMetric(t)
		require.NotNil(t, M)
	})
	t.Run("option error", func(t *testing.T) {
		Init(Config{})
	})
}

func TestGather(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	var err error
	m := M.(*metrics)

	t.Run("prometheus", func(t *testing.T) {
		m.prometheusEnabled = true
		_, err = m.Gather(FormatPrometheus)
		require.NoError(t, err)
	})

	t.Run("text", func(t *testing.T) {
		_, err = m.Gather(FormatText)
		require.NoError(t, err)
	})

	t.Run("default", func(t *testing.T) {
		_, err = m.Gather(FormatDefault)
		require.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err = m.Gather("InvalidFormat")
		require.Error(t, err)
	})

}

func TestModuleMeasureSince(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.ModuleMeasureSince("module-module-measure-since", time.Now().Add(-1*time.Minute), "key-module-measure-since")
}

func TestModuleSetGauge(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.ModuleSetGauge("module-set-gauge-name", 1.234, "key-module-set-gauge")
}

func TestIncrCounter(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.IncrCounter(1.234)
}

func TestIncrCounterWithLabels(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.IncrCounterWithLabels([]string{"some-key-counter-with-labels"}, 1.234, []gometrics.Label{})
}

func TestSetGauge(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.SetGauge(1.234, "key-set-gauge")

}

func TestSetGaugeWithLabels(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.SetGaugeWithLabels([]string{"some-key-gauge-with-labels"}, 1.234, []gometrics.Label{})
}

func TestMeasureSince(t *testing.T) {
	M := initMetric(t)
	require.NotNil(t, M)
	m := M.(*metrics)
	m.MeasureSince(time.Now().Add(-1 * time.Minute))
}

func TestGatherPrometheus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		M := initMetric(t)
		require.NotNil(t, M)
		m := M.(*metrics)
		m.prometheusEnabled = true
		res, err := m.gatherPrometheus()
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("not enabled", func(t *testing.T) {
		M := initMetric(t)
		require.NotNil(t, M)
		m := M.(*metrics)
		m.prometheusEnabled = false
		_, err := m.gatherPrometheus()
		require.Error(t, err)
	})

}

func TestGatherGeneric(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		M := initMetric(t)
		require.NotNil(t, M)
		m := M.(*metrics)
		res, err := m.gatherGeneric()
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}
