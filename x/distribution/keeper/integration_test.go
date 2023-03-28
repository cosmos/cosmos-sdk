package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/codec"
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
		stakingtypes.BondedPoolName:    nil,
		stakingtypes.NotBondedPoolName: nil,
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		keys[authtypes.StoreKey],
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
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper, distrkeeper.NewQuerier(f.distrKeeper))

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

	f.distrKeeper.SetParams(integrationApp.Ctx, distrtypes.DefaultParams())
	f.distrKeeper.SetFeePool(integrationApp.Ctx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	initTokens := f.stakingKeeper.TokensFromConsensusPower(integrationApp.Ctx, int64(100))
	f.bankKeeper.MintCoins(integrationApp.Ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(f.distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper, distrkeeper.NewQuerier(f.distrKeeper))

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
				result := distrtypes.MsgCommunityPoolSpend{}
				err = f.cdc.Unmarshal(res[0].Value, &result)
				assert.NilError(t, err)
			}
		})
	}
}
