package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWriteReadHomeDirToFromFile tests if a given home directory path can be written to
// the corresponding configuration file and if the written information can be read from file.
func TestWriteReadHomeDirToFromFile(t *testing.T) {
	testFolder := t.TempDir()
	defaultHomeDir := filepath.Join(testFolder, "test1")
	// TODO: Change to TOML
	homeFilePath := filepath.Join(defaultHomeDir, "config", "home.txt")

	testcases := []struct {
		name        string
		homeDir     string
		expPass     bool
		errContains string
	}{
		{
			"pass - valid folder path",
			filepath.Join(testFolder, "test2"),
			false,
			"",
		},
		{
			"fail - invalid folder path",
			"invalid-path!",
			false,
			"no such file or directory",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := writeHomeDirToFile(homeFilePath, tc.homeDir)
			if tc.expPass {
				require.NoErrorf(t, err, "no error expected when writing home dir to file")
				homeDir, err := ReadHomeDirFromFile(homeFilePath)
				require.NoError(t, err, "expected no error reading the home dir from the configuration file")
				require.Equal(t,
					tc.homeDir,
					homeDir,
					"expected home dir read from configuration file to equal the testcase filename",
				)
			} else {
				require.Error(t, err, "expected error when writing home dir to file")
				require.ErrorContains(t, err,
					tc.errContains,
					"expected error message to contain %s",
					tc.errContains,
				)
			}
		})
	}
}
