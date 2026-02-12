package iavl

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/telemetry"
)

func TestMain(m *testing.M) {
	// make sure we shutdown telemetry at the end of the test, to collect benchmark metrics
	telemetry.TestingMain(m, nil)
}
