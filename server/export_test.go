package server

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/tendermint/tendermint/libs/log"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"os"
	"bytes"
	"io"
		"github.com/cosmos/cosmos-sdk/server/mock"
	)

func TestEmptyState(t *testing.T) {
	defer setupViper(t)()
	logger := log.NewNopLogger()
	cfg, err := tcmd.ParseConfig()
	require.Nil(t, err)
	ctx := NewContext(cfg, logger)
	cdc := wire.NewCodec()
	appInit := AppInit{
		AppGenTx:    mock.AppGenTx,
		AppGenState: mock.AppGenStateEmpty,
	}
	cmd := InitCmd(ctx, cdc, appInit)
	err = cmd.RunE(nil, nil)
	require.NoError(t, err)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	cmd = ExportCmd(ctx, cdc, nil)
	err = cmd.RunE(nil, nil)
	require.NoError(t, err)

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = old
	out := <-outC
	require.Contains(t, out, "WARNING: State is not initialized")
	require.Contains(t, out, "genesis_time")
	require.Contains(t, out, "chain_id")
	require.Contains(t, out, "consensus_params")
	require.Contains(t, out, "validators")
	require.Contains(t, out, "app_hash")
}
