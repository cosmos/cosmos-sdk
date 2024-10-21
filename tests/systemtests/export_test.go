//go:build system_test

package systemtests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestExportCmd(t *testing.T) {
	// scenario: test bank send command
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	exportFile := "foobar.json"

	sut.StartChain(t)

	testCases := []struct {
		name          string
		args          []string
		expErr        bool
		errMsg        string
		expZeroHeight bool
	}{
		{"invalid home dir", []string{"genesis", "export", "--home=foo"}, true, "no such file or directory", false},
		{"should export correct height", []string{"genesis", "export"}, false, "", false},
		{"should export correct height with --height", []string{"genesis", "export", "--height=5"}, false, "", false},
		{"should export height 0 with --for-zero-height", []string{"genesis", "export", "--for-zero-height=true"}, false, "", true},
		{"should export state to the specified file", []string{"genesis", "export", fmt.Sprintf("--output-document=%s", exportFile)}, false, "", false},
	}

	for _, tc := range testCases {
		
		// fmt.Println(tc.name, res)
		if tc.expErr {
			assertOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Contains(t, gotOutputs[0], tc.errMsg)
				return false
			}
			cli.WithRunErrorMatcher(assertOutput).RunCommandWithArgs(tc.args...)
		} else {
			res := cli.RunCommandWithArgs(tc.args...)
			if res == "" {
				require.FileExists(t, exportFile)
				os.Remove(exportFile)
			} else {
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
	}
}
