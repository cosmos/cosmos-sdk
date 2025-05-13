package keeper_test

import (
	"fmt"
	"testing"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	types "github.com/cometbft/cometbft/api/cometbft/types/v1"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
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

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, stakingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, distribution.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
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
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))

	distrKeeper := distrkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[distrtypes.StoreKey]), accountKeeper, bankKeeper, stakingKeeper, distrtypes.ModuleName, authority.String(),
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
			BlockIdFlag: types.BlockIDFlagCommit,
		},
	})

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:    authModule,
		banktypes.ModuleName:    bankModule,
		stakingtypes.ModuleName: stakingModule,
		distrtypes.ModuleName:   distrModule,
	})

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

	err := f.distrKeeper.FeePool.Set(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(10000)}),
	})
	require.NoError(t, err)
	require.NoError(t, f.distrKeeper.Params.Set(f.sdkCtx, distrtypes.DefaultParams()))
	initFeePool, err := f.distrKeeper.FeePool.Get(f.sdkCtx)
	assert.NilError(t, err)

	delAddr := sdk.AccAddress(PKS[1].Address())
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

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
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validator))

	// set module account coins
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	err = f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	// send funds to val addr
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)

	initBalance := f.bankKeeper.GetAllBalances(f.sdkCtx, delAddr)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2, sdk.DefaultPowerReduction)
	validator, issuedShares := validator.AddTokensFromDel(delTokens)

	valBz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	require.NoError(t, err)
	delegation := stakingtypes.NewDelegation(delAddr.String(), validator.GetOperator(), issuedShares)
	require.NoError(t, f.stakingKeeper.SetDelegation(f.sdkCtx, delegation))
	require.NoError(t, f.distrKeeper.SetDelegatorStartingInfo(f.sdkCtx, valBz, delAddr, distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 20)))
	// setup validator rewards
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	err = f.distrKeeper.SetValidatorHistoricalRewards(f.sdkCtx, valBz, 2, historicalRewards)
	require.NoError(t, err)
	// setup current rewards and outstanding rewards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	err = f.distrKeeper.SetValidatorCurrentRewards(f.sdkCtx, f.valAddr, currentRewards)
	require.NoError(t, err)
	err = f.distrKeeper.SetValidatorOutstandingRewards(f.sdkCtx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})
	require.NoError(t, err)
	initOutstandingRewards, err := f.distrKeeper.GetValidatorOutstandingRewardsCoins(f.sdkCtx, f.valAddr)
	assert.NilError(t, err)

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
			expErrMsg: "no delegation for (address, validator) tuple",
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

	proposerAddr, _ := f.distrKeeper.GetPreviousProposerConsAddr(f.sdkCtx)
	assert.Equal(t, proposerAddr.Empty(), true)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
				result := distrtypes.MsgWithdrawDelegatorRewardResponse{}
				err := f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// check current balance is greater than initial balance
				curBalance := f.bankKeeper.GetAllBalances(f.sdkCtx, sdk.AccAddress(f.valAddr))
				assert.Assert(t, initBalance.IsAllLTE(curBalance))

				// check rewards
				curFeePool, _ := f.distrKeeper.FeePool.Get(f.sdkCtx)
				rewards := curFeePool.GetCommunityPool().Sub(initFeePool.CommunityPool)
				curOutstandingRewards, _ := f.distrKeeper.GetValidatorOutstandingRewards(f.sdkCtx, f.valAddr)
				assert.DeepEqual(t, rewards, initOutstandingRewards.Sub(curOutstandingRewards.Rewards))
			}

			prevProposerConsAddr, err := f.distrKeeper.GetPreviousProposerConsAddr(f.sdkCtx)
			assert.NilError(t, err)
			assert.Assert(t, prevProposerConsAddr.Empty() == false)
			assert.DeepEqual(t, prevProposerConsAddr, valConsAddr)
			var previousTotalPower int64
			for _, voteInfo := range f.sdkCtx.VoteInfos() {
				previousTotalPower += voteInfo.Validator.Power
			}
			assert.Equal(t, previousTotalPower, int64(100))
		})
	}
}

