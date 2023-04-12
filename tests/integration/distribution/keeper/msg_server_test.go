package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtestutil "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	distrKeeper   distrkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper

	addr    sdk.AccAddress
	valAddr sdk.ValAddress
}

func initFixture(t testing.TB) *fixture {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, stakingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, distribution.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, types.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		distrtypes.ModuleName:          {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		keys[banktypes.StoreKey],
		accountKeeper,
		blockedAddresses,
		authority.String(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, keys[stakingtypes.StoreKey], accountKeeper, bankKeeper, authority.String())

	distrKeeper := distrkeeper.NewKeeper(
		cdc, keys[distrtypes.StoreKey], accountKeeper, bankKeeper, stakingKeeper, distrtypes.ModuleName, authority.String(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, nil)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

	// set proposer and vote infos
	ctx := newCtx.WithProposer(valConsAddr).WithVoteInfos([]cmtabcitypes.VoteInfo{
		{
			Validator: cmtabcitypes.Validator{
				Address: valAddr,
				Power:   100,
			},
			SignedLastBlock: true,
		},
	})

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc, authModule, bankModule, stakingModule, distrModule)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(distrKeeper))

	return &fixture{
		app:           integrationApp,
		sdkCtx:        sdkCtx,
		cdc:           cdc,
		keys:          keys,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		stakingKeeper: stakingKeeper,
		addr:          addr,
		valAddr:       valAddr,
	}
}

func TestMsgWithdrawDelegatorReward(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	f.distrKeeper.SetFeePool(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(10000)}),
	})
	f.distrKeeper.SetParams(f.sdkCtx, distrtypes.DefaultParams())
	initFeePool := f.distrKeeper.GetFeePool(f.sdkCtx)

	delAddr := sdk.AccAddress(PKS[1].Address())
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	// setup staking validator
	validator, err := stakingtypes.NewValidator(f.valAddr, PKS[0], stakingtypes.Description{})
	assert.NilError(t, err)
	commission := stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyOneDec(), math.LegacyOneDec())
	validator, err = validator.SetInitialCommission(commission)
	assert.NilError(t, err)
	validator.DelegatorShares = math.LegacyNewDec(100)
	validator.Tokens = sdk.NewInt(1000000)
	f.stakingKeeper.SetValidator(f.sdkCtx, validator)

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// send funds to val addr
	f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	initBalance := f.bankKeeper.GetAllBalances(f.sdkCtx, delAddr)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := validator.AddTokensFromDel(delTokens)
	delegation := stakingtypes.NewDelegation(delAddr, validator.GetOperator(), issuedShares)
	f.stakingKeeper.SetDelegation(f.sdkCtx, delegation)
	f.distrKeeper.SetDelegatorStartingInfo(f.sdkCtx, validator.GetOperator(), delAddr, distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 20))

	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	f.distrKeeper.SetValidatorHistoricalRewards(f.sdkCtx, validator.GetOperator(), 2, historicalRewards)
	// setup current rewards and outstanding rewards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	f.distrKeeper.SetValidatorCurrentRewards(f.sdkCtx, f.valAddr, currentRewards)
	f.distrKeeper.SetValidatorOutstandingRewards(f.sdkCtx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})
	initOutstandingRewards := f.distrKeeper.GetValidatorOutstandingRewardsCoins(f.sdkCtx, f.valAddr)

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
			expErrMsg: "no delegation distribution info",
		},
		{
			name: "validator with no delegations",
			msg: &distrtypes.MsgWithdrawDelegatorReward{
				DelegatorAddress: delAddr.String(),
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress(PKS[2].Address())).String(),
			},
			expErr:    true,
			expErrMsg: "no validator distribution info",
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
	require.Panics(t, func() {
		f.distrKeeper.GetPreviousProposerConsAddr(f.sdkCtx)
	})
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)

			height++
			if f.app.LastBlockHeight() != height {
				panic(fmt.Errorf("expected block height to be %d, got %d", height, f.app.LastBlockHeight()))
			}

			prevProposerConsAddr := f.distrKeeper.GetPreviousProposerConsAddr(f.sdkCtx)
			assert.Assert(t, prevProposerConsAddr.Empty() == false)
			assert.DeepEqual(t, prevProposerConsAddr, valConsAddr)
			var previousTotalPower int64
			for _, voteInfo := range f.sdkCtx.VoteInfos() {
				previousTotalPower += voteInfo.Validator.Power
			}
			assert.Equal(t, previousTotalPower, int64(100))

			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgWithdrawDelegatorRewardResponse{}
				err := f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// check current balance is greater than initial balance
				curBalance := f.bankKeeper.GetAllBalances(f.sdkCtx, sdk.AccAddress(f.valAddr))
				assert.Assert(t, initBalance.IsAllLTE(curBalance))

				// check rewards
				curFeePool := f.distrKeeper.GetFeePool(f.sdkCtx)
				rewards := curFeePool.GetCommunityPool().Sub(initFeePool.CommunityPool)
				curOutstandingRewards := f.distrKeeper.GetValidatorOutstandingRewards(f.sdkCtx, f.valAddr)
				assert.DeepEqual(t, rewards, initOutstandingRewards.Sub(curOutstandingRewards.Rewards))
			}
		})
	}
}

func TestMsgSetWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	f.distrKeeper.SetParams(f.sdkCtx, distrtypes.DefaultParams())

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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = false
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
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
				params := f.distrKeeper.GetParams(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.SetParams(f.sdkCtx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  withdrawAddr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.preRun()
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)

				// query the delegator withdraw address
				addr := f.distrKeeper.GetDelegatorWithdrawAddr(f.sdkCtx, delAddr)
				assert.DeepEqual(t, addr, delAddr)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgSetWithdrawAddressResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the delegator withdraw address
				addr := f.distrKeeper.GetDelegatorWithdrawAddr(f.sdkCtx, delAddr)
				assert.DeepEqual(t, addr.String(), tc.msg.WithdrawAddress)
			}
		})
	}
}

func TestMsgWithdrawValidatorCommission(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(5).Quo(math.LegacyNewDec(4))),
		sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(3).Quo(math.LegacyNewDec(2))),
	}

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// send funds to val addr
	f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	coins := sdk.NewCoins(sdk.NewCoin("mytoken", sdk.NewInt(2)), sdk.NewCoin("stake", sdk.NewInt(2)))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, coins)

	// check initial balance
	balance := f.bankKeeper.GetAllBalances(f.sdkCtx, sdk.AccAddress(f.valAddr))
	expTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	assert.DeepEqual(t, expCoins, balance)

	// set outstanding rewards
	f.distrKeeper.SetValidatorOutstandingRewards(f.sdkCtx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})

	// set commission
	f.distrKeeper.SetValidatorAccumulatedCommission(f.sdkCtx, f.valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})

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

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgWithdrawValidatorCommissionResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// check balance increase
				balance = f.bankKeeper.GetAllBalances(f.sdkCtx, sdk.AccAddress(f.valAddr))
				assert.DeepEqual(t, sdk.NewCoins(
					sdk.NewCoin("mytoken", sdk.NewInt(1)),
					sdk.NewCoin("stake", expTokens.AddRaw(1)),
				), balance)

				// check remainder
				remainder := f.distrKeeper.GetValidatorAccumulatedCommission(f.sdkCtx, f.valAddr).Commission
				assert.DeepEqual(t, sdk.DecCoins{
					sdk.NewDecCoinFromDec("mytoken", math.LegacyNewDec(1).Quo(math.LegacyNewDec(4))),
					sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1).Quo(math.LegacyNewDec(2))),
				}, remainder)
			}
		})

	}
}

func TestMsgFundCommunityPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// reset fee pool
	f.distrKeeper.SetFeePool(f.sdkCtx, distrtypes.InitialFeePool())
	initPool := f.distrKeeper.GetFeePool(f.sdkCtx)
	assert.Assert(t, initPool.CommunityPool.Empty())

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	addr := sdk.AccAddress(PKS[0].Address())
	addr2 := sdk.AccAddress(PKS[1].Address())
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))

	// fund the account by minting and sending amount from distribution module to addr
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amount)
	assert.NilError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, amount)
	assert.NilError(t, err)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgFundCommunityPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "no depositor address",
			msg: &distrtypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
				Depositor: emptyDelAddr.String(),
			},
			expErr:    true,
			expErrMsg: "invalid depositor address",
		},
		{
			name: "invalid coin",
			msg: &distrtypes.MsgFundCommunityPool{
				Amount:    sdk.Coins{sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("stake", 10)},
				Depositor: addr.String(),
			},
			expErr:    true,
			expErrMsg: "10stake,10stake: invalid coins",
		},
		{
			name: "depositor address with no funds",
			msg: &distrtypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
				Depositor: addr2.String(),
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
				Depositor: addr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgFundCommunityPool{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the community pool funds
				assert.DeepEqual(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), f.distrKeeper.GetFeePool(f.sdkCtx).CommunityPool)
				assert.Assert(t, f.bankKeeper.GetAllBalances(f.sdkCtx, addr).Empty())
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// default params
	communityTax := sdk.NewDecWithPrec(2, 2) // 2%
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
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "community tax > 1",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "negative community tax",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        sdk.NewDecWithPrec(-2, 1),
					WithdrawAddrEnabled: withdrawAddrEnabled,
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward set",
			msg: &distrtypes.MsgUpdateParams{
				Authority: f.distrKeeper.GetAuthority(),
				Params: distrtypes.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(1, 2),
					BonusProposerReward: sdk.ZeroDec(),
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
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.NewDecWithPrec(1, 2),
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
					BaseProposerReward:  sdk.ZeroDec(),
					BonusProposerReward: sdk.ZeroDec(),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgUpdateParams{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the params and verify it has been updated
				params := f.distrKeeper.GetParams(f.sdkCtx)
				assert.DeepEqual(t, distrtypes.DefaultParams(), params)
			}
		})
	}
}

func TestMsgCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	f.distrKeeper.SetParams(f.sdkCtx, distrtypes.DefaultParams())
	f.distrKeeper.SetFeePool(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(10000)}),
	})
	initialFeePool := f.distrKeeper.GetFeePool(f.sdkCtx)

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	recipient := sdk.AccAddress([]byte("addr1"))

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgCommunityPoolSpend
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: "invalid",
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid recipient",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgCommunityPoolSpend{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the community pool to verify it has been updated
				communityPool := f.distrKeeper.GetFeePoolCommunityCoins(f.sdkCtx)
				newPool, negative := initialFeePool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(tc.msg.Amount...))
				assert.Assert(t, negative == false)
				assert.DeepEqual(t, communityPool, newPool)
			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	f.distrKeeper.SetParams(f.sdkCtx, distrtypes.DefaultParams())
	f.distrKeeper.SetFeePool(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(10000))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// Set default staking params
	f.stakingKeeper.SetParams(f.sdkCtx, stakingtypes.DefaultParams())

	addr := sdk.AccAddress([]byte("addr"))
	addr1 := sdk.AccAddress(PKS[0].Address())
	valAddr1 := sdk.ValAddress(addr1)

	// send funds to val addr
	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(valAddr1), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))

	// send funds from module to addr to perform DepositValidatorRewardsPool
	f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))

	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr1, valConsPk0, sdk.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amt)
	f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, amt)

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
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(f.sdkCtx), sdk.NewInt(100))),
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
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(f.sdkCtx), sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticBeginEndBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgDepositValidatorRewardsPoolResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				val, err := sdk.ValAddressFromBech32(tc.msg.ValidatorAddress)
				assert.NilError(t, err)

				// check validator outstanding rewards
				outstandingRewards := f.distrKeeper.GetValidatorOutstandingRewards(f.sdkCtx, val)
				for _, c := range tc.msg.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					assert.DeepEqual(t, x, sdk.NewDecFromInt(c.Amount))
				}

			}
		})
	}
}
