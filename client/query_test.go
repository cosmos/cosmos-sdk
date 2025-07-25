package client_test

import (
	"context"
	"testing"

	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
)

func TestQueryABCICancellation(t *testing.T) {
	cmdCtx, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := client.Context{}.WithClient(client.MockClient{}).WithCmdContext(cmdCtx)
	_, err := ctx.QueryABCI(abci.QueryRequest{})
	require.ErrorIs(t, err, context.Canceled)
}
