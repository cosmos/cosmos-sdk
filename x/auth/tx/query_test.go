package tx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func TestQueryTxCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := tx.QueryTx(ctx, "")
	require.ErrorIs(t, err, context.Canceled)
}

func TestQueryTxsByEventsCancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := tx.QueryTxsByEvents(ctx, 1, 100, "query", "")
	require.ErrorIs(t, err, context.Canceled)
}
