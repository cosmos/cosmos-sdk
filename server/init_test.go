package server

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	"github.com/cosmos/cosmos-sdk/mock"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
)

func TestInit(t *testing.T) {
	defer setupViper(t)()

	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := NewContext(cfg, logger)
	cdc := app.MakeCodec()
	cmd := InitCmd(cdc, mock.GenInitOptions, ctx)
	err = cmd.RunE(nil, nil)
	require.NoError(t, err)
}
