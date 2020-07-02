package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func TestGRPCQueryParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	app.SlashingKeeper.SetParams(ctx, keeper.TestParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.SlashingKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	params, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	require.NoError(t, err)

	require.Equal(t, keeper.TestParams(), params)
}
