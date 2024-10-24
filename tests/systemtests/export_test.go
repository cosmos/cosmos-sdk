//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestExportCmd_WithHeight(t *testing.T) {
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	sut.StartChain(t)

	// Wait 10s for producing blocks
	time.Sleep(10 * time.Second)

	sut.StopChain()

	testCases := []struct {
		name          string
		args          []string
		expZeroHeight bool
	}{
		{"should export correct height", []string{"genesis", "export", "--home", sut.nodePath(0)}, false},
		{"should export correct height with --height", []string{"genesis", "export", "--height=5", "--home", sut.nodePath(0), "--log_level=disabled"}, false},
		{"should export height 0 with --for-zero-height", []string{"genesis", "export", "--for-zero-height=true", "--home", sut.nodePath(0)}, true},
	}

	for _, tc := range testCases {
		res := cli.RunCommandWithArgs(tc.args...)
		height := gjson.Get(res, "initial_height").Int()
		if tc.expZeroHeight {
			require.Equal(t, height, int64(0))
		} else {
			require.Greater(t, height, int64(0))
		}

		// Check consensus params of exported state
		maxGas := gjson.Get(res, "consensus.params.block.max_gas").Int()
		require.Equal(t, maxGas, int64(MaxGas))
	}
}

func TestExportCmd_WithFileFlag(t *testing.T) {
	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	exportFile := "foobar.json"

	sut.StartChain(t)

	// Wait 10s for producing blocks
	time.Sleep(10 * time.Second)

	sut.StopChain()

	testCases := []struct {
		name   string
		args   []string
		expErr bool
		errMsg string
	}{
		{"invalid home dir", []string{"genesis", "export", "--home=foo"}, true, "no such file or directory"},
		{"should export state to the specified file", []string{"genesis", "export", fmt.Sprintf("--output-document=%s", exportFile), "--home", sut.nodePath(0)}, false, ""},
	}

	for _, tc := range testCases {
		if tc.expErr {
			assertOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Contains(t, gotOutputs[0], tc.errMsg)
				return false
			}
			cli.WithRunErrorMatcher(assertOutput).RunCommandWithArgs(tc.args...)
		} else {
			cli.RunCommandWithArgs(tc.args...)
			require.FileExists(t, exportFile)
			err := os.Remove(exportFile)
			require.NoError(t, err)
		}
	}
}
