package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
)

// For https://github.com/cosmos/cosmos-sdk/issues/3899
func Test_runConfigCmdTwiceWithShorterNodeValue(t *testing.T) {
	// Prepare environment
	configHome, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	_ = os.RemoveAll(filepath.Join(configHome, "config"))
	viper.Set(flags.FlagHome, configHome)

	// Init command config
	cmd := ConfigCmd(configHome)
	require.NotNil(t, cmd)
	require.NoError(t, cmd.RunE(cmd, []string{"node", "tcp://localhost:26657"}))
	require.NoError(t, cmd.RunE(cmd, []string{"node", "--get"}))
	require.NoError(t, cmd.RunE(cmd, []string{"node", "tcp://local:26657"}))
	require.NoError(t, cmd.RunE(cmd, []string{"node", "--get"}))
}

func TestConfigCmd_UnknownOption(t *testing.T) {
	// Prepare environment
	configHome, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	_ = os.RemoveAll(filepath.Join(configHome, "config"))
	viper.Set(flags.FlagHome, configHome)

	// Init command config
	cmd := ConfigCmd(configHome)
	require.NotNil(t, cmd)
	require.Error(t, cmd.RunE(cmd, []string{"invalid", "true"}), "unknown configuration key: \"invalid\"")
}

func TestConfigCmd_OfflineFlag(t *testing.T) {
	// Prepare environment
	configHome, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	_ = os.RemoveAll(filepath.Join(configHome, "config"))
	viper.Set(flags.FlagHome, configHome)

	// Init command config
	cmd := ConfigCmd(configHome)
	_, out, _ := tests.ApplyMockIO(cmd)
	require.NotNil(t, cmd)

	viper.Set(flagGet, true)
	require.NoError(t, cmd.RunE(cmd, []string{"offline"}))
	require.Contains(t, out.String(), "false")
	out.Reset()

	viper.Set(flagGet, false)
	require.NoError(t, cmd.RunE(cmd, []string{"offline", "true"}))

	viper.Set(flagGet, true)
	require.NoError(t, cmd.RunE(cmd, []string{"offline"}))
	require.Contains(t, out.String(), "true")
}
