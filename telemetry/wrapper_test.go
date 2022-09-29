// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry_test

import (
	"testing"
	"time"

	gometrics "github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNewLabel(t *testing.T) {
	var L gometrics.Label = telemetry.NewLabel("some-key", "some-value")
	require.NotNil(t, L)
}

func TestWrapperGather(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().Gather(gomock.Any()).Return(telemetry.GatherResponse{}, nil).AnyTimes()
	telemetry.Default = MM
	_, err := telemetry.Gather("some-format")
	require.NoError(t, err)
}

func TestWrapperModuleMeasureSince(t *testing.T) {
	telemetry.ModuleMeasureSince("some-format", time.Now())
}

func TestWrapperModuleSetGauge(t *testing.T) {
	telemetry.ModuleSetGauge("some-format", 1.234)
}

func TestWrapperIncrCounter(t *testing.T) {
	telemetry.IncrCounter(1.234)
}

func TestWrapperIncrCounterWithLabels(t *testing.T) {
	telemetry.IncrCounterWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperSetGauge(t *testing.T) {
	telemetry.SetGauge(1.234)
}

func TestWrapperSetGaugeWithLabels(t *testing.T) {
	telemetry.SetGaugeWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperMeasureSince(t *testing.T) {
	telemetry.MeasureSince(time.Now())
}
