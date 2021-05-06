package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/poolx/keeper"
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestKeeperParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	var ( // TODO(levi) figure how how to initialize these -- footnote: interesting that the test passes and doesn't care
		cdc        codec.Codec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramSpace paramtypes.Subspace
	)

	queryServer := keeper.NewKeeper(cdc, storeKey, memKey, paramSpace)
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, queryServer)

	queryClient := types.NewQueryClient(queryHelper)

	req := &types.QueryParamsRequest{}

	res, err := queryClient.Params(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, res, &types.QueryParamsResponse{})
}
