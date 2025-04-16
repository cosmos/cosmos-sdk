//go:build system_test

package systemtests

import (
	"testing"

	"cosmossdk.io/systemtests"
)

func TestMain(m *testing.M) {
	systemtests.RunTests(m)
}
