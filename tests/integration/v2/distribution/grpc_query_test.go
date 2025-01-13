package distribution

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"cosmossdk.io/x/distribution/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func TestGRPCParams(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, types.DefaultParams()))

	var (
		params    types.Params
		expParams types.Params
	)

	testCases := []struct {
		name      string
		malleate  func()
		msg       *types.QueryParamsRequest
		expErrMsg string
	}{
		{
			name: "empty params request",
			malleate: func() {
				expParams = types.DefaultParams()
			},
			msg: &types.QueryParamsRequest{},
		},
		{
			name: "valid request",
			malleate: func() {
				params = types.Params{
					CommunityTax:        math.LegacyNewDecWithPrec(3, 1),
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
					WithdrawAddrEnabled: true,
				}

				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
				expParams = params
			},
			msg: &types.QueryParamsRequest{},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			tc.malleate()

			paramsRes, err := f.queryClient.Params(f.ctx, tc.msg)
			assert.NilError(t, err)
			assert.Assert(t, paramsRes != nil)
			assert.DeepEqual(t, paramsRes.Params, expParams)
		})

	}
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, types.DefaultParams()))
	setupValidatorWithCommission(t, f, f.valAddr, 10) // Setup a validator with commission

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5000)),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300)),
	}

	// set outstanding rewards
	err := f.distrKeeper.ValidatorOutstandingRewards.Set(f.ctx, f.valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
	assert.NilError(t, err)

	rewards, err := f.distrKeeper.ValidatorOutstandingRewards.Get(f.ctx, f.valAddr)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		msg       *types.QueryValidatorOutstandingRewardsRequest
		expPass   bool
		expErrMsg string
	}{
		{
			name:      "empty request",
			msg:       &types.QueryValidatorOutstandingRewardsRequest{},
			expPass:   false,
			expErrMsg: "empty validator address",
		},
		{
			name:      "invalid address",
			msg:       &types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: sdk.ValAddress("addr1_______________").String()},
			expPass:   false,
			expErrMsg: "validator does not exist",
		},
		{
			name:    "valid request",
			msg:     &types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: f.valAddr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			validatorOutstandingRewards, err := f.queryClient.ValidatorOutstandingRewards(f.ctx, tc.msg)

			if tc.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, rewards, validatorOutstandingRewards.Rewards)
				assert.DeepEqual(t, valCommission, validatorOutstandingRewards.Rewards.Rewards)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, validatorOutstandingRewards == nil)
			}
		})
	}
}

func TestGRPCValidatorCommission(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, types.DefaultParams())) // Set default distribution parameters
	setupValidatorWithCommission(t, f, f.valAddr, 10)                          // Setup a validator with commission

	commission := sdk.DecCoins{sdk.DecCoin{Denom: "token1", Amount: math.LegacyNewDec(4)}, {Denom: "token2", Amount: math.LegacyNewDec(2)}}
	assert.NilError(t, f.distrKeeper.ValidatorsAccumulatedCommission.Set(f.ctx, f.valAddr, types.ValidatorAccumulatedCommission{Commission: commission}))

	testCases := []struct {
		name      string
		msg       *types.QueryValidatorCommissionRequest
		expPass   bool
		expErrMsg string
	}{
		{
			name:      "empty request",
			msg:       &types.QueryValidatorCommissionRequest{},
			expPass:   false,
			expErrMsg: "empty validator address",
		},
		{
			name:      "invalid validator",
			msg:       &types.QueryValidatorCommissionRequest{ValidatorAddress: sdk.ValAddress("addr1_______________").String()},
			expPass:   false,
			expErrMsg: "validator does not exist",
		},
		{
			name:    "valid request",
			msg:     &types.QueryValidatorCommissionRequest{ValidatorAddress: f.valAddr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			commissionRes, err := f.queryClient.ValidatorCommission(f.ctx, tc.msg)

			if tc.expPass {
				assert.NilError(t, err)
				assert.Assert(t, commissionRes != nil)
				assert.DeepEqual(t, commissionRes.Commission.Commission, commission)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, commissionRes == nil)
			}
		})
	}
}

