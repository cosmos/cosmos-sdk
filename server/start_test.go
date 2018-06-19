package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/server/mock"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/tendermint/abci/server"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tmlibs/log"
)

func TestStartStandAlone(t *testing.T) {
	home, err := ioutil.TempDir("", "mock-sdk-cmd")
	require.Nil(t, err)
	defer func() {
		os.RemoveAll(home)
	}()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := NewContext(cfg, logger)
	cdc := wire.NewCodec()
	appInit := AppInit{
		AppGenState: mock.AppGenState,
		AppGenTx:    mock.AppGenTx,
	}
	initCmd := InitCmd(ctx, cdc, appInit)
	err = initCmd.RunE(nil, nil)
	require.NoError(t, err)

	app, err := mock.NewApp(home, logger)
	require.Nil(t, err)
	svrAddr, _, err := FreeTCPAddr()
	require.Nil(t, err)
	svr, err := server.NewServer(svrAddr, "socket", app)
	require.Nil(t, err, "error creating listener")
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	timer := time.NewTimer(time.Duration(5) * time.Second)
	select {
	case <-timer.C:
		svr.Stop()
	}
}

func TestStartWithTendermint(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).
		With("module", "mock-cmd")
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := NewContext(cfg, logger)
	cdc := wire.NewCodec()
	appInit := AppInit{
		AppGenState: mock.AppGenState,
		AppGenTx:    mock.AppGenTx,
	}
	initCmd := InitCmd(ctx, cdc, appInit)
	err = initCmd.RunE(nil, nil)
	require.NoError(t, err)

	// set up app and start up
	viper.Set(flagWithTendermint, true)
	startCmd := StartCmd(ctx, mock.NewApp)
	svrAddr, _, err := FreeTCPAddr()
	require.NoError(t, err)
	startCmd.Flags().Set(flagAddress, svrAddr) // set to a new free address
	timeout := time.Duration(5) * time.Second

	close(RunOrTimeout(startCmd, timeout, t))
}
