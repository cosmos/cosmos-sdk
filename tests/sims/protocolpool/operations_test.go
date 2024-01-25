package protocolpool

import (
	"math/rand"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	authkeeper "cosmossdk.io/x/auth/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/protocolpool/keeper"
	"cosmossdk.io/x/protocolpool/simulation"
	"cosmossdk.io/x/protocolpool/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type suite struct {
	Ctx sdk.Context
	App *runtime.App

	TxConfig      client.TxConfig
	Cdc           codec.Codec
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	PoolKeeper    keeper.Keeper
}

func setUpTest(t *testing.T) suite {
	t.Helper()
	res := suite{}

	var (
		appBuilder *runtime.AppBuilder
		err        error
	)

	app, err := simtestutil.Setup(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&res.AccountKeeper,
		&res.BankKeeper,
		&res.Cdc,
		&appBuilder,
		&res.StakingKeeper,
		&res.PoolKeeper,
		&res.TxConfig,
	)
	require.NoError(t, err)

	res.App = app
	res.Ctx = app.BaseApp.NewContext(false)
	return res
}

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	suite := setUpTest(t)

	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, suite.Cdc, suite.TxConfig, suite.AccountKeeper,
		suite.BankKeeper, suite.PoolKeeper)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, suite.Ctx, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgFundCommunityPool, types.ModuleName, sdk.MsgTypeURL(&types.MsgFundCommunityPool{})},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.App.BaseApp, suite.Ctx, accs, "")
		require.NoError(t, err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgFundCommunityPool tests the normal scenario of a valid message of type TypeMsgFundCommunityPool.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgFundCommunityPool(t *testing.T) {
	suite := setUpTest(t)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, suite.AccountKeeper, suite.BankKeeper, suite.StakingKeeper, suite.Ctx, 3)

	// execute operation
	op := simulation.SimulateMsgFundCommunityPool(suite.TxConfig, suite.AccountKeeper, suite.BankKeeper, suite.PoolKeeper)
	operationMsg, futureOperations, err := op(r, suite.App.BaseApp, suite.Ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgFundCommunityPool
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(t, err)
	require.True(t, operationMsg.OK)
	require.Equal(t, "4896096stake", msg.Amount.String())
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor)
	require.Equal(t, sdk.MsgTypeURL(&types.MsgFundCommunityPool{}), sdk.MsgTypeURL(&msg))
	require.Len(t, futureOperations, 0)
}

func getTestingAccounts(
	t *testing.T, r *rand.Rand,
	accountKeeper authkeeper.AccountKeeper, bankKeeper bankkeeper.Keeper,
	stakingKeeper *stakingkeeper.Keeper, ctx sdk.Context, n int,
) []simtypes.Account {
	t.Helper()
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := stakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := accountKeeper.NewAccountWithAddress(ctx, account.Address)
		accountKeeper.SetAccount(ctx, acc)
		require.NoError(t, banktestutil.FundAccount(ctx, bankKeeper, account.Address, initCoins))
	}

	return accounts
}
