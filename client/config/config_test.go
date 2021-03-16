package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ErrWrongNumberOfArgs = errors.New("wrong number of arguments")

// For https://github.com/cosmos/cosmos-sdk/issues/3899
func Test_runConfigCmdTwiceWithShorterNodeValue(t *testing.T) {
	// Prepare environment
	t.Parallel()
	cleanup := tmpDir(t)
	defer cleanup()
	_ = os.RemoveAll(filepath.Join("config"))

	// Init command config
	cmd := Cmd()
	assert.NotNil(t, cmd)

	err := cmd.RunE(cmd, []string{"node", "tcp://localhost:26657"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "--get"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "tcp://local:26657"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"node", "--get"})
	assert.Nil(t, err)

	err = cmd.RunE(cmd, nil)
	assert.Nil(t, err)

	err = cmd.RunE(cmd, []string{"invalidKey", "--get"})
	require.Equal(t, err, errUnknownConfigKey("invalidKey"))

	err = cmd.RunE(cmd, []string{"invalidArg1"})
	require.Equal(t, err, ErrWrongNumberOfArgs)

	err = cmd.RunE(cmd, []string{"invalidKey", "invalidValue"})
	require.Equal(t, err, errUnknownConfigKey("invalidKey"))

	// TODO add testing of pririty environmental variable, flag and file
	// for now manual testign is ok

}

func tmpDir(t *testing.T) func() {
	dir, err := ioutil.TempDir("", t.Name()+"_")
	require.NoError(t, err)
	return func() { _ = os.RemoveAll(dir) }
}
