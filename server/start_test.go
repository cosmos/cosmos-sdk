package server

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/mock"
	"github.com/tendermint/tmlibs/log"
)

func TestStartStandAlone(t *testing.T) {
	defer setupViper()()

	logger := log.NewNopLogger()
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	rootDir := viper.GetString("home")
	app, err := mock.NewApp(logger, rootDir)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, false)
	viper.Set(flagAddress, "localhost:11122")
	startCmd := StartCmd(app, logger)
	timeout := time.Duration(3) * time.Second

	err = runOrTimeout(startCmd, timeout)
	require.NoError(t, err)
}

func TestStartWithTendermint(t *testing.T) {
	defer setupViper()()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock-cmd")
		// logger := log.NewNopLogger()
	initCmd := InitCmd(mock.GenInitOptions, logger)
	err := initCmd.RunE(nil, nil)
	require.NoError(t, err)

	rootDir := viper.GetString("home")
	app, err := mock.NewApp(logger, rootDir)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, true)
	startCmd := StartCmd(app, logger)
	timeout := time.Duration(3) * time.Second

	err = runOrTimeout(startCmd, timeout)
	require.NoError(t, err)
}

func runOrTimeout(cmd *cobra.Command, timeout time.Duration) error {
	done := make(chan error)
	go func(out chan<- error) {
		// this should NOT exit
		err := cmd.RunE(nil, nil)
		if err != nil {
			out <- err
		}
		out <- fmt.Errorf("start died for unknown reasons")
	}(done)
	timer := time.NewTimer(timeout)

	select {
	case err := <-done:
		return err
	case <-timer.C:
		return nil
	}
}
