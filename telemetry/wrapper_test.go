package telemetry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	initTelemetry(true)

	currentTime := time.Now()
	telemetryTime := Now()

	assert.NotEqual(t, time.Time{}, telemetryTime, "Now() should not return zero time when telemetry is enabled")
	assert.WithinDuration(t, currentTime, telemetryTime, time.Second, "Now() should be close to current time")

	initTelemetry(false)

	telemetryTime = Now()
	assert.Equal(t, time.Time{}, telemetryTime, "Now() should return zero time when telemetry is disabled")
}

func TestIsTelemetryEnabled(t *testing.T) {
	initTelemetry(true)
	if !isTelemetryEnabled() {
		t.Errorf("isTelemetryEnabled() should return true when globalTelemetryEnabled is set to true")
	}

	initTelemetry(false)
	if isTelemetryEnabled() {
		t.Errorf("isTelemetryEnabled() should return false when globalTelemetryEnabled is set to false")
	}
}
