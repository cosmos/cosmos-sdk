package tx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func TestContextCancellation(t *testing.T) {
	testCases := []struct {
		name  string
		query func(ctx client.Context) error
	}{
		{
			name: "query tx cancellation",
			query: func(ctx client.Context) error {
				_, err := tx.QueryTx(ctx, "")
				return err
			},
		},
		{
			name: "query txs by events cancellation",
			query: func(ctx client.Context) error {
				_, err := tx.QueryTxsByEvents(ctx, 1, 100, "query", "")
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
