package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app               *runtime.App
	legacyAmino       *codec.LegacyAmino
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	slashingKeeper    slashingkeeper.Keeper
	distrKeeper       distributionkeeper.Keeper
	mintKeeper        mintkeeper.Keeper
	accs              []simtypes.Account
}

func (suite *SimTestSuite) SetupTest() {
	app, err := simtestutil.Setup(
		testutil.AppConfig,
		&suite.legacyAmino,
		&suite.codec,
		&suite.interfaceRegistry,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.stakingKeeper,
		&suite.mintKeeper,
		&suite.slashingKeeper,
		&suite.distrKeeper,
	)

	suite.Require().NoError(err)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{})

	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := simtypes.RandomAccounts(r, 3)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	initAmt := suite.stakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(ctx, account.Address)
		suite.accountKeeper.SetAccount(ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.bankKeeper, ctx, account.Address, initCoins))
	}

	suite.mintKeeper.SetParams(ctx, minttypes.DefaultParams())
	suite.mintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())
	suite.accs = accounts
}

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accs := suite.app, suite.ctx, suite.accs
	ctx.WithChainID("test-chain")

	cdc := suite.codec
	appParams := make(simtypes.AppParams)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{{simtestutil.DefaultWeightMsgUnjail, types.ModuleName, types.TypeMsgUnjail}}

	weightesOps := simulation.WeightedOperations(appParams, cdc, suite.accountKeeper, suite.bankKeeper, suite.slashingKeeper, suite.stakingKeeper)
	for i, w := range weightesOps {
		operationMsg, _, err := w.Op()(r, app.BaseApp, ctx, accs, ctx.ChainID())
		require.NoError(t, err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(t, expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(t, expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(t, expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgUnjail tests the normal scenario of a valid message of type types.MsgUnjail.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgUnjail(t *testing.T) {
	// setup 3 accounts
	s := rand.NewSource(5)
	r := rand.New(s)
	app, ctx, accounts := suite.app, suite.ctx, suite.accs
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator0
	validator0 := suite.getTestingValidator0(ctx, accounts)

	// setup validator0 by consensus address
	suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator0)
	val0ConsAddress, err := validator0.GetConsAddr()
	require.NoError(t, err)
	info := types.NewValidatorSigningInfo(val0ConsAddress, int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	suite.slashingKeeper.SetValidatorSigningInfo(ctx, val0ConsAddress, info)

	// put validator0 in jail
	suite.stakingKeeper.Jail(ctx, val0ConsAddress)

	// setup self delegation
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	val0AccAddress, err := sdk.ValAddressFromBech32(validator0.OperatorAddress)
	require.NoError(t, err)
	selfDelegation := stakingtypes.NewDelegation(val0AccAddress.Bytes(), validator0.GetOperator(), issuedShares)
	suite.stakingKeeper.SetDelegation(ctx, selfDelegation)
	suite.distrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), val0AccAddress.Bytes(), distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgUnjail(codec.NewProtoCodec(suite.interfaceRegistry), suite.accountKeeper, suite.bankKeeper, suite.slashingKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgUnjail
	types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgUnjail, msg.Type())
	require.Equal(t, "cosmosvaloper17s94pzwhsn4ah25tec27w70n65h5t2scgxzkv2", msg.ValidatorAddr)
	require.Len(t, futureOperations, 0)
}

func (suite *SimTestSuite) getTestingValidator0(ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return suite.getTestingValidator(commission0, 0)
}

func (suite *SimTestSuite) getTestingValidator(commission stakingtypes.Commission, n int) stakingtypes.Validator {
	ctx, accounts := suite.ctx, suite.accs
	account := accounts[n]
	valPubKey := account.ConsKey.PubKey()
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, valPubKey, stakingtypes.Description{})
	suite.Require().NoError(err)
	validator, err = validator.SetInitialCommission(commission)
	suite.Require().NoError(err)

	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = sdk.NewInt(1000000)

	suite.stakingKeeper.SetValidator(ctx, validator)

	return validator
}
