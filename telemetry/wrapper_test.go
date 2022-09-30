// Package telemetry provides instrumentation tools for other modules and packages.
package telemetry

import (
	"testing"
	"time"

	gometrics "github.com/armon/go-metrics"
	"github.com/stretchr/testify/require"
)

func TestNewLabel(t *testing.T) {
	var L gometrics.Label = NewLabel("some-key", "some-value")
	require.NotNil(t, L)
}

func TestWrapperModuleMeasureSince(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	ModuleMeasureSince("some-format", time.Now())
}

func TestWrapperModuleSetGauge(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	ModuleSetGauge("some-format", 1.234)
}

func TestWrapperIncrCounter(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	IncrCounter(1.234)
}

func TestWrapperIncrCounterWithLabels(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	IncrCounterWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperSetGauge(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	SetGauge(1.234)
}

func TestWrapperSetGaugeWithLabels(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	SetGaugeWithLabels([]string{"some-key"}, 1.234, []gometrics.Label{})
}

func TestWrapperMeasureSince(t *testing.T) {
	Default = &metrics{cnf: &Config{globalLabels: []gometrics.Label{}, GlobalLabels: [][]string{}}}
	MeasureSince(time.Now())
}
