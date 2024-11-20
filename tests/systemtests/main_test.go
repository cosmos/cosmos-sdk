//go:build system_test

package systemtests

import (
	systest "cosmossdk.io/systemtests"
	"testing"
)

func TestMain(m *testing.M) {
	systest.RunTests(m)
}
