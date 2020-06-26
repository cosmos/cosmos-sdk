package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestGRPCParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.MintKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	params, err := queryClient.Params(gocontext.Background(), &types.QueryEmptyRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, params.Params)

	inflation, err := queryClient.Inflation(gocontext.Background(), &types.QueryEmptyRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, inflation.Inflation)

	annualProvisions, err := queryClient.AnnualProvisions(gocontext.Background(), &types.QueryEmptyRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, annualProvisions.AnnualProvisions)
}