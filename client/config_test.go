package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/tests"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// For https://github.com/cosmos/cosmos-sdk/issues/3899
func Test_runConfigCmdTwiceWithShorterNodeValue(t *testing.T) {
	// Prepare environment
	t.Parallel()

	configHome, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	_ = os.RemoveAll(filepath.Join(configHome, "config"))
	viper.Set(flags.FlagHome, configHome)

	// Init command config
	cmd := ConfigCmd(configHome)
	assert.NotNil(t, cmd)

	err := cmd.RunE(cmd, []string{"node", "tcp://localhost:26657"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "--get"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "tcp://local:26657"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "--get"})
	assert.Nil(t, err)
}

func TestConfigCmd_OfflineFlag(t *testing.T) {
	// Prepare environment
	t.Parallel()

	configHome, cleanup := tests.NewTestCaseDir(t)
	t.Cleanup(cleanup)

	_ = os.RemoveAll(filepath.Join(configHome, "config"))
	viper.Set(flags.FlagHome, configHome)

	// Init command config
	cmd := ConfigCmd(configHome)
	assert.NotNil(t, cmd)

	viper.Set(flagGet, true)
	err := cmd.RunE(cmd, []string{"offline"})
	assert.Nil(t, err)

	viper.Set(flagGet, false)
	err = cmd.RunE(cmd, []string{"offline", "true"})
	assert.Nil(t, err)

	viper.Set(flagGet, true)
	err = cmd.RunE(cmd, []string{"offline"})
	assert.Nil(t, err)
}
