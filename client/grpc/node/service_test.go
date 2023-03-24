package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestServiceServer_Config(t *testing.T) {
	svr := NewQueryServer(client.Context{})

	ctx := sdk.Context{}.WithContext(context.Background()).WithMinGasPrices(sdk.NewDecCoins(sdk.NewInt64DecCoin("stake", 15)))
	goCtx := sdk.WrapSDKContext(ctx)

	resp, err := svr.Config(goCtx, &ConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, ctx.MinGasPrices().String(), resp.MinimumGasPrice)
}
