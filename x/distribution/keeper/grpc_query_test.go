package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"cosmossdk.io/x/distribution/keeper"
	distrtestutil "cosmossdk.io/x/distribution/testutil"
	"cosmossdk.io/x/distribution/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestQueryParams(t *testing.T) {
	ctx, _, distrKeeper, _ := initFixture(t)
	queryServer := keeper.NewQuerier(distrKeeper)

	cases := []struct {
		name   string
		req    *types.QueryParamsRequest
		resp   *types.QueryParamsResponse
		errMsg string
	}{
		{
			name: "success",
			req:  &types.QueryParamsRequest{},
			resp: &types.QueryParamsResponse{
				Params: types.DefaultParams(),
			},
			errMsg: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := queryServer.Params(ctx, tc.req)
			if tc.errMsg == "" {
				require.NoError(t, err)
				require.Equal(t, tc.resp, out)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, out)
			}
		})
	}
}

func TestQueryValidatorDistributionInfo(t *testing.T) {
	ctx, addrs, distrKeeper, dep := initFixture(t)
	queryServer := keeper.NewQuerier(distrKeeper)

	val, err := distrtestutil.CreateValidator(valConsPk0, math.NewInt(100))
	require.NoError(t, err)

	del := stakingtypes.NewDelegation(addrs[0].String(), val.OperatorAddress, val.DelegatorShares)

	dep.stakingKeeper.EXPECT().Validator(gomock.Any(), gomock.Any()).Return(val, nil).AnyTimes()
	dep.stakingKeeper.EXPECT().Delegation(gomock.Any(), gomock.Any(), gomock.Any()).Return(del, nil).AnyTimes()

	cases := []struct {
		name   string
		req    *types.QueryValidatorDistributionInfoRequest
		resp   *types.QueryValidatorDistributionInfoResponse
		errMsg string
	}{
		{
			name: "invalid validator address",
			req: &types.QueryValidatorDistributionInfoRequest{
				ValidatorAddress: "invalid address",
			},
			resp:   &types.QueryValidatorDistributionInfoResponse{},
			errMsg: "decoding bech32 failed",
		},
		{
			name: "not a validator",
			req: &types.QueryValidatorDistributionInfoRequest{
				ValidatorAddress: addrs[0].String(),
			},
			resp:   &types.QueryValidatorDistributionInfoResponse{},
			errMsg: `expected 'cosmosvaloper' got 'cosmos'`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := queryServer.ValidatorDistributionInfo(ctx, tc.req)
			if tc.errMsg == "" {
				require.NoError(t, err)
				require.Equal(t, tc.resp, out)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, out)
			}
		})
	}
}

func TestQueryValidatorOutstandingRewards(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryValidatorCommission(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryValidatorSlashes(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryDelegationRewards(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryDelegationTotalRewards(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryDelegatorValidators(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryDelegatorWithdrawAddress(t *testing.T) {
	// TODO https://github.com/cosmos/cosmos-sdk/issues/16757
	// currently tested in tests/e2e/distribution/grpc_query_suite.go
}

func TestQueryCommunityPool(t *testing.T) {
	ctx, _, distrKeeper, dep := initFixture(t)
	queryServer := keeper.NewQuerier(distrKeeper)

	coins := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	decCoins := sdk.NewDecCoinsFromCoins(coins...)
	dep.poolKeeper.EXPECT().GetCommunityPool(gomock.Any()).Return(coins, nil).AnyTimes()

	cases := []struct {
		name   string
		req    *types.QueryCommunityPoolRequest  //nolint:staticcheck // Testing deprecated method
		resp   *types.QueryCommunityPoolResponse //nolint:staticcheck // Testing deprecated method
		errMsg string
	}{
		{
			name: "success",
			req:  &types.QueryCommunityPoolRequest{}, //nolint:staticcheck // Testing deprecated method
			resp: &types.QueryCommunityPoolResponse{ //nolint:staticcheck // Testing deprecated method
				Pool: decCoins,
			},
			errMsg: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			out, err := queryServer.CommunityPool(ctx, tc.req)
			if tc.errMsg == "" {
				require.NoError(t, err)
				require.Equal(t, tc.resp, out)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
				require.Nil(t, out)
			}
		})
	}
}
