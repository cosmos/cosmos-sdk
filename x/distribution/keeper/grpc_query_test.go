package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGRPCParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	f.distrKeeper.SetParams(integrationApp.SDKContext(), types.DefaultParams())

	qr := integrationApp.QueryHelper()
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
					CommunityTax:        sdk.NewDecWithPrec(3, 1),
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
					WithdrawAddrEnabled: true,
				}

				assert.NilError(t, f.distrKeeper.SetParams(integrationApp.SDKContext(), params))
				expParams = params
			},
			msg: &types.QueryParamsRequest{},
		},
	}
	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			tc.malleate()

			paramsRes, err := queryClient.Params(integrationApp.SDKContext(), tc.msg)
			assert.NilError(t, err)
			assert.Assert(t, paramsRes != nil)
			assert.DeepEqual(t, paramsRes.Params, expParams)
		})

	}
}

func TestGRPCValidatorOutstandingRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5000)),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300)),
	}

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)

	// set outstanding rewards
	f.distrKeeper.SetValidatorOutstandingRewards(integrationApp.SDKContext(), valAddr, types.ValidatorOutstandingRewards{Rewards: valCommission})
	rewards := f.distrKeeper.GetValidatorOutstandingRewards(integrationApp.SDKContext(), valAddr)

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
			name:    "valid request",
			msg:     &types.QueryValidatorOutstandingRewardsRequest{ValidatorAddress: valAddr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			validatorOutstandingRewards, err := queryClient.ValidatorOutstandingRewards(integrationApp.SDKContext(), tc.msg)

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

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)

	commission := sdk.DecCoins{{Denom: "token1", Amount: math.LegacyNewDec(4)}, {Denom: "token2", Amount: math.LegacyNewDec(2)}}
	f.distrKeeper.SetValidatorAccumulatedCommission(integrationApp.SDKContext(), valAddr, types.ValidatorAccumulatedCommission{Commission: commission})

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
			name:    "valid request",
			msg:     &types.QueryValidatorCommissionRequest{ValidatorAddress: valAddr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			commissionRes, err := queryClient.ValidatorCommission(integrationApp.SDKContext(), tc.msg)

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

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	addr2 := sdk.AccAddress(PKS[1].Address())
	valAddr2 := sdk.ValAddress(addr2)

	slashes := []types.ValidatorSlashEvent{
		types.NewValidatorSlashEvent(3, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(5, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(7, sdk.NewDecWithPrec(5, 1)),
		types.NewValidatorSlashEvent(9, sdk.NewDecWithPrec(5, 1)),
	}

	for i, slash := range slashes {
		f.distrKeeper.SetValidatorSlashEvent(integrationApp.SDKContext(), valAddr, uint64(i+2), 0, slash)
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
					ValidatorAddress: valAddr.String(),
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
					ValidatorAddress: valAddr.String(),
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
					ValidatorAddress: valAddr.String(),
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

			slashesRes, err := queryClient.ValidatorSlashes(integrationApp.SDKContext(), req)

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

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	f.distrKeeper.SetParams(integrationApp.SDKContext(), types.DefaultParams())

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := sdk.AccAddress(PKS[0].Address())
	addr2 := sdk.AccAddress(PKS[1].Address())

	err := f.distrKeeper.SetWithdrawAddr(integrationApp.SDKContext(), addr, addr2)
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
			msg:     &types.QueryDelegatorWithdrawAddressRequest{DelegatorAddress: addr.String()},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			withdrawAddress, err := queryClient.DelegatorWithdrawAddress(integrationApp.SDKContext(), tc.msg)

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

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	f.distrKeeper.SetFeePool(integrationApp.SDKContext(), types.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(0)}),
	})

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := sdk.AccAddress(PKS[0].Address())

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
				assert.NilError(t, f.bankKeeper.MintCoins(integrationApp.SDKContext(), types.ModuleName, amount))
				assert.NilError(t, f.bankKeeper.SendCoinsFromModuleToAccount(integrationApp.SDKContext(), types.ModuleName, addr, amount))

				err := f.distrKeeper.FundCommunityPool(integrationApp.SDKContext(), amount, addr)
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

			pool, err := queryClient.CommunityPool(integrationApp.SDKContext(), req)

			assert.NilError(t, err)
			assert.DeepEqual(t, expPool, pool)
		})
	}
}

func TestGRPCDelegationRewards(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.NewIntegrationApp(t.Name(), log.NewTestLogger(t), f.keys, authModule, bankModule, stakingModule, distrModule)

	f.distrKeeper.SetFeePool(integrationApp.SDKContext(), types.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(1000)}),
	})

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(integrationApp.SDKContext(), int64(1000))
	f.bankKeeper.MintCoins(integrationApp.SDKContext(), types.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// Set default staking params
	f.stakingKeeper.SetParams(integrationApp.SDKContext(), stakingtypes.DefaultParams())

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	addr2 := sdk.AccAddress(PKS[1].Address())
	valAddr2 := sdk.ValAddress(addr2)
	delAddr := sdk.AccAddress(PKS[2].Address())

	// send funds to val addr
	funds := f.stakingKeeper.TokensFromConsensusPower(integrationApp.SDKContext(), int64(1000))
	f.bankKeeper.SendCoinsFromModuleToAccount(integrationApp.SDKContext(), types.ModuleName, sdk.AccAddress(valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, funds)))

	initialStake := int64(10)
	tstaking := stakingtestutil.NewHelper(t, integrationApp.SDKContext(), f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr, valConsPk0, sdk.NewInt(initialStake), true)

	val, found := f.stakingKeeper.GetValidator(integrationApp.SDKContext(), valAddr)
	assert.Assert(t, found)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := val.AddTokensFromDel(delTokens)
	delegation := stakingtypes.NewDelegation(delAddr, valAddr, issuedShares)
	f.stakingKeeper.SetDelegation(integrationApp.SDKContext(), delegation)
	f.distrKeeper.SetDelegatorStartingInfo(integrationApp.SDKContext(), validator.GetOperator(), delAddr, types.NewDelegatorStartingInfo(2, math.LegacyNewDec(initialStake), 20))

	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := types.NewValidatorHistoricalRewards(decCoins, 2)
	f.distrKeeper.SetValidatorHistoricalRewards(integrationApp.SDKContext(), validator.GetOperator(), 2, historicalRewards)
	// setup current rewards and outstanding rewards
	currentRewards := types.NewValidatorCurrentRewards(decCoins, 3)
	f.distrKeeper.SetValidatorCurrentRewards(integrationApp.SDKContext(), valAddr, currentRewards)
	f.distrKeeper.SetValidatorOutstandingRewards(integrationApp.SDKContext(), valAddr, types.ValidatorOutstandingRewards{Rewards: decCoins})

	expRes := &types.QueryDelegationRewardsResponse{
		Rewards: sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: math.LegacyNewDec(initialStake / 10)}},
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
				ValidatorAddress: valAddr.String(),
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
				ValidatorAddress: valAddr.String(),
			},
			expPass: true,
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			rewards, err := queryClient.DelegationRewards(integrationApp.SDKContext(), tc.msg)

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
