package simulation_test

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/depinject"
	sdklog "cosmossdk.io/log"
	"cosmossdk.io/math"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktestutil "cosmossdk.io/x/bank/testutil"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	minttypes "cosmossdk.io/x/mint/types"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/simulation"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/tests/integration/staking"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

type SimTestSuite struct {
	suite.Suite

	r             *rand.Rand
	txConfig      client.TxConfig
	accounts      []simtypes.Account
	ctx           sdk.Context
	app           *runtime.App
	bankKeeper    bankkeeper.Keeper
	accountKeeper authkeeper.AccountKeeper
	distrKeeper   distrkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper

	encCfg moduletestutil.TestEncodingConfig
}

func (s *SimTestSuite) SetupTest() {
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	s.r = rand.New(rand.NewSource(1))
	accounts := simtypes.RandomAccounts(s.r, 4)

	// create genesis accounts
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	accs := []simtestutil.GenesisAccount{
		{GenesisAccount: acc, Coins: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000)))},
	}

	// create validator set with single validator
	account := accounts[0]
	cmtPk, err := cryptocodec.ToCmtPubKeyInterface(account.ConsKey.PubKey())
	require.NoError(s.T(), err)
	validator := cmttypes.NewValidator(cmtPk, 1)

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = accs
	startupCfg.ValidatorSet = func() (*cmttypes.ValidatorSet, error) {
		return cmttypes.NewValidatorSet([]*cmttypes.Validator{validator}), nil
	}

	var (
		accountKeeper authkeeper.AccountKeeper
		mintKeeper    mintkeeper.Keeper
		bankKeeper    bankkeeper.Keeper
		distrKeeper   distrkeeper.Keeper
		stakingKeeper *stakingkeeper.Keeper
	)

	cfg := depinject.Configs(
		staking.AppConfig,
		depinject.Supply(sdklog.NewNopLogger()),
	)

	app, err := simtestutil.SetupWithConfiguration(cfg, startupCfg, &s.txConfig, &bankKeeper, &accountKeeper, &mintKeeper, &distrKeeper, &stakingKeeper)
	require.NoError(s.T(), err)

	ctx := app.BaseApp.NewContext(false)
	s.Require().NoError(mintKeeper.Params.Set(ctx, minttypes.DefaultParams()))
	s.Require().NoError(mintKeeper.Minter.Set(ctx, minttypes.DefaultInitialMinter()))

	initAmt := stakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	s.accounts = accounts
	// remove genesis validator account
	// add coins to the accounts
	for _, account := range accounts[1:] {
		acc := accountKeeper.NewAccountWithAddress(ctx, account.Address)
		accountKeeper.SetAccount(ctx, acc)
		s.Require().NoError(banktestutil.FundAccount(ctx, bankKeeper, account.Address, initCoins))
	}

	s.accountKeeper = accountKeeper
	s.bankKeeper = bankKeeper
	s.distrKeeper = distrKeeper
	s.stakingKeeper = stakingKeeper
	s.ctx = ctx
	s.app = app
}

// TestWeightedOperations tests the weights of the operations.
func (s *SimTestSuite) TestWeightedOperations() {
	require := s.Require()

	s.ctx.WithChainID("test-chain")

	cdc := s.encCfg.Codec
	appParams := make(simtypes.AppParams)

	weightedOps := simulation.WeightedOperations(appParams, cdc, s.txConfig, s.accountKeeper,
		s.bankKeeper, s.stakingKeeper,
	)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgCreateValidator, types.ModuleName, sdk.MsgTypeURL(&types.MsgCreateValidator{})},
		{simulation.DefaultWeightMsgEditValidator, types.ModuleName, sdk.MsgTypeURL(&types.MsgEditValidator{})},
		{simulation.DefaultWeightMsgDelegate, types.ModuleName, sdk.MsgTypeURL(&types.MsgDelegate{})},
		{simulation.DefaultWeightMsgUndelegate, types.ModuleName, sdk.MsgTypeURL(&types.MsgUndelegate{})},
		{simulation.DefaultWeightMsgBeginRedelegate, types.ModuleName, sdk.MsgTypeURL(&types.MsgBeginRedelegate{})},
		{simulation.DefaultWeightMsgCancelUnbondingDelegation, types.ModuleName, sdk.MsgTypeURL(&types.MsgCancelUnbondingDelegation{})},
		{simulation.DefaultWeightMsgRotateConsPubKey, types.ModuleName, sdk.MsgTypeURL(&types.MsgRotateConsPubKey{})},
	}

	for i, w := range weightedOps {
		operationMsg, _, _ := w.Op()(s.r, s.app.BaseApp, s.ctx, s.accounts, s.ctx.ChainID())
		// require.NoError(t, err) // TODO check if it should be NoError

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		require.Equal(expected[i].weight, w.Weight(), "weight should be the same")
		require.Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		require.Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestSimulateMsgCreateValidator tests the normal scenario of a valid message of type TypeMsgCreateValidator.
// Abonormal scenarios, where the message are created by an errors are not tested here.
func (s *SimTestSuite) TestSimulateMsgCreateValidator() {
	require := s.Require()
	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash})
	require.NoError(err)
	// execute operation
	op := simulation.SimulateMsgCreateValidator(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, s.ctx, s.accounts[1:], "")
	require.NoError(err)

	var msg types.MsgCreateValidator
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(sdk.MsgTypeURL(&types.MsgCreateValidator{}), sdk.MsgTypeURL(&msg))
	valaddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	require.NoError(err)
	require.Equal("cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", sdk.AccAddress(valaddr).String())
	require.Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

