package cmd_test

import (
	"path/filepath"
	"testing"

	"cosmossdk.io/tools/confix/cmd"
	"github.com/cosmos/cosmos-sdk/client/config"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"gotest.tools/v3/assert"
)

func TestHomeCommand(t *testing.T) {
	testDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		expErr      bool
		errContains string
	}{
		{
			name:   "home command - no args",
			args:   []string{},
			expErr: false,
		},
		{
			name:   "home command one arg - valid directory",
			args:   []string{filepath.Join(testDir, "newHome")},
			expErr: false,
		},
		{
			name:        "home command one arg - invalid directory",
			args:        []string{"/invalid/dir/:path"},
			expErr:      true,
			errContains: "couldn't make client config",
		},
		{
			name:        "home command two args - invalid number of args",
			args:        []string{"arg1", "arg2"},
			expErr:      true,
			errContains: "accepts between 0 and 1 arg(s)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clientCtx, cleanup := initClientContext(t)
			defer cleanup()

			_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd.HomeCommand(), tc.args)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.errContains, "expected error to contain %s", tc.errContains)
			} else {
				assert.NilError(t, err, "expected no error")
			}
		})
	}
}

func TestUpdateHome(t *testing.T) {
	testDir := t.TempDir()

	clientCtx, cleanup := initClientContext(t)
	defer cleanup()

	setNewHome := []string{filepath.Join(testDir, "newHome")}
	_, err := clitestutil.ExecTestCLICmd(clientCtx, cmd.HomeCommand(), setNewHome)
	assert.NilError(t, err, "expected no error")

	homeFileDir := filepath.Dir(clientCtx.HomeFilePath)
	homeDir, err := config.ReadHomeDir(homeFileDir, clientCtx.Viper)
	assert.NilError(t, err, "expected no error reading the home dir from the configuration file")
	assert.Equal(t, homeDir, setNewHome[0], "expected home dir read from configuration file to be %q; got: %q", setNewHome[0], homeDir)

	setUpdatedHome := []string{filepath.Join(testDir, "updatedHome")}
	_, err = clitestutil.ExecTestCLICmd(clientCtx, cmd.HomeCommand(), setUpdatedHome)
	assert.NilError(t, err, "expected no error")

	homeDir, err = config.ReadHomeDir(homeFileDir, clientCtx.Viper)
	assert.NilError(t, err, "expected no error reading the home dir from the configuration file")
	assert.Equal(t, homeDir, setUpdatedHome[0], "expected home dir read from configuration file to be %q; got: %q", setUpdatedHome[0], homeDir)
}
