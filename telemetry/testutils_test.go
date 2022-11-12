// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func initMetric(t *testing.T) Metrics {

	err := Init(Config{
		Enabled:                     true,
		useGlobalMetricRegistration: false,
	})
	require.NoError(t, err)

	return Default
}
