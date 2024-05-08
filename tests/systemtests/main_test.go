//go:build system_test

package systemtests

import "testing"

func TestMain(m *testing.M) {
	RunTests(m)
}
