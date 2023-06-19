package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"gotest.tools/v3/assert"
)

// TestWriteReadHomeDirToFromFile tests if a given home directory path can be written to
// the corresponding configuration file and if the written information can be read from file.
func TestWriteReadHomeDirToFromFile(t *testing.T) {
	testFolder := t.TempDir()
	newHome := "path/to/new/home"
	defaultConfigPath := filepath.Join(testFolder, "config")
	err := os.MkdirAll(defaultConfigPath, os.ModePerm)
	assert.NilError(t, err, "expected no error creating the default configuration path")

	// create empty context with viper setup for parsing the home.toml file
	ctx := client.Context{}.WithViper("")

	testcases := []struct {
		name         string
		homeFilePath string
		expPass      bool
		errContains  string
	}{
		{
			name:         "pass - valid folder path",
			homeFilePath: filepath.Join(defaultConfigPath, "home.toml"),
			expPass:      true,
		},
		{
			name:         "fail - invalid folder path",
			homeFilePath: "invalid/path!",
			expPass:      false,
			errContains:  "no such file or directory",
		},
	}
	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err := WriteHomeDirToFile(tc.homeFilePath, newHome)
			if tc.expPass {
				assert.NilError(t, err, "no error expected when writing home dir to file")
				homeDir, err := ReadOrMakeHomeDir(defaultConfigPath, ctx.Viper)
				assert.NilError(t, err, "expected no error reading the home dir from the configuration file")
				assert.Equal(t, homeDir, newHome,
					"expected home dir read from configuration file to be %q; got: %q", newHome, homeDir,
				)
			} else {
				assert.ErrorContains(t, err, tc.errContains,
					"expected error message to contain %s", tc.errContains,
				)
			}
		})
	}
}
