package rpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
)

func TestGetChainHeightCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := rpc.GetChainHeight(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestQueryBlocksCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := rpc.QueryBlocks(ctx, 1, 100, "", "")
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetBlockByHeightCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	height := int64(1)
	_, err := rpc.GetBlockByHeight(ctx, &height)
	require.ErrorIs(t, err, context.Canceled)
}

func TestGetBlockByHashCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := rpc.GetBlockByHash(ctx, "")
	require.ErrorIs(t, err, context.Canceled)
}
