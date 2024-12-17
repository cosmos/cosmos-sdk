package protocolpool

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	_ "cosmossdk.io/x/bank"     // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	_ "cosmossdk.io/x/consensus"    // import as blank for app wiring
	_ "cosmossdk.io/x/distribution" // import as blank for app wiring
	_ "cosmossdk.io/x/mint"         // import as blank for app wiring
	"cosmossdk.io/x/mint/types"
	_ "cosmossdk.io/x/protocolpool" // import as blank for app wiring
	protocolpoolkeeper "cosmossdk.io/x/protocolpool/keeper"
	protocolpooltypes "cosmossdk.io/x/protocolpool/types"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/genutil" // import as blank for app wiring
)

var moduleConfigs = []configurator.ModuleOption{
	configurator.AccountsModule(),
	configurator.AuthModule(),
	configurator.BankModule(),
	configurator.StakingModule(),
	configurator.TxModule(),
	configurator.ValidateModule(),
	configurator.ConsensusModule(),
	configurator.GenutilModule(),
	configurator.MintModule(),
	configurator.DistributionModule(),
	configurator.ProtocolPoolModule(),
}

type fixture struct {
	accountKeeper      authkeeper.AccountKeeper
	protocolpoolKeeper protocolpoolkeeper.Keeper
	bankKeeper         bankkeeper.Keeper
	stakingKeeper      *stakingkeeper.Keeper
}

// TestWithdrawAnytime tests if withdrawing funds many times vs withdrawing funds once
// yield the same end balance.
func TestWithdrawAnytime(t *testing.T) {
	res := fixture{}

	startupCfg := integration.DefaultStartUpConfig(t)
	startupCfg.HeaderService = &integration.HeaderService{}

	app, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg, &res.accountKeeper, &res.protocolpoolKeeper, &res.bankKeeper, &res.stakingKeeper)
	require.NoError(t, err)

	ctx := app.StateLatestContext(t)
	acc := res.accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)

	testAddrs := simtestutil.AddTestAddrs(res.bankKeeper, res.stakingKeeper, ctx, 5, math.NewInt(1))
	testAddr0Str, err := res.accountKeeper.AddressCodec().BytesToString(testAddrs[0])
	require.NoError(t, err)

	msgServer := protocolpoolkeeper.NewMsgServerImpl(res.protocolpoolKeeper)
	_, err = msgServer.CreateContinuousFund(
		ctx,
		&protocolpooltypes.MsgCreateContinuousFund{
			Authority:  res.protocolpoolKeeper.GetAuthority(),
			Recipient:  testAddr0Str,
			Percentage: math.LegacyMustNewDecFromStr("0.5"),
		},
	)
	require.NoError(t, err)

	// increase the community pool by a bunch
	for i := 0; i < 30; i++ {
		_, state := app.Deliver(t, ctx, nil)
		_, err = app.Commit(state)
		require.NoError(t, err)

		headerInfo := integration.HeaderInfoFromContext(ctx)
		headerInfo.Time = headerInfo.Time.Add(time.Minute)
		ctx = integration.SetHeaderInfo(ctx, headerInfo)

		// withdraw funds randomly, but it must always land on the same end balance
		if rand.Intn(100) > 50 {
			_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
				RecipientAddress: testAddr0Str,
			})
			require.NoError(t, err)
		}
	}

	pool, err := res.protocolpoolKeeper.GetCommunityPool(ctx)
	require.NoError(t, err)
	require.True(t, pool.IsAllGT(sdk.NewCoins(sdk.NewInt64Coin("stake", 100000))))

	_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
		RecipientAddress: testAddr0Str,
	})
	require.NoError(t, err)

	endBalance := res.bankKeeper.GetBalance(ctx, testAddrs[0], sdk.DefaultBondDenom)
	require.Equal(t, "11883031stake", endBalance.String())
}

// TestExpireInTheMiddle tests if a continuous fund that expires without anyone
// calling the withdraw function, the funds are still distributed correctly.
func TestExpireInTheMiddle(t *testing.T) {
	res := fixture{}

	startupCfg := integration.DefaultStartUpConfig(t)
	startupCfg.HeaderService = &integration.HeaderService{}

	app, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg, &res.accountKeeper, &res.protocolpoolKeeper, &res.bankKeeper, &res.stakingKeeper)
	require.NoError(t, err)

	ctx := app.StateLatestContext(t)

	acc := res.accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)

	testAddrs := simtestutil.AddTestAddrs(res.bankKeeper, res.stakingKeeper, ctx, 5, math.NewInt(1))
	testAddr0Str, err := res.accountKeeper.AddressCodec().BytesToString(testAddrs[0])
	require.NoError(t, err)

	msgServer := protocolpoolkeeper.NewMsgServerImpl(res.protocolpoolKeeper)

	headerInfo := integration.HeaderInfoFromContext(ctx)
	expirationTime := headerInfo.Time.Add(time.Minute * 2)
	_, err = msgServer.CreateContinuousFund(
		ctx,
		&protocolpooltypes.MsgCreateContinuousFund{
			Authority:  res.protocolpoolKeeper.GetAuthority(),
			Recipient:  testAddr0Str,
			Percentage: math.LegacyMustNewDecFromStr("0.1"),
			Expiry:     &expirationTime,
		},
	)
	require.NoError(t, err)

	// increase the community pool by a bunch
	for i := 0; i < 30; i++ {
		_, state := app.Deliver(t, ctx, nil)
		_, err = app.Commit(state)
		require.NoError(t, err)

		headerInfo := integration.HeaderInfoFromContext(ctx)
		headerInfo.Time = headerInfo.Time.Add(time.Minute)
		ctx = integration.SetHeaderInfo(ctx, headerInfo)
		require.NoError(t, err)
	}

	_, err = msgServer.WithdrawContinuousFund(ctx, &protocolpooltypes.MsgWithdrawContinuousFund{
		RecipientAddress: testAddr0Str,
	})
	require.NoError(t, err)

	endBalance := res.bankKeeper.GetBalance(ctx, testAddrs[0], sdk.DefaultBondDenom)
	require.Equal(t, "237661stake", endBalance.String())
}
