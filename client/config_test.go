package client

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// For https://github.com/cosmos/cosmos-sdk/issues/3899
func Test_runConfigCmdTwiceWithShorterNodeValue(t *testing.T) {
	// Prepare environment
	t.Parallel()
	configHome, cleanup := tmpDir(t)
	defer cleanup()
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

func tmpDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", t.Name()+"_")
	require.NoError(t, err)
	return dir, func() { _ = os.RemoveAll(dir) }
}
