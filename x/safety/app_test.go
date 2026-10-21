package safety_test

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	modulev1 "cosmossdk.io/api/cosmos/safety/module/v1"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log/v2"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/safety"
	"github.com/cosmos/cosmos-sdk/x/safety/keeper"
	safetytestutil "github.com/cosmos/cosmos-sdk/x/safety/testutil"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

// TestDepinjectWiring builds an app through depinject and checks that the
// keeper is installed as the BaseApp circuit breaker: a tripped msg fails in
// the router, and a reset brings it back.
func TestDepinjectWiring(t *testing.T) {
	// the consensus authority override would replace params.authority; clear it
	// so that the x/safety authority set in genesis is the one that counts
	prevAuthority := sims.DefaultConsensusParams.Authority
	sims.DefaultConsensusParams.Authority = nil
	t.Cleanup(func() { sims.DefaultConsensusParams.Authority = prevAuthority })

	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	other := sdk.AccAddress("other_______________")

	coins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1_000_000))
	startupCfg := sims.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = []sims.GenesisAccount{{
		GenesisAccount: &authtypes.BaseAccount{Address: addr.String()},
		Coins:          coins,
	}}

	var (
		safetyKeeper keeper.Keeper
		bankKeeper   bankkeeper.Keeper
	)
	app, err := sims.SetupWithConfiguration(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AuthModule(),
				configurator.BankModule(),
				configurator.StakingModule(),
				configurator.TxModule(),
				configurator.ConsensusModule(),
				safetytestutil.SafetyModule(addr.String()),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg, &safetyKeeper, &bankKeeper,
	)
	require.NoError(t, err)

	ctx := app.NewContext(true)
	params, err := safetyKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(addr.String()), params)

	txConfig := moduletestutil.MakeTestTxConfig()
	seq := uint64(0)
	deliver := func(expPass bool, msgs ...sdk.Msg) error {
		t.Helper()
		header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
		_, _, err := sims.SignCheckDeliver(t, txConfig, app.BaseApp, header, msgs, "", []uint64{0}, []uint64{seq}, expPass, expPass, priv)
		// this app has no SafetyDecorator, so a blocked tx still passes the
		// ante handler and bumps the sequence before the router rejects it
		seq++
		return err
	}
	send := banktypes.NewMsgSend(addr, other, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))

	// make addr the operator as well, through the authority path
	newParams := types.Params{Authority: addr.String(), Operator: addr.String(), Scope: []string{"/cosmos.bank.*"}}
	require.NoError(t, deliver(true, &types.MsgUpdateParams{Authority: addr.String(), Params: newParams}))
	require.NoError(t, deliver(true, send))

	require.NoError(t, deliver(true, types.NewMsgTrip(addr.String(), "/cosmos.bank.v1beta1.MsgSend")))
	err = deliver(false, send)
	require.ErrorContains(t, err, "/cosmos.bank.v1beta1.MsgSend")

	require.NoError(t, deliver(true, types.NewMsgReset(addr.String(), "/cosmos.bank.v1beta1.MsgSend")))
	require.NoError(t, deliver(true, send))

	ctx = app.NewContext(true)
	require.Equal(t, int64(2), bankKeeper.GetBalance(ctx, other, sdk.DefaultBondDenom).Amount.Int64())
}

// ProvideModule reads Config.Authority as a bech32 address when it decodes,
// and as a module name otherwise; empty means the gov module.
func TestProvideModuleAuthority(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	ac := addresscodec.NewBech32Codec("cosmos")
	direct := sdk.AccAddress("direct_authority____").String()

	for _, tc := range []struct {
		config, want string
	}{
		{"", authtypes.NewModuleAddress("gov").String()},
		{direct, direct},
		{"some_module", authtypes.NewModuleAddress("some_module").String()},
	} {
		t.Run(tc.config, func(t *testing.T) {
			out := safety.ProvideModule(safety.ModuleInputs{
				Config:       &modulev1.Module{Authority: tc.config},
				Cdc:          encCfg.Codec,
				StoreService: runtime.NewKVStoreService(storetypes.NewKVStoreKey(types.StoreKey)),
				AddressCodec: ac,
			})
			require.Equal(t, tc.want, out.SafetyKeeper.DefaultAuthority())
		})
	}
}
