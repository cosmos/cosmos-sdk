package server

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/mock"
	"github.com/tendermint/tmlibs/log"
)

func TestStartStandAlone(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewNopLogger()
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, false)
	viper.Set(flagAddress, "localhost:11122")
	startCmd := StartCmd(mock.NewApp, logger)
	startCmd.Flags().Set(flagAddress, FreeTCPAddr(t)) // set to a new free address
	timeout := time.Duration(10) * time.Second

	close(RunOrTimeout(startCmd, timeout, t))
}

func TestStartWithTendermint(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock-cmd")
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, true)
	startCmd := StartCmd(mock.NewApp, logger)
	startCmd.Flags().Set(flagAddress, FreeTCPAddr(t)) // set to a new free address
	timeout := time.Duration(10) * time.Second

	close(RunOrTimeout(startCmd, timeout, t))
}
