package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SimTestSuite struct {
	suite.Suite

	ctx      sdk.Context
	r        *rand.Rand
	accounts []simtypes.Account

	app               *sdkapp.SDKApp
	legacyAmino       *codec.LegacyAmino
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	slashingKeeper    slashingkeeper.Keeper
	distrKeeper       distributionkeeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	s := rand.NewSource(1)
	suite.r = rand.New(s)
	accounts := simtypes.RandomAccounts(suite.r, 4)

	ta := testapp.Setup(suite.T())
	suite.app = ta
	suite.legacyAmino = ta.LegacyAmino()
	suite.codec = ta.AppCodec()
	suite.interfaceRegistry = ta.InterfaceRegistry()
	suite.txConfig = ta.TxConfig()
	suite.accountKeeper = ta.AccountKeeper
	suite.bankKeeper = ta.BankKeeper
	suite.stakingKeeper = ta.StakingKeeper
	suite.slashingKeeper = ta.SlashingKeeper
	suite.distrKeeper = ta.DistrKeeper
	suite.ctx = testapp.NewContext(ta)

	// accounts[0] used as validator in tests, rest as delegators
	suite.accounts = accounts[1:]

	initAmt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range suite.accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	suite.Require().NoError(ta.MintKeeper.Params.Set(suite.ctx, minttypes.DefaultParams()))
	suite.Require().NoError(ta.MintKeeper.Minter.Set(suite.ctx, minttypes.DefaultInitialMinter()))
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

// TestWeightedOperations tests the weights of the operations.
func (suite *SimTestSuite) TestWeightedOperations() {
	ctx := suite.ctx.WithChainID("test-chain")
	appParams := make(simtypes.AppParams)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgUnjail, types.ModuleName, sdk.MsgTypeURL(&types.MsgUnjail{})},
	}

	weightedOps := simulation.WeightedOperations(suite.interfaceRegistry, appParams, suite.codec, suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.slashingKeeper, suite.stakingKeeper)
	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(suite.r, suite.app.BaseApp, ctx, suite.accounts, ctx.ChainID())
		suite.Require().NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgUnjail tests the normal scenario of a valid message of type types.MsgUnjail.
// Abnormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgUnjail() {
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithBlockTime(blockTime)

	// setup accounts[0] as validator0
	validator0, err := getTestingValidator0(ctx, suite.stakingKeeper, suite.accounts)
	suite.Require().NoError(err)

	// setup validator0 by consensus address
	suite.Require().NoError(suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator0))
	val0ConsAddress, err := validator0.GetConsAddr()
	suite.Require().NoError(err)
	info := types.NewValidatorSigningInfo(val0ConsAddress, int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	suite.Require().NoError(suite.slashingKeeper.SetValidatorSigningInfo(ctx, val0ConsAddress, info))

	// put validator0 in jail
	suite.Require().NoError(suite.stakingKeeper.Jail(ctx, val0ConsAddress))

	// setup self delegation
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	val0AccAddress, err := sdk.ValAddressFromBech32(validator0.OperatorAddress)
	suite.Require().NoError(err)

	selfDelegation := stakingtypes.NewDelegation(suite.accounts[0].Address.String(), validator0.GetOperator(), issuedShares)
	suite.Require().NoError(suite.stakingKeeper.SetDelegation(ctx, selfDelegation))
	suite.Require().NoError(suite.distrKeeper.SetDelegatorStartingInfo(ctx, val0AccAddress, val0AccAddress.Bytes(), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	// begin a new block
	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.app.LastBlockHeight() + 1, Hash: suite.app.LastCommitID().Hash, Time: blockTime})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUnjail(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.slashingKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, ctx, suite.accounts, suite.ctx.ChainID())
	suite.Require().NoError(err)

	var msg types.MsgUnjail
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	suite.Require().NoError(err)
	suite.Require().True(operationMsg.OK)
	suite.Require().Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddr)
	suite.Require().Len(futureOperations, 0)
}

func getTestingValidator0(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, accounts []simtypes.Account) (stakingtypes.Validator, error) {
	commission0 := stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyOneDec(), math.LegacyOneDec())
	return getTestingValidator(ctx, stakingKeeper, accounts, commission0, 0)
}

func getTestingValidator(ctx sdk.Context, stakingKeeper *stakingkeeper.Keeper, accounts []simtypes.Account, commission stakingtypes.Commission, n int) (stakingtypes.Validator, error) {
	account := accounts[n]
	valPubKey := account.ConsKey.PubKey()
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr.String(), valPubKey, stakingtypes.Description{})
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to create validator: %w", err)
	}

	validator, err = validator.SetInitialCommission(commission)
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to set initial commission: %w", err)
	}

	validator.DelegatorShares = math.LegacyNewDec(100)
	validator.Tokens = math.NewInt(1000000)

	err = stakingKeeper.SetValidator(ctx, validator)
	if err != nil {
		return stakingtypes.Validator{}, fmt.Errorf("failed to set validator: %w", err)
	}

	return validator, nil
}
