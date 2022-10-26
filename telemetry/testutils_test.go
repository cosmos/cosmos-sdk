// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	initOnce sync.Once
)

func initMetric(t *testing.T) Metrics {
	initOnce.Do(func() {
		err := Init(OptionWithEnable(true), OptionUseGlobalMetricRegistration(false))
		require.NoError(t, err)
	})
	return Default
}
