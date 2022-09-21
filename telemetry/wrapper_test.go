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
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().ModuleMeasureSince(gomock.Any(), gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.ModuleMeasureSince("some-format", time.Now())

}

func TestWrapperModuleSetGauge(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().ModuleSetGauge(gomock.Any(), gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.ModuleSetGauge("some-format", 1.234)
}

func TestWrapperIncrCounter(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().IncrCounter(gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.IncrCounter(1.234)
}

func TestWrapperIncrCounterWithLabels(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().IncrCounterWithLabels(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.IncrCounterWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperSetGauge(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().SetGauge(gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.SetGauge(1.234)
}

func TestWrapperSetGaugeWithLabels(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().SetGaugeWithLabels(gomock.Any(), gomock.Any(), gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.SetGaugeWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperMeasureSince(t *testing.T) {
	C := gomock.NewController(t)
	MM := mock.NewMockMetrics(C)
	MM.EXPECT().MeasureSince(gomock.Any()).Return().AnyTimes()
	telemetry.Default = MM
	telemetry.MeasureSince(time.Now())
}
