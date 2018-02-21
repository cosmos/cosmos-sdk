package server

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/mock"
)

// setupViper creates a homedir to run inside,
// and returns a cleanup function to defer
func setupViper() func() {
	rootDir, err := ioutil.TempDir("", "mock-sdk-cmd")
	if err != nil {
		panic(err) // fuck it!
	}
	viper.Set("home", rootDir)
	return func() {
		os.RemoveAll(rootDir)
	}
}

func TestInit(t *testing.T) {
	defer setupViper()()

	logger := log.NewNopLogger()
	cmd := InitCmd(mock.GenInitOptions, logger)
	err := cmd.RunE(nil, nil)
	require.NoError(t, err)
}