func TestGRPCValidatorSlashes(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	addr2 := sdk.AccAddress(PKS[1].Address())
	valAddr2 := sdk.ValAddress(addr2)

	slashes := []types.ValidatorSlashEvent{
		types.NewValidatorSlashEvent(3, math.LegacyNewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(5, math.LegacyNewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(7, math.LegacyNewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(9, math.LegacyNewDecWithPrec(5, 1)),
	}

	for i, slash := range slashes {
		err := f.distrKeeper.ValidatorSlashEvents.Set(
			f.ctx,
			collections.Join3(f.valAddr, uint64(i+2), uint64(0)),
			slash,
		)
		assert.NilError(t, err)
	}

	var (
		req    *types.QueryValidatorSlashesRequest
		express *types.QueryValidatorSlashesResponse
	)

	testCases := []struct {
		name      string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			name: "empty request",
			malleate: func() {
				req = &types.QueryValidatorSlashesRequest{}
				express = &types.QueryValidatorSlashesResponse{}
			},
			expPass:   false,
			expErrMsg: "empty validator address",
		},
		{
			name: "Ending height lesser than start height request",
			malleate: func() {
				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddr2.String(),
					StartingHeight:   10,
					EndingHeight:     1,
				}
				express = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
			},
			expPass:   false,
			expErrMsg: "starting height greater than ending height",
		},
		{
			name: "no slash event validator request",
			malleate: func() {
				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: valAddr2.String(),
					StartingHeight:   1,
					EndingHeight:     10,
				}
				express = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
			},
			expPass: true,
		},
		{
			name: "request slashes with offset 2 and limit 2",
			malleate: func() {
				pageReq := &query.PageRequest{
					Offset: 2,
					Limit:  2,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: f.valAddr.String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				express = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[2:],
				}
			},
			expPass: true,
		},
		{
			name: "request slashes with page limit 3 and count total",
			malleate: func() {
				pageReq := &query.PageRequest{
					Limit:      3,
					CountTotal: true,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: f.valAddr.String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				express = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[:3],
				}
			},
			expPass: true,
		},
		{
			name: "request slashes with page limit 4 and count total",
			malleate: func() {
				pageReq := &query.PageRequest{
					Limit:      4,
					CountTotal: true,
				}

				req = &types.QueryValidatorSlashesRequest{
					ValidatorAddress: f.valAddr.String(),
					StartingHeight:   1,
					EndingHeight:     10,
					Pagination:       pageReq,
				}

				express = &types.QueryValidatorSlashesResponse{
					Slashes: slashes[:4],
				}
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			tc.malleate()

			slashesRes, err := f.queryClient.ValidatorSlashes(f.ctx, req)

			if tc.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, express.GetSlashes(), slashesRes.GetSlashes())
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, slashesRes == nil)
			}
		})
	}
}

func TestGRPCDelegatorWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, types.DefaultParams()))

	addr2 := sdk.AccAddress(PKS[1].Address())

	err := f.distrKeeper.SetWithdrawAddr(f.ctx, f.addr, addr2)
	assert.Assert(t, err == nil)

	testCases := []struct {
		name      string
		msg       *types.QueryDelegatorWithdrawAddressRequest
		expPass   bool
		expErrMsg string
	}{
		{
			name:      "empty request",
			msg:       &types.QueryDelegatorWithdrawAddressRequest{},
			expPass:   false,
			expErrMsg: "empty delegator address",
		},
		{
			name:    "valid request",
			msg:     &types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: f.addr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			withdrawAddress, err := f.queryClient.DelegatorWithdrawAddress(f.ctx, tc.msg)

			if tc.expPass {
				assert.NilError(t, err)
				assert.Equal(t, withdrawAddress.WithdrawAddress, addr2.String())
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, withdrawAddress == nil)
			}
		})
	}
}

