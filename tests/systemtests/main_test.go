//go:build system_test

package systemtests

import (
	"testing"

	systest "cosmossdk.io/systemtests"
)

func TestMain(m *testing.M) {
	systest.RunTests(m)
}
