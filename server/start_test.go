package server

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/server/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/tendermint/tendermint/abci/server"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/libs/log"
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
	ctx := sdk.NewServerContext(cfg, logger)
	cdc := wire.NewCodec()
	appInit := AppInit{
		AppGenState: mock.AppGenState,
		AppGenTx:    mock.AppGenTx,
	}
	initCmd := InitCmd(ctx, cdc, appInit)
	err = initCmd.RunE(nil, nil)
	require.NoError(t, err)

	app, err := mock.NewApp(home, ctx)
	require.Nil(t, err)
	svrAddr, _, err := FreeTCPAddr()
	require.Nil(t, err)
	svr, err := server.NewServer(svrAddr, "socket", app)
	require.Nil(t, err, "error creating listener")
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	timer := time.NewTimer(time.Duration(2) * time.Second)
	select {
	case <-timer.C:
		svr.Stop()
	}
}
