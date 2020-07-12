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
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {

	app, ctx := createTestApp(false)

	ctx.WithChainID("test-chain")

	cdc := app.Codec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper,
		app.BankKeeper, app.StakingKeeper,
	)

	s := rand.NewSource(1)
	r := rand.New(s)
	accs := simtypes.RandomAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{{simappparams.DefaultWeightMsgCreateValidator, types.ModuleName, types.TypeMsgCreateValidator},
		{simappparams.DefaultWeightMsgEditValidator, types.ModuleName, types.TypeMsgEditValidator},
		{simappparams.DefaultWeightMsgDelegate, types.ModuleName, types.TypeMsgDelegate},
		{simappparams.DefaultWeightMsgUndelegate, types.ModuleName, types.TypeMsgUndelegate},
		{simappparams.DefaultWeightMsgBeginRedelegate, types.ModuleName, types.TypeMsgBeginRedelegate},
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

// TestSimulateMsgCreateValidator tests the normal scenario of a valid message of type TypeMsgCreateValidator.
// Abonormal scenarios, where the message are created by an errors are not tested here.
func TestSimulateMsgCreateValidator(t *testing.T) {
	app, ctx := createTestApp(false)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgCreateValidator(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgCreateValidator
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "0.170063593193511020", msg.Commission.MaxChangeRate.String())
	require.Equal(t, "0.660000000000000000", msg.Commission.MaxRate.String())
	require.Equal(t, "0.047464127245687382", msg.Commission.Rate.String())
	require.Equal(t, types.TypeMsgCreateValidator, msg.Type())
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress.String())
	require.Equal(t, "cosmosvalconspub1addwnpepqwr8k5g44urevkvz5ys2qjag0nnp6xkd2f8lejn5pw2rehkjt6ftv5d9nrp", msg.Pubkey)
	require.Equal(t, "cosmosvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnsvnaes", msg.ValidatorAddress.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgEditValidator tests the normal scenario of a valid message of type TypeMsgEditValidator.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgEditValidator(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator
	_ = getTestingValidator0(t, app, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgEditValidator(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgEditValidator
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "0.280623462081924936", msg.CommissionRate.String())
	require.Equal(t, "jLxzIivHSl", msg.GetDescription().Moniker)
	require.Equal(t, "rBqDOTtGTO", msg.GetDescription().Identity)
	require.Equal(t, "BSpYuLyYgg", msg.GetDescription().Website)
	require.Equal(t, "wNbeHVIkPZ", msg.GetDescription().SecurityContact)
	require.Equal(t, types.TypeMsgEditValidator, msg.Type())
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgDelegate tests the normal scenario of a valid message of type TypeMsgDelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgDelegate(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)
	setupValidatorRewards(app, ctx, validator0.OperatorAddress)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgDelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgDelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress.String())
	require.Equal(t, "4896096", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgDelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress.String())
	require.Len(t, futureOperations, 0)
}

// TestSimulateMsgUndelegate tests the normal scenario of a valid message of type TypeMsgUndelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func TestSimulateMsgUndelegate(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// setup delegation
	delTokens := sdk.TokensFromConsensusPower(2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := types.NewDelegation(delegator.Address, validator0.OperatorAddress, issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator0.OperatorAddress, delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.OperatorAddress)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgUndelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgUndelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.DelegatorAddress.String())
	require.Equal(t, "560969", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgUndelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z", msg.ValidatorAddress.String())
	require.Len(t, futureOperations, 0)

}

// TestSimulateMsgBeginRedelegate tests the normal scenario of a valid message of type TypeMsgBeginRedelegate.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgBeginRedelegate(t *testing.T) {
	app, ctx := createTestApp(false)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup 3 accounts
	s := rand.NewSource(5)
	r := rand.New(s)
	accounts := getTestingAccounts(t, r, app, ctx, 3)

	// setup accounts[0] as validator0 and accounts[1] as validator1
	validator0 := getTestingValidator0(t, app, ctx, accounts)
	validator1 := getTestingValidator1(t, app, ctx, accounts)

	delTokens := sdk.TokensFromConsensusPower(2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)

	// setup accounts[2] as delegator
	delegator := accounts[2]
	delegation := types.NewDelegation(delegator.Address, validator1.OperatorAddress, issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator1.OperatorAddress, delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(app, ctx, validator0.OperatorAddress)
	setupValidatorRewards(app, ctx, validator1.OperatorAddress)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgBeginRedelegate(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgBeginRedelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, "cosmos12gwd9jchc69wck8dhstxgwz3z8qs8yv67ps8mu", msg.DelegatorAddress.String())
	require.Equal(t, "692322", msg.Amount.Amount.String())
	require.Equal(t, "stake", msg.Amount.Denom)
	require.Equal(t, types.TypeMsgBeginRedelegate, msg.Type())
	require.Equal(t, "cosmosvaloper1h6a7shta7jyc72hyznkys683z98z36e0zdk8g9", msg.GetValidatorDstAddress().String())
	require.Equal(t, "cosmosvaloper17s94pzwhsn4ah25tec27w70n65h5t2scgxzkv2", msg.GetValidatorSrcAddress().String())
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

func getTestingValidator0(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account) types.Validator {
	commission0 := types.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return getTestingValidator(t, app, ctx, accounts, commission0, 0)
}

func getTestingValidator1(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account) types.Validator {
	commission1 := types.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	return getTestingValidator(t, app, ctx, accounts, commission1, 1)
}

func getTestingValidator(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account, commission types.Commission, n int) types.Validator {
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := types.NewValidator(valAddr, valPubKey, types.Description{})
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
