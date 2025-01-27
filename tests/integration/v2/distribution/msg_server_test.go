package distribution

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/math"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	stakingtestutil "cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgWithdrawDelegatorReward(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	err := f.distrKeeper.FeePool.Set(f.ctx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(10000)}),
	})
	require.NoError(t, err)
	require.NoError(t, f.distrKeeper.Params.Set(f.ctx, distrtypes.DefaultParams()))

	delAddr := sdk.AccAddress(PKS[1].Address())

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	// setup staking validator
	validator, err := stakingtypes.NewValidator(f.valAddr.String(), PKS[0], stakingtypes.Description{})
	assert.NilError(t, err)
	commission := stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyOneDec(), math.LegacyOneDec())
	validator, err = validator.SetInitialCommission(commission)
	assert.NilError(t, err)
	validator.DelegatorShares = math.LegacyNewDec(100)
	validator.Tokens = math.NewInt(1000000)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.ctx, validator))

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(1000))
	err = f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	// send funds to val addr
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)

	initBalance := f.bankKeeper.GetAllBalances(f.ctx, delAddr)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	valBz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	require.NoError(t, err)
	delegation := stakingtypes.NewDelegation(delAddr.String(), validator.GetOperator(), issuedShares)
	require.NoError(t, f.stakingKeeper.SetDelegation(f.ctx, delegation))
	require.NoError(t, f.distrKeeper.DelegatorStartingInfo.Set(f.ctx, collections.Join(sdk.ValAddress(valBz), delAddr), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 20)))
	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	err = f.distrKeeper.ValidatorHistoricalRewards.Set(f.ctx, collections.Join(sdk.ValAddress(valBz), uint64(2)), historicalRewards)
	require.NoError(t, err)
	// setup current rewards and outstanding rewards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	err = f.distrKeeper.ValidatorCurrentRewards.Set(f.ctx, f.valAddr, currentRewards)
	require.NoError(t, err)
	err = f.distrKeeper.ValidatorOutstandingRewards.Set(f.ctx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})
	require.NoError(t, err)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgWithdrawDelegatorReward
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty delegator address",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: emptyDelAddr.String(),
				ValidatorAddress: f.valAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid delegator address",
		},
		{
			name: "empty validator address",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: emptyValAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "both empty addresses",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: emptyDelAddr.String(),
				ValidatorAddress: emptyValAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "delegator with no delegations",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: sdk.AccAddress([]byte("invalid")).String(),
				ValidatorAddress: f.valAddr.String(),
			},
			expErr:    true,
			expErrMsg: "not found",
		},
		{
			name: "validator with no delegations",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress(PKS[2].Address())).String(),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
		{
			name: "valid msg",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: f.valAddr.String(),
			},
			expErr: false,
		},
	}
	height := f.app.LastBlockHeight()
	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.WithdrawDelegatorReward(ctx, tc.msg)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)

			height++
			if f.app.LastBlockHeight() != height {
				panic(fmt.Errorf("expected block height to be %d, got %d", height, f.app.LastBlockHeight()))
			}

			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgWithdrawDelegatorRewardResponse)
				assert.Assert(t, ok, true)

				// check current balance is greater than initial balance
				curBalance := f.bankKeeper.GetAllBalances(f.ctx, sdk.AccAddress(f.valAddr))
				assert.Assert(t, initBalance.IsAllLTE(curBalance))
			}

			var previousTotalPower int64
			cometInfo := f.ctx.Value(corecontext.CometInfoKey).(comet.Info)
			for _, vote := range cometInfo.LastCommit.Votes {
				previousTotalPower += vote.Validator.Power
			}
			assert.Equal(t, previousTotalPower, int64(100))
		})
	}
}

func TestMsgSetWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	require.NoError(t, f.distrKeeper.Params.Set(f.ctx, distrtypes.DefaultParams()))

	delAddr := sdk.AccAddress(PKS[0].Address())
	withdrawAddr := sdk.AccAddress(PKS[1].Address())

	testCases := []struct {
		name      string
		preRun    func()
		msg       *distrtypes.MsgSetWithdrawAddress
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty delegator address",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: emptyDelAddr.String(),
				WithdrawAddress:  withdrawAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid delegator address",
		},
		{
			name: "empty withdraw address",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  emptyDelAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid withdraw address",
		},
		{
			name: "both empty addresses",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: emptyDelAddr.String(),
				WithdrawAddress:  emptyDelAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid delegator address",
		},
		{
			name: "withdraw address disabled",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = false
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  withdrawAddr.String(),
			},
			expErr:    true,
			expErrMsg: "set withdraw address disabled",
		},
		{
			name: "valid msg with same delegator and withdraw address",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  delAddr.String(),
			},
			expErr: false,
		},
		{
			name: "valid msg",
			preRun: func() {
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.ctx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  withdrawAddr.String(),
			},
			expErr: false,
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.preRun()
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.SetWithdrawAddress(ctx, tc.msg)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)

				// query the delegator withdraw address
				addr, _ := f.distrKeeper.GetDelegatorWithdrawAddr(f.ctx, delAddr)
				assert.DeepEqual(t, addr, delAddr)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgSetWithdrawAddressResponse)
				assert.Assert(t, ok, true)

				// query the delegator withdraw address
				addr, _ := f.distrKeeper.GetDelegatorWithdrawAddr(f.ctx, delAddr)
				assert.DeepEqual(t, addr.String(), tc.msg.WithdrawAddress)
			}
		})
	}
}

func TestMsgWithdrawValidatorCommission(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(1000))
	err := f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	// send funds to val addr
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	coins := sdk.NewCoins(sdk.NewCoin("mytoken", math.NewInt(2)), sdk.NewCoin("stake", math.NewInt(2)))
	err = f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, coins)
	require.NoError(t, err)

	// check initial balance
	balance := f.bankKeeper.GetAllBalances(f.ctx, sdk.AccAddress(f.valAddr))
	expTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, 1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	assert.DeepEqual(t, expCoins, balance)

	// set outstanding rewards
	err = f.distrKeeper.ValidatorOutstandingRewards.Set(f.ctx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})
	require.NoError(t, err)

	// set commission
	err = f.distrKeeper.ValidatorsAccumulatedCommission.Set(f.ctx, f.valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})
	require.NoError(t, err)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgWithdrawValidatorCommission
		expErr    bool
		expErrMsg string
	}{
		{
			name: "empty validator address",
			msg: &distrtypes.MsgWithdrawValidatorCommission{
				ValidatorAddress: emptyValAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid validator address",
		},
		{
			name: "validator with no commission",
			msg: &distrtypes.MsgWithdrawValidatorCommission{
				ValidatorAddress: sdk.ValAddress([]byte("addr1_______________")).String(),
			},
			expErr:    true,
			expErrMsg: "no validator commission to withdraw",
		},
		{
			name: "valid msg",
			msg: &distrtypes.MsgWithdrawValidatorCommission{
				ValidatorAddress: f.valAddr.String(),
			},
			expErr: false,
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.WithdrawValidatorCommission(ctx, tc.msg)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgWithdrawValidatorCommissionResponse)
				assert.Assert(t, ok, true)

				// check balance increase
				balance = f.bankKeeper.GetAllBalances(f.ctx, sdk.AccAddress(f.valAddr))
				assert.DeepEqual(t, sdk.NewCoins(
					sdk.NewCoin("mytoken", math.NewInt(1)),
					sdk.NewCoin("stake", expTokens.AddRaw(1)),
				), balance)

				// check remainder
				remainder, err := f.distrKeeper.ValidatorsAccumulatedCommission.Get(f.ctx, f.valAddr)
				require.NoError(t, err)
				assert.DeepEqual(t, sdk.DecCoins{
					sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(1).Quo(math.LegacyNewDec(4))),
					sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1).Quo(math.LegacyNewDec(2))),
				}, remainder.Commission)
			}
		})
	}
}