// TestSimulateMsgCancelUnbondingDelegation tests the normal scenario of a valid message of type TypeMsgCancelUnbondingDelegation.
// Abonormal scenarios, where the message is
func (s *SimTestSuite) TestSimulateMsgCancelUnbondingDelegation() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup accounts[1] as validator
	validator0 := s.getTestingValidator0(ctx)

	// setup delegation
	delTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := s.accounts[2]
	delegation := types.NewDelegation(delegator.Address.String(), validator0.GetOperator(), issuedShares)
	require.NoError(s.stakingKeeper.SetDelegation(ctx, delegation))
	val0bz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator0.GetOperator())
	s.Require().NoError(err)
	s.Require().NoError(s.distrKeeper.DelegatorStartingInfo.Set(ctx, collections.Join(sdk.ValAddress(val0bz), delegator.Address), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	s.setupValidatorRewards(ctx, val0bz)

	// unbonding delegation
	udb := types.NewUnbondingDelegation(delegator.Address, val0bz, s.app.LastBlockHeight()+1, blockTime.Add(2*time.Minute), delTokens, 0, address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"))
	require.NoError(s.stakingKeeper.SetUnbondingDelegation(ctx, udb))
	s.setupValidatorRewards(ctx, val0bz)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash, Time: blockTime})
	require.NoError(err)
	// execute operation
	op := simulation.SimulateMsgCancelUnbondingDelegate(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	accounts := []simtypes.Account{delegator}
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, accounts, "")
	require.NoError(err)

	var msg types.MsgCancelUnbondingDelegation
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(sdk.MsgTypeURL(&types.MsgCancelUnbondingDelegation{}), sdk.MsgTypeURL(&msg))
	require.Equal(delegator.Address.String(), msg.DelegatorAddress)
	require.Equal(validator0.GetOperator(), msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

// TestSimulateMsgEditValidator tests the normal scenario of a valid message of type TypeMsgEditValidator.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (s *SimTestSuite) TestSimulateMsgEditValidator() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup accounts[0] as validator
	_ = s.getTestingValidator0(ctx)

	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash, Time: blockTime})
	require.NoError(err)
	// execute operation
	op := simulation.SimulateMsgEditValidator(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, s.accounts, "")
	require.NoError(err)

	var msg types.MsgEditValidator
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal(sdk.MsgTypeURL(&types.MsgEditValidator{}), sdk.MsgTypeURL(&msg))
	require.Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

// TestSimulateMsgDelegate tests the normal scenario of a valid message of type TypeMsgDelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (s *SimTestSuite) TestSimulateMsgDelegate() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// execute operation
	op := simulation.SimulateMsgDelegate(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, s.accounts[1:], "")
	require.NoError(err)

	var msg types.MsgDelegate
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal("cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.DelegatorAddress)
	require.Equal("stake", msg.Amount.Denom)
	require.Equal(sdk.MsgTypeURL(&types.MsgDelegate{}), sdk.MsgTypeURL(&msg))
	require.Equal("cosmosvaloper122js6qry7nlgp63gcse8muknspuxur77vj3kkr", msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

// TestSimulateMsgUndelegate tests the normal scenario of a valid message of type TypeMsgUndelegate.
// Abonormal scenarios, where the message is created by an errors are not tested here.
func (s *SimTestSuite) TestSimulateMsgUndelegate() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup accounts[1] as validator
	validator0 := s.getTestingValidator0(ctx)

	// setup delegation
	delTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	delegator := s.accounts[2]
	delegation := types.NewDelegation(delegator.Address.String(), validator0.GetOperator(), issuedShares)
	require.NoError(s.stakingKeeper.SetDelegation(ctx, delegation))
	val0bz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator0.GetOperator())
	s.Require().NoError(err)
	s.Require().NoError(s.distrKeeper.DelegatorStartingInfo.Set(ctx, collections.Join(sdk.ValAddress(val0bz), delegator.Address), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	s.setupValidatorRewards(ctx, val0bz)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash, Time: blockTime})
	require.NoError(err)
	// execute operation
	op := simulation.SimulateMsgUndelegate(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, s.accounts, "")
	require.NoError(err)

	var msg types.MsgUndelegate
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.DelegatorAddress)
	require.Equal("1646627814093010272", msg.Amount.Amount.String())
	require.Equal("stake", msg.Amount.Denom)
	require.Equal(sdk.MsgTypeURL(&types.MsgUndelegate{}), sdk.MsgTypeURL(&msg))
	require.Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

