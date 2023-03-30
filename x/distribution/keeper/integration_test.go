package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
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

type fixture struct {
	cdc  codec.Codec
	keys map[string]*storetypes.KVStoreKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	distrKeeper   distrkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
}

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, stakingtypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, distribution.AppModuleBasic{}).Codec

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
		cdc, keys[distrtypes.StoreKey], accountKeeper, bankKeeper, stakingKeeper, authtypes.FeeCollectorName, authority.String(),
	)

	f.cdc = cdc
	f.keys = keys
	f.accountKeeper = accountKeeper
	f.bankKeeper = bankKeeper
	f.stakingKeeper = stakingKeeper
	f.distrKeeper = distrKeeper

	return f
}

func TestMsgUpdateParams(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// default params
	communityTax := sdk.NewDecWithPrec(2, 2) // 2%
	withdrawAddrEnabled := true

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.SetupTestApp(t, f.keys, authModule, bankModule, stakingModule, distrModule)

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper(), distrkeeper.NewQuerier(f.distrKeeper))

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
			expErrMsg: "base and bonus proposer reward are deprecated fields and should not be used",
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
			expErrMsg: "base and bonus proposer reward are deprecated fields and should not be used",
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
			res, err := integrationApp.ExecMsgs(tc.msg)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgUpdateParams{}
				err = f.cdc.Unmarshal(res[0].Value, &result)
				assert.NilError(t, err)
			}
		})
	}
}

func TestCommunityPoolSpend(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.SetupTestApp(t, f.keys, authModule, bankModule, stakingModule, distrModule)

	f.distrKeeper.SetParams(integrationApp.SDKContext(), distrtypes.DefaultParams())
	f.distrKeeper.SetFeePool(integrationApp.SDKContext(), distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	initTokens := f.stakingKeeper.TokensFromConsensusPower(integrationApp.SDKContext(), int64(100))
	f.bankKeeper.MintCoins(integrationApp.SDKContext(), distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper(), distrkeeper.NewQuerier(f.distrKeeper))

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
			res, err := integrationApp.ExecMsgs(tc.msg)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				// check the result
				result := distrtypes.MsgCommunityPoolSpend{}
				err = f.cdc.Unmarshal(res[0].Value, &result)
				assert.NilError(t, err)
			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	authModule := auth.NewAppModule(f.cdc, f.accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(f.cdc, f.bankKeeper, f.accountKeeper, nil)
	stakingModule := staking.NewAppModule(f.cdc, f.stakingKeeper, f.accountKeeper, f.bankKeeper, nil)
	distrModule := distribution.NewAppModule(f.cdc, f.distrKeeper, f.accountKeeper, f.bankKeeper, f.stakingKeeper, nil)

	integrationApp := integration.SetupTestApp(t, f.keys, authModule, bankModule, stakingModule, distrModule)

	f.distrKeeper.SetParams(integrationApp.SDKContext(), distrtypes.DefaultParams())
	f.distrKeeper.SetFeePool(integrationApp.SDKContext(), distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	initTokens := f.stakingKeeper.TokensFromConsensusPower(integrationApp.SDKContext(), int64(10000))
	f.bankKeeper.MintCoins(integrationApp.SDKContext(), distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// Set default staking params
	f.stakingKeeper.SetParams(integrationApp.SDKContext(), stakingtypes.DefaultParams())

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper(), distrkeeper.NewQuerier(f.distrKeeper))

	addr := sdk.AccAddress([]byte("addr"))
	addr1 := sdk.AccAddress(PKS[0].Address())
	valAddr1 := sdk.ValAddress(addr1)

	// send funds to val addr
	tokens := f.stakingKeeper.TokensFromConsensusPower(integrationApp.SDKContext(), int64(1000))
	f.bankKeeper.SendCoinsFromModuleToAccount(integrationApp.SDKContext(), distrtypes.ModuleName, sdk.AccAddress(valAddr1), sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))

	// send funds from module to addr to perform DepositValidatorRewardsPool
	f.bankKeeper.SendCoinsFromModuleToAccount(integrationApp.SDKContext(), distrtypes.ModuleName, addr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tokens)))

	tstaking := stakingtestutil.NewHelper(t, integrationApp.SDKContext(), f.stakingKeeper)
	tstaking.Commission = stakingtypes.NewCommissionRates(sdk.NewDecWithPrec(5, 1), sdk.NewDecWithPrec(5, 1), math.LegacyNewDec(0))
	tstaking.CreateValidator(valAddr1, valConsPk0, sdk.NewInt(100), true)

	// mint a non-staking token and send to an account
	amt := sdk.NewCoins(sdk.NewInt64Coin("foo", 500))
	f.bankKeeper.MintCoins(integrationApp.SDKContext(), distrtypes.ModuleName, amt)
	f.bankKeeper.SendCoinsFromModuleToAccount(integrationApp.SDKContext(), distrtypes.ModuleName, addr, amt)

	testCases := []struct {
		name      string
		msg       *distrtypes.MsgDepositValidatorRewardsPool
		expErr    bool
		expErrMsg string
	}{
		{
			name: "happy path (staking token)",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Authority:        addr.String(),
				ValidatorAddress: valAddr1.String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(integrationApp.SDKContext()), sdk.NewInt(100))),
			},
		},
		{
			name: "happy path (non-staking token)",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Authority:        addr.String(),
				ValidatorAddress: valAddr1.String(),
				Amount:           amt,
			},
		},
		{
			name: "invalid validator",
			msg: &distrtypes.MsgDepositValidatorRewardsPool{
				Authority:        addr.String(),
				ValidatorAddress: sdk.ValAddress([]byte("addr1_______________")).String(),
				Amount:           sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(integrationApp.SDKContext()), sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "validator does not exist",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := integrationApp.ExecMsgs(tc.msg)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				assert.Assert(t, res != nil)

				valAddr, err := sdk.ValAddressFromBech32(tc.msg.ValidatorAddress)
				assert.NilError(t, err)

				// check validator outstanding rewards
				outstandingRewards := f.distrKeeper.GetValidatorOutstandingRewards(integrationApp.SDKContext(), valAddr)
				for _, c := range tc.msg.Amount {
					x := outstandingRewards.Rewards.AmountOf(c.Denom)
					assert.DeepEqual(t, x, sdk.NewDecFromInt(c.Amount))
				}

				// check the result
				result := distrtypes.MsgDepositValidatorRewardsPoolResponse{}
				err = f.cdc.Unmarshal(res[0].Value, &result)
				assert.NilError(t, err)
			}
		})
	}
}
