package keeper_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGRPCParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, types.DefaultParams()))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

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

				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
				expParams = params
			},
			msg: &types.QueryParamsRequest{},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			tc.malleate()

			paramsRes, err := queryClient.Params(f.sdkCtx, tc.msg)
			assert.NilError(t, err)
			assert.Assert(t, paramsRes != nil)
			assert.DeepEqual(t, paramsRes.Params, expParams)
		})

	}
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	// Set default staking params
	assert.NilError(t, f.stakingKeeper.SetParams(f.sdkCtx, stakingtypes.DefaultParams()))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5000)),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300)),
	}

	// send funds to val addr
	funds := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, types.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds))))

	initialStake := int64(10)
	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(f.valAddr, valConsPk0, math.NewInt(initialStake), true)

	// set outstanding rewards
	err := f.distrKeeper.ValidatorOutstandingRewards.Set(f.sdkCtx, f.valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
	assert.NilError(t, err)

	rewards, err := f.distrKeeper.ValidatorOutstandingRewards.Get(f.sdkCtx, f.valAddr)
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
			validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(f.sdkCtx, tc.msg)

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
	f := initFixture(t)

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	// Set default staking params
	assert.NilError(t, f.stakingKeeper.SetParams(f.sdkCtx, stakingtypes.DefaultParams()))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	// send funds to val addr
	funds := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, types.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds))))

	initialStake := int64(10)
	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(f.valAddr, valConsPk0, math.NewInt(initialStake), true)

	commission := sdk.DecCoins{sdk.DecCoin{Denom: "token1", Amount: math.LegacyNewDec(4)}, {Denom: "token2", Amount: math.LegacyNewDec(2)}}
	assert.NilError(t, f.distrKeeper.ValidatorsAccumulatedCommission.Set(f.sdkCtx, f.valAddr, types.ValidatorAccumulatedCommission{Commission: commission}))

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
			commissionRes, err := queryClient.ValidatorCommission(f.sdkCtx, tc.msg)

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
	f := initFixture(t)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

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
			f.sdkCtx,
			collections.Join3(f.valAddr, uint64(i+2), uint64(0)),
			slash,
		)
		assert.NilError(t, err)
	}

	var (
		req    *types.QueryValidatorSlashesRequest
		expRes *types.QueryValidatorSlashesResponse
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
				expRes = &types.QueryValidatorSlashesResponse{}
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
				expRes = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
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
				expRes = &types.QueryValidatorSlashesResponse{Pagination: &query.PageResponse{}}
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

				expRes = &types.QueryValidatorSlashesResponse{
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

				expRes = &types.QueryValidatorSlashesResponse{
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

				expRes = &types.QueryValidatorSlashesResponse{
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

			slashesRes, err := queryClient.ValidatorSlashes(f.sdkCtx, req)

			if tc.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, expRes.GetSlashes(), slashesRes.GetSlashes())
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, slashesRes == nil)
			}
		})
	}
}

func TestGRPCDelegatorWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, types.DefaultParams()))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr2 := sdk.AccAddress(PKS[1].Address())

	err := f.distrKeeper.SetWithdrawAddr(f.sdkCtx, f.addr, addr2)
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
			withdrawAddress, err := queryClient.DelegatorWithdrawAddress(f.sdkCtx, tc.msg)

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
	f := initFixture(t)

	assert.NilError(t, f.distrKeeper.FeePool.Set(f.sdkCtx, types.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(0)}),
	}))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	var (
		req     *types.QueryCommunityPoolRequest
		expPool *types.QueryCommunityPoolResponse
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			name: "valid request empty community pool",
			malleate: func() {
				req = &types.QueryCommunityPoolRequest{}
				expPool = &types.QueryCommunityPoolResponse{}
			},
		},
		{
			name: "valid request",
			malleate: func() {
				amount := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
				assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, types.ModuleName, amount))
				assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, types.ModuleName, f.addr, amount))

				err := f.distrKeeper.FundCommunityPool(f.sdkCtx, amount, f.addr)
				assert.Assert(t, err == nil)
				req = &types.QueryCommunityPoolRequest{}

				expPool = &types.QueryCommunityPoolResponse{Pool: sdk.NewDecCoinsFromCoins(amount...)}
			},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			testCase.malleate()

			pool, err := queryClient.CommunityPool(f.sdkCtx, req)

			assert.NilError(t, err)
			assert.DeepEqual(t, expPool, pool)
		})
	}
}

func TestGRPCDelegationRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	assert.NilError(t, f.distrKeeper.FeePool.Set(f.sdkCtx, types.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(1000)}),
	}))

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.MintCoins(f.sdkCtx, types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	// Set default staking params
	assert.NilError(t, f.stakingKeeper.SetParams(f.sdkCtx, stakingtypes.DefaultParams()))

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr2 := sdk.AccAddress(PKS[1].Address())
	valAddr2 := sdk.ValAddress(addr2)
	delAddr := sdk.AccAddress(PKS[2].Address())

	// send funds to val addr
	funds := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, types.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds))))

	initialStake := int64(10)
	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(f.valAddr, valConsPk0, math.NewInt(initialStake), true)

	val, found := f.stakingKeeper.GetValidator(f.sdkCtx, f.valAddr)
	assert.Assert(t, found)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := val.AddTokensFromDel(delTokens)
	delegation := stakingtypes.NewDelegation(delAddr.String(), f.valAddr.String(), issuedShares)
	assert.NilError(t, f.stakingKeeper.SetDelegation(f.sdkCtx, delegation))
	valBz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)
	assert.NilError(t, f.distrKeeper.DelegatorStartingInfo.Set(f.sdkCtx, collections.Join(sdk.ValAddress(valBz), delAddr), types.NewDelegatorStartingInfo(2, math.LegacyNewDec(initialStake), 20)))

	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := types.NewValidatorHistoricalRewards(decCoins, 2)
	assert.NilError(t, f.distrKeeper.ValidatorHistoricalRewards.Set(f.sdkCtx, collections.Join(sdk.ValAddress(valBz), uint64(2)), historicalRewards))
	// setup current rewards and outstanding rewards
	currentRewards := types.NewValidatorCurrentRewards(decCoins, 3)
	assert.NilError(t, f.distrKeeper.ValidatorCurrentRewards.Set(f.sdkCtx, f.valAddr, currentRewards))
	assert.NilError(t, f.distrKeeper.ValidatorOutstandingRewards.Set(f.sdkCtx, f.valAddr, types.ValidatorOutstandingRewards{Rewards: decCoins}))

	expRes := &types.QueryDelegationRewardsResponse{
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
			rewards, err := queryClient.DelegationRewards(f.sdkCtx, tc.msg)

			if tc.expPass {
				assert.NilError(t, err)
				assert.DeepEqual(t, expRes, rewards)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, rewards == nil)
			}
		})
	}
}
