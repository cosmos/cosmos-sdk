package simulation_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SimTestSuite struct {
	suite.Suite

	r   *rand.Rand
	ctx sdk.Context

	app      *runtime.App
	codec    codec.Codec
	txConfig client.TxConfig

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *keeper.Keeper
	distrKeeper   distributionkeeper.Keeper
	mintKeeper    mintkeeper.Keeper

	accounts []simtypes.Account
}

func (suite *SimTestSuite) SetupTest() {
	sdk.DefaultPowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	s := rand.NewSource(12)
	suite.r = rand.New(s)
	accounts := simtypes.RandomAccounts(suite.r, 4)

	// create validator (non random as using a seed)
	createValidator := func() (*tmtypes.ValidatorSet, error) {
		account := accounts[0]
		tmPk, err := cryptocodec.ToTmPubKeyInterface(account.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create pubkey: %w", err)
		}

		validator := tmtypes.NewValidator(tmPk, 1)

		return tmtypes.NewValidatorSet([]*tmtypes.Validator{validator}), nil
	}

	app, err := simtestutil.SetupWithConfiguration(
		testutil.AppConfig,
		createValidator,
		nil,
		false,
		&suite.codec,
		&suite.txConfig,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.stakingKeeper,
		&suite.mintKeeper,
		&suite.distrKeeper,
	)
	suite.Require().NoError(err)

	suite.app = app
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{})

	// remove genesis validator account
	suite.accounts = accounts[1:]

	suite.mintKeeper.SetParams(suite.ctx, minttypes.DefaultParams())
	suite.mintKeeper.SetMinter(suite.ctx, minttypes.DefaultInitialMinter())

	initAmt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range suite.accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.bankKeeper, suite.ctx, account.Address, initCoins))
	}
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams,
		suite.codec,
		suite.accountKeeper,
		suite.bankKeeper,
		suite.stakingKeeper,
	)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simtestutil.DefaultWeightMsgCreateValidator, types.ModuleName, types.TypeMsgCreateValidator},
		{simtestutil.DefaultWeightMsgEditValidator, types.ModuleName, types.TypeMsgEditValidator},
		{simtestutil.DefaultWeightMsgDelegate, types.ModuleName, types.TypeMsgDelegate},
		{simtestutil.DefaultWeightMsgUndelegate, types.ModuleName, types.TypeMsgUndelegate},
		{simtestutil.DefaultWeightMsgBeginRedelegate, types.ModuleName, types.TypeMsgBeginRedelegate},
		{simtestutil.DefaultWeightMsgCancelUnbondingDelegation, types.ModuleName, types.TypeMsgCancelUnbondingDelegation},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(suite.r, suite.app.BaseApp, suite.ctx, suite.accounts, suite.ctx.ChainID())

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgCreateValidator tests the normal scenario of a valid message of type TypeMsgCreateValidator.
// Abonormal scenarios, where the message are created by an errors are not tested here.
func (suite *SimTestSuite) TestSimulateMsgCreateValidator() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash}})

	// execute operation
	op := simulation.SimulateMsgCreateValidator(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgCreateValidator
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("0.000000000000000000", msg.Commission.MaxChangeRate.String())
	suite.Require().Equal("0.100000000000000000", msg.Commission.MaxRate.String())
	suite.Require().Equal("0.001385997968089102", msg.Commission.Rate.String())
	suite.Require().Equal(types.TypeMsgCreateValidator, msg.Type())
	suite.Require().Equal([]byte{0xa, 0x20, 0x9a, 0xa4, 0xc5, 0x27, 0x6c, 0xa0, 0x25, 0x4, 0x0, 0xfa, 0x40, 0x94, 0xe4, 0xd0, 0x9a, 0x32, 0x27, 0x8e, 0xd8, 0xfc, 0x1a, 0xe1, 0x10, 0x44, 0x81, 0x2, 0x59, 0x1f, 0xf5, 0xcc, 0x4, 0x12}, msg.Pubkey.Value)
	suite.Require().Equal("cosmos1092v0qgulpejj8y8hs6dmlw82x9gv8f7jfc7jl", msg.DelegatorAddress)
	suite.Require().Equal("cosmosvaloper1092v0qgulpejj8y8hs6dmlw82x9gv8f7havt7v", msg.ValidatorAddress)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgCancelUnbondingDelegation tests the normal scenario of a valid message of type TypeMsgCancelUnbondingDelegation.
// Abonormal scenarios, where the message is
func (suite *SimTestSuite) TestSimulateMsgCancelUnbondingDelegation() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(suite.T(), suite.stakingKeeper, ctx, accounts)

	// setup delegation
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := types.NewDelegation(delegator.Address, validator0.GetOperator(), issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)
	suite.distrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(suite.distrKeeper, ctx, validator0.GetOperator())

	// unbonding delegation
	udb := types.NewUnbondingDelegation(delegator.Address, validator0.GetOperator(), app.LastBlockHeight(), blockTime.Add(2*time.Minute), delTokens)
	suite.stakingKeeper.SetUnbondingDelegation(ctx, udb)
	setupValidatorRewards(suite.distrKeeper, ctx, validator0.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgCancelUnbondingDelegate(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	accounts = []simtypes.Account{accounts[1]}
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgCancelUnbondingDelegation
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal(types.TypeMsgCancelUnbondingDelegation, msg.Type())
	suite.Require().Equal(delegator.Address.String(), msg.DelegatorAddress)
	suite.Require().Equal(validator0.GetOperator().String(), msg.ValidatorAddress)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgEditValidator tests the normal scenario of a valid message of type TypeMsgEditValidator.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (suite *SimTestSuite) TestSimulateMsgEditValidator() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup accounts[0] as validator
	_ = getTestingValidator0(suite.T(), suite.stakingKeeper, ctx, accounts)

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgEditValidator(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgEditValidator
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("0.054423227322998987", msg.CommissionRate.String())
	suite.Require().Equal("uKsnWClNgX", msg.Description.Moniker)
	suite.Require().Equal("lYcbnxUNtv", msg.Description.Identity)
	suite.Require().Equal("TEOJLpsMBI", msg.Description.Website)
	suite.Require().Equal("BsutAaKyBg", msg.Description.SecurityContact)
	suite.Require().Equal(types.TypeMsgEditValidator, msg.Type())
	suite.Require().Equal("cosmosvaloper1gnkw3uqzflagcqn6ekjwpjanlne928qhruemah", msg.ValidatorAddress)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgDelegate tests the normal scenario of a valid message of type TypeMsgDelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (suite *SimTestSuite) TestSimulateMsgDelegate() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// execute operation
	op := simulation.SimulateMsgDelegate(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgDelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmos1092v0qgulpejj8y8hs6dmlw82x9gv8f7jfc7jl", msg.DelegatorAddress)
	suite.Require().Equal("155698826349247340748", msg.Amount.Amount.String())
	suite.Require().Equal("stake", msg.Amount.Denom)
	suite.Require().Equal(types.TypeMsgDelegate, msg.Type())
	suite.Require().Equal("cosmosvaloper1sgheev7n8cgxwf40xpuphnmlasf8da9537sdvd", msg.ValidatorAddress)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgUndelegate tests the normal scenario of a valid message of type TypeMsgUndelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (suite *SimTestSuite) TestSimulateMsgUndelegate() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup accounts[0] as validator
	validator0 := getTestingValidator0(suite.T(), suite.stakingKeeper, ctx, accounts)

	// setup delegation
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := accounts[1]
	delegation := types.NewDelegation(delegator.Address, validator0.GetOperator(), issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)
	suite.distrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(suite.distrKeeper, ctx, validator0.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgUndelegate(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgUndelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmos1kk653svg7ksj9fmu85x9ygj4jzwlyrgsz38xle", msg.DelegatorAddress)
	suite.Require().Equal("1207344731929845964", msg.Amount.Amount.String())
	suite.Require().Equal("stake", msg.Amount.Denom)
	suite.Require().Equal(types.TypeMsgUndelegate, msg.Type())
	suite.Require().Equal("cosmosvaloper1gnkw3uqzflagcqn6ekjwpjanlne928qhruemah", msg.ValidatorAddress)
	suite.Require().Len(futureOperations, 0)
}

// TestSimulateMsgBeginRedelegate tests the normal scenario of a valid message of type TypeMsgBeginRedelegate.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgBeginRedelegate() {
	app, ctx, accounts := suite.app, suite.ctx, suite.accounts

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// setup accounts[0] as validator0 and accounts[1] as validator1
	validator0 := getTestingValidator0(suite.T(), suite.stakingKeeper, ctx, accounts)
	validator1 := getTestingValidator1(suite.T(), suite.stakingKeeper, ctx, accounts)

	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)

	// setup accounts[2] as delegator
	delegator := accounts[2]
	delegation := types.NewDelegation(delegator.Address, validator1.GetOperator(), issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, delegation)
	suite.distrKeeper.SetDelegatorStartingInfo(ctx, validator1.GetOperator(), delegator.Address, distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	setupValidatorRewards(suite.distrKeeper, ctx, validator0.GetOperator())
	setupValidatorRewards(suite.distrKeeper, ctx, validator1.GetOperator())

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgBeginRedelegate(suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, app.BaseApp, ctx, accounts, "")
	suite.Require().NoError(err)

	var msg types.MsgBeginRedelegate
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmos1092v0qgulpejj8y8hs6dmlw82x9gv8f7jfc7jl", msg.DelegatorAddress)
	suite.Require().Equal("1883752832348281252", msg.Amount.Amount.String())
	suite.Require().Equal("stake", msg.Amount.Denom)
	suite.Require().Equal(types.TypeMsgBeginRedelegate, msg.Type())
	suite.Require().Equal("cosmosvaloper1gnkw3uqzflagcqn6ekjwpjanlne928qhruemah", msg.ValidatorDstAddress)
	suite.Require().Equal("cosmosvaloper1kk653svg7ksj9fmu85x9ygj4jzwlyrgs89nnn2", msg.ValidatorSrcAddress)
	suite.Require().Len(futureOperations, 0)
}

func getTestingValidator0(t *testing.T, stakingKeeper *keeper.Keeper, ctx sdk.Context, accounts []simtypes.Account) types.Validator {
	commission0 := types.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return getTestingValidator(t, stakingKeeper, ctx, accounts, commission0, 0)
}

func getTestingValidator1(t *testing.T, stakingKeeper *keeper.Keeper, ctx sdk.Context, accounts []simtypes.Account) types.Validator {
	commission1 := types.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	return getTestingValidator(t, stakingKeeper, ctx, accounts, commission1, 1)
}

func getTestingValidator(t *testing.T, stakingKeeper *keeper.Keeper, ctx sdk.Context, accounts []simtypes.Account, commission types.Commission, n int) types.Validator {
	account := accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := teststaking.NewValidator(t, valAddr, valPubKey)
	validator, err := validator.SetInitialCommission(commission)
	require.NoError(t, err)

	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = stakingKeeper.TokensFromConsensusPower(ctx, 100)

	stakingKeeper.SetValidator(ctx, validator)

	return validator
}

func setupValidatorRewards(distrKeeper distrkeeper.Keeper, ctx sdk.Context, valAddress sdk.ValAddress) {
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, sdk.OneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	distrKeeper.SetValidatorHistoricalRewards(ctx, valAddress, 2, historicalRewards)
	// setup current revards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	distrKeeper.SetValidatorCurrentRewards(ctx, valAddress, currentRewards)
}
