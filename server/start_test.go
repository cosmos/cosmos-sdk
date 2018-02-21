package server

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/mock"
	"github.com/tendermint/tmlibs/log"
)

func TestStart(t *testing.T) {
	defer setupViper()()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock-cmd")
		// logger := log.NewNopLogger()
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// try to start up
	// this should hang forever on success.... how to close???

	rootDir := viper.GetString("home")
	app, err := mock.NewApp(logger, rootDir)
	require.NoError(t, err)
	_ = StartCmd(app, logger)
	// startCmd := StartCmd(app, logger)

	// // TODO: test with tendermint
	// err = startCmd.RunE(nil, nil)
	// require.NoError(t, err)
}
