package server_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	server "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/cometbft"
	"github.com/stretchr/testify/require"
)

func TestServer_Start(t *testing.T) {
	logger := log.NewNopLogger()
	// db := dbm.New()
	builder := runtime.AppBuilder{}
	app, err := builder.Build(nil)
	require.NoError(t, err)

	cmt := cometbft.NewCometBFTServer(logger, app, nil, nil)

	svr := server.NewServer(logger, cmt)

	require.NoError(t, svr.Start(context.TODO()))
}
