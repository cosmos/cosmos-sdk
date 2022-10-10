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
	goCtx := sdk.WrapSDKContext(ctx)

	resp, err := svr.Config(goCtx, &ConfigRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, ctx.MinGasPrices().String(), resp.MinimumGasPrice)
	require.NotEmpty(t, resp.Bech32AccountAddr)
	require.NotEmpty(t, resp.Bech32ValidatorAddr)
	require.NotEmpty(t, resp.Bech32ConsensusAddr)
	require.NotEmpty(t, resp.Bech32AccountPub)
	require.NotEmpty(t, resp.Bech32ValidatorPub)
	require.NotEmpty(t, resp.Bech32ConsensusPub)
}
