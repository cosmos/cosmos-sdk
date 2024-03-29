package telemetry

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// TelemetrySuite is a struct that holds the setup for the telemetry tests.
// It includes a mutex to ensure that tests that depend on the global state
// do not run in parallel, which can cause race conditions and unpredictable results.
type TelemetrySuite struct {
	suite.Suite
	mu sync.Mutex
}

// SetupTest is called before each test to reset the global state to a known disabled state.
// This ensures each test starts with the telemetry disabled
func (suite *TelemetrySuite) SetupTest() {
	initTelemetry(false)
}

// TestNow tests the Now function when telemetry is enabled and disabled.
func (suite *TelemetrySuite) TestNow() {
	suite.mu.Lock()
	defer suite.mu.Unlock()

	initTelemetry(true)
	telemetryTime := Now()
	suite.NotEqual(time.Time{}, telemetryTime, "Now() should not return zero time when telemetry is enabled")

	initTelemetry(false)
	telemetryTime = Now()
	suite.Equal(time.Time{}, telemetryTime, "Now() should return zero time when telemetry is disabled")
}

// TestIsTelemetryEnabled tests the isTelemetryEnabled function.
func (suite *TelemetrySuite) TestIsTelemetryEnabled() {
	suite.mu.Lock()
	defer suite.mu.Unlock()

	initTelemetry(true)
	suite.True(isTelemetryEnabled(), "isTelemetryEnabled() should return true when globalTelemetryEnabled is set to true")

	initTelemetry(false)
	suite.False(isTelemetryEnabled(), "isTelemetryEnabled() should return false when globalTelemetryEnabled is set to false")
}

// TestTelemetrySuite initiates the test suite.
func TestTelemetrySuite(t *testing.T) {
	suite.Run(t, new(TelemetrySuite))
}