func TestMsgFundCommunityPool(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	addr := sdk.AccAddress(PKS[0].Address())
	addr2 := sdk.AccAddress(PKS[1].Address())
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))

	poolAcc := f.authKeeper.GetModuleAccount(f.ctx, pooltypes.ModuleName)

	// check that the pool account balance is empty
	assert.Assert(t, f.bankKeeper.GetAllBalances(f.ctx, poolAcc.GetAddress()).Empty())

	// fund the account by minting and sending amount from distribution module to addr
	err := f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, amount)
	assert.NilError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, addr, amount)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgFundCommunityPool //nolint:staticcheck // we're using a deprecated call
		expErr    bool
		expErrMsg string
	}{
		{
			name: "no depositor address",
			msg: &distrtypes.MsgFundCommunityPool{ //nolint:staticcheck // we're using a deprecated call
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: emptyDelAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid depositor address",
		},
		{
			name: "invalid coin",
			msg: &distrtypes.MsgFundCommunityPool{ //nolint:staticcheck // we're using a deprecated call
				Amount:    sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("stake", 10)},
				Depositor: addr.String(),
			},
			expErr:    true,
			expErrMsg: "10stake,10stake: invalid coins",
		},
		{
			name: "depositor address with no funds",
			msg: &distrtypes.MsgFundCommunityPool{ //nolint:staticcheck // we're using a deprecated call
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: addr2.String(),
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgFundCommunityPool{ //nolint:staticcheck // we're using a deprecated call
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: addr.String(),
			},
			expErr: false,
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					res, err := msgServer.FundCommunityPool(ctx, tc.msg) //nolint:staticcheck // we're using a deprecated call
					return res, err
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgFundCommunityPoolResponse) //nolint:staticcheck // we're using a deprecated call
				assert.Assert(t, ok, true)

				// query the community pool funds
				poolBal := f.bankKeeper.GetAllBalances(f.ctx, poolAcc.GetAddress())
				assert.Assert(t, poolBal.Equal(amount))

				assert.Assert(t, f.bankKeeper.GetAllBalances(f.ctx, addr).Empty())
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	// default params
	communityTax := math.LegacyNewDecWithPrec(2, 2) // 2%
	withdrawAddrEnabled := true

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &distrtypes.MsgUpdateParams{
				Authority: "invalid",
				Params: distrtypes.Params{
					CommunityTax:        math.LegacyNewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "community tax is nil",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        math.LegacyDec{},
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax must be not nil",
		},
		{
			name: "community tax > 1",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        math.LegacyNewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax too large: 2.000000000000000000",
		},
		{
			name: "negative community tax",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        math.LegacyNewDecWithPrec(-2, 1),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax must be positive: -0.200000000000000000",
		},
		{
			name: "base proposer reward set",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  math.LegacyNewDecWithPrec(1, 2),
					BonusProposerReward: math.LegacyZeroDec(),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "cannot update base or bonus proposer reward because these are deprecated fields",
		},
		{
			name: "bonus proposer reward set",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyNewDecWithPrec(1, 2),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "cannot update base or bonus proposer reward because these are deprecated fields",
		},
		{
			name: "all good",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  math.LegacyZeroDec(),
					BonusProposerReward: math.LegacyZeroDec(),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr: false,
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.UpdateParams(ctx, tc.msg)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgUpdateParamsResponse)
				assert.Assert(t, ok, true)

				// query the params and verify it has been updated
				params, _ := f.distrKeeper.Params.Get(f.ctx)
				assert.DeepEqual(t, distrtypes.DefaultParams(), params)
			}
		})
	}
}

func TestMsgCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(100))
	err := f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)

	// fund pool module account
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	poolAcc := f.authKeeper.GetModuleAccount(f.ctx, pooltypes.ModuleName)
	err = f.bankKeeper.SendCoinsFromModuleToModule(f.ctx, distrtypes.ModuleName, poolAcc.GetName(), amount)
	require.NoError(t, err)

	// query the community pool to verify it has been updated with balance
	poolBal := f.bankKeeper.GetAllBalances(f.ctx, poolAcc.GetAddress())
	assert.Assert(t, poolBal.Equal(amount))

	recipient := sdk.AccAddress([]byte("addr1"))

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgCommunityPoolSpend //nolint:staticcheck // we're using a deprecated call
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &distrtypes.MsgCommunityPoolSpend{ //nolint:staticcheck // we're using a deprecated call
				Authority: "invalid",
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid recipient",
			msg: &distrtypes.MsgCommunityPoolSpend{ //nolint:staticcheck // we're using a deprecated call
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgCommunityPoolSpend{ //nolint:staticcheck // we're using a deprecated call
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr: false,
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					res, e := msgServer.CommunityPoolSpend(ctx, tc.msg) //nolint:staticcheck // we're using a deprecated call
					return res, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgCommunityPoolSpendResponse) //nolint:staticcheck // we're using a deprecated call
				assert.Assert(t, ok, true)

				// query the community pool to verify it has been updated
				poolBal := f.bankKeeper.GetAllBalances(f.ctx, poolAcc.GetAddress())
				assert.Assert(t, poolBal.Empty())

			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	t.Parallel()
	f := createTestFixture(t)

	require.NoError(t, f.distrKeeper.Params.Set(f.ctx, distrtypes.DefaultParams()))
	err := f.distrKeeper.FeePool.Set(f.ctx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	require.NoError(t, err)
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(10000))
	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	// Set default staking params
	require.NoError(t, f.stakingKeeper.Params.Set(f.ctx, stakingtypes.DefaultParams()))

	addr := sdk.AccAddress("addr")
	addr1 := sdk.AccAddress(PKS[0].Address())
	valAddr1 := sdk.ValAddress(addr1)

	// send funds to val addr
	tokens := f.stakingKeeper.TokensFromConsensusPower(f.ctx, int64(1000))
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, sdk.AccAddress(valAddr1), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))
	require.NoError(t, err)
	// send funds from module to addr to perform DepositValidatorRewardsPool
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))
	f.authKeeper.SetAccount(f.ctx, f.authKeeper.NewAccountWithAddress(f.ctx, sdk.AccAddress(valAddr1)))
	require.NoError(t, err)

	tstaking := stakingtestutil.NewHelper(t, f.ctx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr1, valConsPk0, math.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	require.NoError(t, f.bankKeeper.MintCoins(f.ctx, distrtypes.ModuleName, amt))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, distrtypes.ModuleName, addr, amt))

	bondDenom, err := f.stakingKeeper.BondDenom(f.ctx)
	require.NoError(t, err)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgDepositValidatorRewardsPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "happy path (staking token)",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Depositor:        addr.String(),
				ValidatorAddress: valAddr1.String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(100))),
			},
		},
		{
			name: "happy path (non-staking token)",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Depositor:        addr.String(),
				ValidatorAddress: valAddr1.String(),
				Amount:           amt,
			},
		},
		{
			name: "invalid validator",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Depositor:        addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("addr1_______________")).String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
	}

	msgServer := distrkeeper.NewMsgServerImpl(f.distrKeeper)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				t,
				f.ctx,
				func(ctx context.Context) (transaction.Msg, error) {
					resp, e := msgServer.DepositValidatorRewardsPool(ctx, tc.msg)
					return resp, e
				},
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				_, ok := res.(*distrtypes.MsgDepositValidatorRewardsPoolResponse)
				assert.Assert(t, ok, true)

				val, err := sdk.ValAddressFromBech32(tc.msg.ValidatorAddress)
				assert.NilError(t, err)

				// check validator outstanding rewards
				outstandingRewards, err := f.distrKeeper.ValidatorOutstandingRewards.Get(f.ctx, val)
				assert.NilError(t, err)
				for _, c := range tc.msg.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					assert.DeepEqual(t, x, math.LegacyNewDecFromInt(c.Amount))
				}

			}
		})
	}
}
