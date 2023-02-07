package node

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestServiceServer_Config(t *testing.T) {
	svr := NewQueryServer(client.Context{})
	ctx := sdk.Context{}.WithMinGasPrices(sdk.NewDecCoins(sdk.NewInt64DecCoin("stake", 15)))

	resp, err := svr.Config(ctx, &ConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, ctx.MinGasPrices().String(), resp.MinimumGasPrice)
}
