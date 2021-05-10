package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeeperParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.PoolxKeeper)

	queryClient := types.NewQueryClient(queryHelper)

	req := &types.QueryParamsRequest{}

	res, err := queryClient.Params(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, res, &types.QueryParamsResponse{})
}