func TestMsgSetWithdrawAddress(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	require.NoError(t, f.distrKeeper.Params.Set(f.sdkCtx, distrtypes.DefaultParams()))

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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = false
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				params.WithdrawAddrEnabled = true
				assert.NilError(t, f.distrKeeper.Params.Set(f.sdkCtx, params))
			},
			msg: &distrtypes.MsgSetWithdrawAddress{
				DelegatorAddress: delAddr.String(),
				WithdrawAddress:  withdrawAddr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.preRun()
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
				integration.WithAutomaticCommit(),
			)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)

				// query the delegator withdraw address
				addr, _ := f.distrKeeper.GetDelegatorWithdrawAddr(f.sdkCtx, delAddr)
				assert.DeepEqual(t, addr, delAddr)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgSetWithdrawAddressResponse{}
				err = f.cdc.Unmarshal(res.Value, &result)
				assert.NilError(t, err)

				// query the delegator withdraw address
				addr, _ := f.distrKeeper.GetDelegatorWithdrawAddr(f.sdkCtx, delAddr)
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
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	// send funds to val addr
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(f.valAddr), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)
	coins := sdk.NewCoins(sdk.NewCoin("mytoken", math.NewInt(2)), sdk.NewCoin("stake", math.NewInt(2)))
	err = f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, coins)
	require.NoError(t, err)

	// check initial balance
	balance := f.bankKeeper.GetAllBalances(f.sdkCtx, sdk.AccAddress(f.valAddr))
	expTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1000)
	expCoins := sdk.NewCoins(sdk.NewCoin("stake", expTokens))
	assert.DeepEqual(t, expCoins, balance)

	// set outstanding rewards
	err = f.distrKeeper.SetValidatorOutstandingRewards(f.sdkCtx, f.valAddr, distrtypes.ValidatorOutstandingRewards{Rewards: valCommission})
	require.NoError(t, err)

	// set commission
	err = f.distrKeeper.SetValidatorAccumulatedCommission(f.sdkCtx, f.valAddr, distrtypes.ValidatorAccumulatedCommission{Commission: valCommission})
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
					sdk.NewCoin("mytoken", math.NewInt(1)),
					sdk.NewCoin("stake", expTokens.AddRaw(1)),
				), balance)

				// check remainder
				remainder, _ := f.distrKeeper.GetValidatorAccumulatedCommission(f.sdkCtx, f.valAddr)
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
	f := initFixture(t)

	// reset fee pool
	initPool := distrtypes.InitialFeePool()
	require.NoError(t, f.distrKeeper.FeePool.Set(f.sdkCtx, initPool))

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)

	addr := sdk.AccAddress(PKS[0].Address())
	addr2 := sdk.AccAddress(PKS[1].Address())
	amount := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))

	// fund the account by minting and sending amount from distribution module to addr
	err = f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amount)
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
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
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
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: addr2.String(),
			},
			expErr:    true,
			expErrMsg: "insufficient funds",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgFundCommunityPool{
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
				Depositor: addr.String(),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
				feePool, err := f.distrKeeper.FeePool.Get(f.sdkCtx)
				require.NoError(t, err)
				assert.DeepEqual(t, initPool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...), feePool.CommunityPool)
				assert.Assert(t, f.bankKeeper.GetAllBalances(f.sdkCtx, addr).Empty())
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// default params
	communityTax := math.LegacyNewDecWithPrec(2, 2) // 2%

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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
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
					WithdrawAddrEnabled: true,
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
				params, _ := f.distrKeeper.Params.Get(f.sdkCtx)
				assert.DeepEqual(t, distrtypes.DefaultParams(), params)
			}
		})
	}
}

func TestMsgCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	require.NoError(t, f.distrKeeper.Params.Set(f.sdkCtx, distrtypes.DefaultParams()))
	initialFeePool := sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(10000)})
	require.NoError(t, f.distrKeeper.FeePool.Set(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: initialFeePool,
	}))

	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(100))
	err := f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))
	require.NoError(t, err)

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
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "invalid recipient",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: f.distrKeeper.GetAuthority(),
				Recipient: recipient.String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			expErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
				communityPool, err := f.distrKeeper.FeePool.Get(f.sdkCtx)
				require.NoError(t, err)
				newPool, negative := initialFeePool.SafeSub(sdk.NewDecCoinsFromCoins(tc.msg.Amount...))
				assert.Assert(t, negative == false)
				assert.DeepEqual(t, communityPool.CommunityPool, newPool)
			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	require.NoError(t, f.distrKeeper.Params.Set(f.sdkCtx, distrtypes.DefaultParams()))
	err := f.distrKeeper.FeePool.Set(f.sdkCtx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	require.NoError(t, err)
	initTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(10000))
	require.NoError(t, f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))))

	// Set default staking params
	require.NoError(t, f.stakingKeeper.SetParams(f.sdkCtx, stakingtypes.DefaultParams()))

	addr := sdk.AccAddress("addr")
	addr1 := sdk.AccAddress(PKS[0].Address())
	valAddr1 := sdk.ValAddress(addr1)

	// send funds to val addr
	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64(1000))
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, sdk.AccAddress(valAddr1), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))
	require.NoError(t, err)
	// send funds from module to addr to perform DepositValidatorRewardsPool
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))
	require.NoError(t, err)
	tstaking := stakingtestutil.NewHelper(t, f.sdkCtx, f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr1, valConsPk0, math.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	require.NoError(t, f.bankKeeper.MintCoins(f.sdkCtx, distrtypes.ModuleName, amt))
	require.NoError(t, f.bankKeeper.SendCoinsFromModuleToAccount(f.sdkCtx, distrtypes.ModuleName, addr, amt))

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := f.app.RunMsg(
				tc.msg,
				integration.WithAutomaticFinalizeBlock(),
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
				outstandingRewards, _ := f.distrKeeper.GetValidatorOutstandingRewards(f.sdkCtx, val)
				for _, c := range tc.msg.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					assert.DeepEqual(t, x, math.LegacyNewDecFromInt(c.Amount))
				}

			}
		})
	}
}
