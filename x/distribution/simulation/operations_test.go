package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	app, ctx := createTestApp(false)
	ctx.WithChainID("test-chain")

	cdc := app.Codec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.DistrKeeper, app.StakingKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := getTestingAccounts(t, r, app, ctx, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simappparams.DefaultWeightMsgSetWithdrawAddress, types.ModuleName, types.TypeMsgSetWithdrawAddress},
		{simappparams.DefaultWeightMsgWithdrawDelegationReward, types.ModuleName, types.TypeMsgWithdrawDelegatorReward},
		{simappparams.DefaultWeightMsgWithdrawValidatorCommission, types.ModuleName, types.TypeMsgWithdrawValidatorCommission},
		{simappparams.DefaultWeightMsgFundCommunityPool, types.ModuleName, types.TypeMsgFundCommunityPool},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgSetWithdrawAddress tests the normal scenario of a valid message of type TypeMsgSetWithdrawAddress.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgSetWithdrawAddress(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgSetWithdrawAddress(app.AccountKeeper, app.BankKeeper, app.DistrKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgSetWithdrawAddress
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.GetDelegatorAddress().String())
	require.Equal(t, "cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.GetWithdrawAddress().String())
	require.Equal(t, types.TypeMsgSetWithdrawAddress, msg.Type())
	require.Equal(t, types.ModuleName, msg.Route())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgWithdrawDelegatorReward tests the normal scenario of a valid message
// of type TypeMsgWithdrawDelegatorReward.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgWithdrawDelegatorReward(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(4)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := stakingtypes.NewDelegation(delegator.Address, validator0.OperatorAddress, issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator0.OperatorAddress, delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.OperatorAddress)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgWithdrawDelegatorReward(app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgWithdrawDelegatorReward
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmosvaloper1l4s054098kk9hmr5753c6k3m2kw65h686d3mhr", msg.GetValidatorAddress().String())
	require.Equal(t, "cosmos1d6u7zhjwmsucs678d7qn95uqajd4ucl9jcjt26", msg.GetDelegatorAddress().String())
	require.Equal(t, types.TypeMsgWithdrawDelegatorReward, msg.Type())
	require.Equal(t, types.ModuleName, msg.Route())
	require.Len(t, futureOperations, 0)

}

// TestSimulateMsgWithdrawValidatorCommission tests the normal scenario of a valid message
// of type TypeMsgWithdrawValidatorCommission.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgWithdrawValidatorCommission(t *testing.T) {
	testSimulateMsgWithdrawValidatorCommission(t, "atoken")
	// TODO: Fix following test: It is failing because of issue #6763
	// testSimulateMsgWithdrawValidatorCommission(t, "tokenxxx")
}

func testSimulateMsgWithdrawValidatorCommission(t *testing.T, tokenName string) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// set module account coins
	distrAcc := app.DistrKeeper.GetDistributionAccount(ctx)
	err := app.BankKeeper.SetBalances(ctx, distrAcc.GetAddress(), sdk.NewCoins(
		sdk.NewCoin(tokenName, sdk.NewInt(10)),
		sdk.NewCoin("stake", sdk.NewInt(5)),
	))
	require.NoError(t, err)
	app.AccountKeeper.SetModuleAccount(ctx, distrAcc)

	// set outstanding rewards
	valCommission := sdk.DecCoins{
		sdk.NewDecCoinFromDec(tokenName, sdk.NewDec(5).Quo(sdk.NewDec(2))),
		sdk.NewDecCoinFromDec("stake", sdk.NewDec(1).Quo(sdk.NewDec(1))),
	}

	app.DistrKeeper.SetValidatorOutstandingRewards(ctx, validator0.OperatorAddress, types.ValidatorOutstandingRewards{Rewards: valCommission})

	// setup validator accumulated commission
	app.DistrKeeper.SetValidatorAccumulatedCommission(ctx, validator0.OperatorAddress, types.ValidatorAccumulatedCommission{Commission: valCommission})

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgWithdrawValidatorCommission(app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgWithdrawValidatorCommission
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.GetValidatorAddress().String())
	require.Equal(t, types.TypeMsgWithdrawValidatorCommission, msg.Type())
	require.Equal(t, types.ModuleName, msg.Route())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgFundCommunityPool tests the normal scenario of a valid message of type TypeMsgFundCommunityPool.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgFundCommunityPool(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgFundCommunityPool(app.AccountKeeper, app.BankKeeper, app.DistrKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgFundCommunityPool
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "4896096stake", msg.GetAmount().String())
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Depositor.String())
	require.Equal(t, types.TypeMsgFundCommunityPool, msg.Type())
	require.Equal(t, types.ModuleName, msg.Route())
	require.Len(t, futureOperations, 0)
}

// returns context and an app with updated mint keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)

	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.MintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())

	return app, ctx
}

func getTestingAccounts(t *testing.T, r *rand.Rand, app *simapp.SimApp, ctx sdk.Context, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		err := app.BankKeeper.SetBalances(ctx, account.Address, initCoins)
		require.NoError(t, err)
	}

	return accounts
}

func getTestingValidator0(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return getTestingValidator(t, app, ctx, accounts, commission0, 0)
}

func getTestingValidator(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := stakingtypes.NewValidator(valAddr, valPubKey, stakingtypes.Description{})
	validator, err := validator.SetInitialCommission(commission)
	require.NoError(t, err)

	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = sdk.NewInt(1000000)

	app.StakingKeeper.SetValidator(ctx, validator)

	return validator
}

func setupValidatorRewards(app *simapp.SimApp, ctx sdk.Context, valAddress sdk.ValAddress) {
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	app.DistrKeeper.SetValidatorHistoricalRewards(ctx, valAddress, 2, historicalRewards)
	// setup current revards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	app.DistrKeeper.SetValidatorCurrentRewards(ctx, valAddress, currentRewards)

}
