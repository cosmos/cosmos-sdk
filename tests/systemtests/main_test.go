//go:build system_test

package systemtests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

func TestMain(m *testing.M) {
	systemtests.RunTests(m)
}
