package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/server/mock"
	"github.com/cosmos/cosmos-sdk/wire"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

// TODO update
func TestInitCmd(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := NewContext(cfg, logger)
	cdc := wire.NewCodec()
	appInit := AppInit{
		AppGenState: mock.AppGenState,
		AppGenTx:    mock.AppGenTx,
	}
	cmd := InitCmd(ctx, cdc, appInit)
	err = cmd.RunE(nil, nil)
	require.NoError(t, err)
}

func TestGenTxCmd(t *testing.T) {
	// TODO
}

func TestSimpleAppGenTx(t *testing.T) {
	// TODO
}

func TestSimpleAppGenState(t *testing.T) {
	// TODO
}
