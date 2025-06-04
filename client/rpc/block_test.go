package rpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rpc"
)

func TestContextCancellation(t *testing.T) {
	testCases := []struct {
		name  string
		query func(ctx client.Context) error
	}{
		{
			name: "get chain height cancellation",
			query: func(ctx client.Context) error {
				_, err := rpc.GetChainHeight(ctx)
				return err
			},
		},
		{
			name: "query blocks cancellation",
			query: func(ctx client.Context) error {
				_, err := rpc.QueryBlocks(ctx, 1, 100, "", "")
				return err
			},
		},
		{
			name: "get block by height cancellation",
			query: func(ctx client.Context) error {
				height := int64(1)
				_, err := rpc.GetBlockByHeight(ctx, &height)
				return err
			},
		},
		{
			name: "get block by hash cancellation",
			query: func(ctx client.Context) error {
				_, err := rpc.GetBlockByHash(ctx, "")
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmdCtx, cancel := context.WithCancel(context.Background())
			cancel()

			ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
			err := tc.query(ctx)
			require.ErrorIs(t, err, context.Canceled)
		})
	}
}
