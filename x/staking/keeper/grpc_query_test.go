package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGRPCQueryValidators(t *testing.T) {
	_, app, ctx := createTestInput()

	addrs := simapp.AddTestAddrs(app, ctx, 500, sdk.TokensFromConsensusPower(10000))

	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8), sdk.NewInt(7)}
	status := []sdk.BondStatus{sdk.Bonded, sdk.Unbonded, sdk.Unbonding}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = types.NewValidator(sdk.ValAddress(addrs[i]), PKs[i], types.Description{})
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		validators[i] = validators[i].UpdateStatus(status[i])
	}

	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidator(ctx, validators[1])
	app.StakingKeeper.SetValidator(ctx, validators[2])

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.StakingKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	_, err := queryClient.Validators(gocontext.Background(), &types.QueryValidatorsRequest{})
	require.Error(t, err)

	for i, s := range status {
		req := &types.QueryValidatorsRequest{Status: s.String(), Req: &query.PageRequest{Limit: 5}}
		valResp, err := queryClient.Validators(gocontext.Background(), req)
		require.NoError(t, err)

		require.Equal(t, 1, len(valResp.Validators))
		require.Equal(t, validators[i].OperatorAddress, valResp.Validators[0].OperatorAddress)
	}

}
