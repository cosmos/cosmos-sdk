package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestGRPCParams(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	params, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, params.Params, app.DistrKeeper.GetParams(ctx))
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.DistrKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", sdk.NewDec(5)),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(3)),
	}

	addr := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(1000000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addr)

	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      1,
		CountTotal: false,
	}

	req := types.NewQueryValidatorOutstandingRewardsRequest(nil, pageReq)

	validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)

	require.Error(t, err)
	require.Nil(t, validatorOutstandingRewards)

	// set outstanding rewards
	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, valAddrs[0], types.ValidatorOutstandingRewards{Rewards: valCommission})

	req = types.NewQueryValidatorOutstandingRewardsRequest(valAddrs[0], pageReq)

	validatorOutstandingRewards, err = queryClient.ValidatorOutstandingRewards(gocontext.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, validatorOutstandingRewards.Rewards)
	require.NotNil(t, validatorOutstandingRewards.Res.NextKey)
}