func TestGRPCCommunityPool(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	var (
		req     *types.QueryCommunityPoolRequest  //nolint:staticcheck // we're using a deprecated call
		expPool *types.QueryCommunityPoolResponse //nolint:staticcheck // we're using a deprecated call
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			name: "valid request empty community pool",
			malleate: func() {
				req = &types.QueryCommunityPoolRequest{}      //nolint:staticcheck // we're using a deprecated call
				expPool = &types.QueryCommunityPoolResponse{} //nolint:staticcheck // we're using a deprecated call
			},
		},
		{
			name: "valid request",
			malleate: func() {
				amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
				assert.NilError(t, f.bankKeeper.MintCoins(f.ctx, types.ModuleName, amount))
				assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, types.ModuleName, f.addr, amount))

				err := f.poolKeeper.FundCommunityPool(f.ctx, amount, f.addr)
				assert.Assert(t, err == nil)
				req = &types.QueryCommunityPoolRequest{} //nolint:staticcheck // we're using a deprecated call

				expPool = &types.QueryCommunityPoolResponse{Pool: sdk.NewDecCoinsFromCoins(amount...)} //nolint:staticcheck // we're using a deprecated call
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			testCase.malleate()

			pool, err := f.queryClient.CommunityPool(f.ctx, req) //nolint:staticcheck // we're using a deprecated call

			assert.NilError(t, err)
			assert.DeepEqual(t, expPool, pool)
		})
	}
}

func TestGRPCDelegationRewards(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	assert.NilError(t, f.distrKeeper.FeePool.Set(f.ctx, types.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(1000)}),
	}))

	initialStake := int64(10)
	assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, types.DefaultParams()))
	setupValidatorWithCommission(t, f, f.valAddr, initialStake) // Setup a validator with commission
	val, found := f.stakingKeeper.GetValidator(f.ctx, f.valAddr)
	assert.Assert(t, found)

	// Set default staking params
	assert.NilError(t, f.stakingKeeper.Params.Set(f.ctx, stakingtypes.DefaultParams()))

	addr2 := sdk.AccAddress(PKS[1].Address())
	valAddr2 := sdk.ValAddress(addr2)
	delAddr := sdk.AccAddress(PKS[2].Address())

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := val.AddTokensFromDel(delTokens)
	delegation := stakingtypes.NewDelegation(delAddr.String(), f.valAddr.String(), issuedShares)
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.ctx, delegation))
	valBz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)
	assert.NilError(t, f.distrKeeper.DelegatorStartingInfo.Set(f.ctx, collections.Join(sdk.ValAddress(valBz), delAddr), types.NewDelegatorStartingInfo(2, math.LegacyNewDec(initialStake), 20)))

	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := types.NewValidatorHistoricalRewards(decCoins, 2)
	assert.NilError(t, f.distrKeeper.ValidatorHistoricalRewards.Set(f.ctx, collections.Join(sdk.ValAddress(valBz), uint64(2)), historicalRewards))
	// setup current rewards and outstanding rewards
	currentRewards := types.NewValidatorCurrentRewards(decCoins, 3)
	assert.NilError(t, f.distrKeeper.ValidatorCurrentRewards.Set(f.ctx, f.valAddr, currentRewards))
	assert.NilError(t, f.distrKeeper.ValidatorOutstandingRewards.Set(f.ctx, f.valAddr, types.ValidatorOutstandingRewards{Rewards: decCoins}))

	express := &types.QueryDelegationRewardsResponse{
		Rewards: sdk.DecCoins{sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initialStake / 10)}},
	}

	// test command delegation rewards grpc
	testCases := []struct {
		name      string
		msg       *types.QueryDelegationRewardsRequest
		expPass   bool
		expErrMsg string
	}{
		{
			name:      "empty request",
			msg:       &types.QueryDelegationRewardsRequest{},
			expPass:   false,
			expErrMsg: "empty delegator address",
		},
		{
			name: "empty delegator address",
			msg: &types.QueryDelegationRewardsRequest{
				DelegatorAddress: "",
				ValidatorAddress: f.valAddr.String(),
			},
			expPass:   false,
			expErrMsg: "empty delegator address",
		},
		{
			name: "empty validator address",
			msg: &types.QueryDelegationRewardsRequest{
				DelegatorAddress: addr2.String(),
				ValidatorAddress: "",
			},
			expPass:   false,
			expErrMsg: "empty validator address",
		},
		{
			name: "request with wrong delegator and validator",
			msg: &types.QueryDelegationRewardsRequest{
				DelegatorAddress: addr2.String(),
				ValidatorAddress: valAddr2.String(),
			},
			expPass:   false,
			expErrMsg: "validator does not exist",
		},
		{
			name: "valid request",
			msg: &types.QueryDelegationRewardsRequest{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: f.valAddr.String(),
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			rewards, err := f.queryClient.DelegationRewards(f.ctx, tc.msg)

			if tc.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, express, rewards)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, rewards == nil)
			}
		})
	}
}
