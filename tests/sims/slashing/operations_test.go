package slashing

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	authkeeper "cosmossdk.io/x/auth/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	distributionkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	minttypes "cosmossdk.io/x/mint/types"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/simulation"
	"cosmossdk.io/x/slashing/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type SimTestSuite struct {
	suite.Suite

	ctx      sdk.Context
	r        *rand.Rand
	accounts []simtypes.Account

	app               *runtime.App
	legacyAmino       *codec.LegacyAmino
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	slashingKeeper    slashingkeeper.Keeper
	distrKeeper       distributionkeeper.Keeper
	mintKeeper        mintkeeper.Keeper
}

func (suite *SimTestSuite) SetupTest() {
	s := rand.NewSource(1)
	suite.r = rand.New(s)
	accounts := simtypes.RandomAccounts(suite.r, 4)

	// create validator (non random as using a seed)
	createValidator := func() (*cmttypes.ValidatorSet, error) {
		account := accounts[0]
		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(account.ConsKey.PubKey())
		if err != nil {
			return nil, fmt.Errorf("failed to create pubkey: %w", err)
		}

		validator := cmttypes.NewValidator(cmtPk, 1)

		return cmttypes.NewValidatorSet([]*cmttypes.Validator{validator}), nil
	}

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.ValidatorSet = createValidator

	app, err := simtestutil.SetupWithConfiguration(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		startupCfg,
		&suite.legacyAmino,
		&suite.codec,
		&suite.interfaceRegistry,
		&suite.txConfig,
		&suite.accountKeeper,
		&suite.bankKeeper,
		&suite.stakingKeeper,
		&suite.mintKeeper,
		&suite.slashingKeeper,
		&suite.distrKeeper,
	)

	suite.Require().NoError(err)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(false)

	// remove genesis validator account
	suite.accounts = accounts[1:]

	initAmt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range suite.accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}
	suite.Require().NoError(suite.mintKeeper.Params.Set(suite.ctx, minttypes.DefaultParams()))
	suite.Require().NoError(suite.mintKeeper.Minter.Set(suite.ctx, minttypes.DefaultInitialMinter()))
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
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (suite *SimTestSuite) TestSimulateMsgUnjail() {
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup accounts[0] as validator0
	validator0, err := getTestingValidator0(ctx, suite.stakingKeeper, suite.accounts)
	suite.Require().NoError(err)

	// setup validator0 by consensus address
	err = suite.stakingKeeper.SetValidatorByConsAddr(ctx, validator0)
	suite.Require().NoError(err)

	val0ConsAddress, err := validator0.GetConsAddr()
	suite.Require().NoError(err)
	val0ConsAddressStr, err := suite.stakingKeeper.ConsensusAddressCodec().BytesToString(val0ConsAddress)
	suite.Require().NoError(err)
	info := types.NewValidatorSigningInfo(val0ConsAddressStr, int64(4),
		time.Unix(2, 0), false, int64(10))
	err = suite.slashingKeeper.ValidatorSigningInfo.Set(ctx, val0ConsAddress, info)
	suite.Require().NoError(err)
	// put validator0 in jail
	suite.Require().NoError(suite.stakingKeeper.Jail(ctx, val0ConsAddress))

	// setup self delegation
	delTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	val0AccAddress, err := sdk.ValAddressFromBech32(validator0.OperatorAddress)
	suite.Require().NoError(err)
	selfDelegation := stakingtypes.NewDelegation(suite.accounts[0].Address.String(), validator0.GetOperator(), issuedShares)
	suite.Require().NoError(suite.stakingKeeper.SetDelegation(ctx, selfDelegation))
	suite.Require().NoError(suite.distrKeeper.DelegatorStartingInfo.Set(ctx, collections.Join(val0AccAddress, sdk.AccAddress(val0AccAddress)), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	// begin a new block
	_, err = suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: suite.app.LastBlockHeight() + 1, Hash: suite.app.LastCommitID().Hash, Time: blockTime})
	suite.Require().NoError(err)

	// execute operation
	op := simulation.SimulateMsgUnjail(codec.NewProtoCodec(suite.interfaceRegistry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.slashingKeeper, suite.stakingKeeper)
	operationMsg, futureOperations, err := op(suite.r, suite.app.BaseApp, ctx, suite.accounts, "")
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
		return stakingtypes.Validator{}, err
	}
	return validator, nil
}