// TestSimulateMsgBeginRedelegate tests the normal scenario of a valid message of type TypeMsgBeginRedelegate.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func (s *SimTestSuite) TestSimulateMsgBeginRedelegate() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	// setup accounts[1] as validator0 and accounts[2] as validator1
	validator0 := s.getTestingValidator0(ctx)
	validator1 := s.getTestingValidator1(ctx)

	delTokens := s.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator1, issuedShares := validator1.AddTokensFromDel(delTokens)

	// setup accounts[3] as delegator
	delegator := s.accounts[3]
	delegation := types.NewDelegation(delegator.Address.String(), validator0.GetOperator(), issuedShares)
	require.NoError(s.stakingKeeper.SetDelegation(ctx, delegation))
	val0bz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator0.GetOperator())
	s.Require().NoError(err)
	val1bz, err := s.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator1.GetOperator())
	s.Require().NoError(err)
	s.Require().NoError(s.distrKeeper.DelegatorStartingInfo.Set(ctx, collections.Join(sdk.ValAddress(val0bz), delegator.Address), distrtypes.NewDelegatorStartingInfo(2, math.LegacyOneDec(), 200)))

	s.setupValidatorRewards(ctx, val0bz)
	s.setupValidatorRewards(ctx, val1bz)

	_, err = s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash, Time: blockTime})
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgBeginRedelegate(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, s.accounts, "")
	s.T().Logf("operation message: %v", operationMsg)
	require.NoError(err)

	var msg types.MsgBeginRedelegate
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)
	require.True(operationMsg.OK)
	require.Equal("cosmos1ua0fwyws7vzjrry3pqkklvf8mny93l9s9zg0h4", msg.DelegatorAddress)
	require.Equal("stake", msg.Amount.Denom)
	require.Equal(sdk.MsgTypeURL(&types.MsgBeginRedelegate{}), sdk.MsgTypeURL(&msg))
	require.Equal("cosmosvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnsvnaes", msg.ValidatorDstAddress)
	require.Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorSrcAddress)
	require.Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateRotateConsPubKey() {
	require := s.Require()
	blockTime := time.Now().UTC()
	ctx := s.ctx.WithHeaderInfo(header.Info{Time: blockTime})

	_ = s.getTestingValidator2(ctx)

	// begin a new block
	_, err := s.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: s.app.LastBlockHeight() + 1, Hash: s.app.LastCommitID().Hash, Time: blockTime})
	require.NoError(err)

	// execute operation
	op := simulation.SimulateMsgRotateConsPubKey(s.txConfig, s.accountKeeper, s.bankKeeper, s.stakingKeeper)
	operationMsg, futureOperations, err := op(s.r, s.app.BaseApp, ctx, s.accounts, "")
	require.NoError(err)

	var msg types.MsgRotateConsPubKey
	err = proto.Unmarshal(operationMsg.Msg, &msg)
	require.NoError(err)

	require.True(operationMsg.OK)
	require.Equal(sdk.MsgTypeURL(&types.MsgRotateConsPubKey{}), sdk.MsgTypeURL(&msg))
	require.Equal("cosmosvaloper1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7epjs3u", msg.ValidatorAddress)
	require.Len(futureOperations, 0)
}

func (s *SimTestSuite) getTestingValidator0(ctx sdk.Context) types.Validator {
	commission0 := types.NewCommission(math.LegacyZeroDec(), math.LegacyOneDec(), math.LegacyOneDec())
	return s.getTestingValidator(ctx, commission0, 1)
}

func (s *SimTestSuite) getTestingValidator1(ctx sdk.Context) types.Validator {
	commission1 := types.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
	return s.getTestingValidator(ctx, commission1, 2)
}

func (s *SimTestSuite) getTestingValidator(ctx sdk.Context, commission types.Commission, n int) types.Validator {
	account := s.accounts[n]
	valPubKey := account.PubKey
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator := testutil.NewValidator(s.T(), valAddr, valPubKey)
	validator, err := validator.SetInitialCommission(commission)
	s.Require().NoError(err)

	validator.DelegatorShares = math.LegacyNewDec(100)
	validator.Tokens = s.stakingKeeper.TokensFromConsensusPower(ctx, 100)

	s.Require().NoError(s.stakingKeeper.SetValidator(ctx, validator))

	return validator
}

func (s *SimTestSuite) getTestingValidator2(ctx sdk.Context) types.Validator {
	val := s.getTestingValidator0(ctx)
	val.Status = types.Bonded
	s.Require().NoError(s.stakingKeeper.SetValidator(ctx, val))
	s.Require().NoError(s.stakingKeeper.SetValidatorByConsAddr(ctx, val))
	return val
}

func (s *SimTestSuite) setupValidatorRewards(ctx sdk.Context, valAddress sdk.ValAddress) {
	decCoins := sdk.DecCoins{sdk.NewDecCoinFromDec(sdk.DefaultBondDenom, math.LegacyOneDec())}
	historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
	s.Require().NoError(s.distrKeeper.ValidatorHistoricalRewards.Set(ctx, collections.Join(valAddress, uint64(2)), historicalRewards))
	// setup current revards
	currentRewards := distrtypes.NewValidatorCurrentRewards(decCoins, 3)
	s.Require().NoError(s.distrKeeper.ValidatorCurrentRewards.Set(ctx, valAddress, currentRewards))
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
