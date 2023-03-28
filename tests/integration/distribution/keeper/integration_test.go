package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"gotest.tools/v3/assert"

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

func TestIntegrationCommunityPoolSpend(t *testing.T) {
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
	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)

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
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)

	stakingKeeper := stakingkeeper.NewKeeper(cdc, keys[stakingtypes.StoreKey], accountKeeper, bankKeeper, authority.String())
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)

	distrKeeper := distrkeeper.NewKeeper(
		cdc, keys[distrtypes.StoreKey], accountKeeper, bankKeeper, stakingKeeper, authtypes.FeeCollectorName, authority.String(),
	)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, nil)

	integrationApp := integration.SetupTestApp(t, keys, authModule, bankModule, stakingModule, distrModule)

	distrKeeper.SetParams(integrationApp.Ctx, distrtypes.DefaultParams())
	distrKeeper.SetFeePool(integrationApp.Ctx, distrtypes.FeePool{
		CommunityPool: sdk.NewDecCoins(sdk.DecCoin{Denom: "stake", Amount: math.LegacyNewDec(100)}),
	})
	initTokens := stakingKeeper.TokensFromConsensusPower(integrationApp.Ctx, int64(100))
	bankKeeper.MintCoins(integrationApp.Ctx, distrtypes.ModuleName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens)))

	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryServiceHelper, distrkeeper.NewQuerier(distrKeeper))

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
				Authority: distrKeeper.GetAuthority(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(100))),
			},
			expErr:    true,
			expErrMsg: "decoding bech32 failed",
		},
		{
			name: "valid message",
			msg: &distrtypes.MsgCommunityPoolSpend{
				Authority: distrKeeper.GetAuthority(),
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
				err = cdc.Unmarshal(res[0].Value, &result)
				assert.NilError(t, err)
			}
		})
	}
}
