package simulation_test

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TestWeightedOperations tests the weights of the operations.
func TestWeightedOperations(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)
	app, ctx, accs := createTestApp(t, false, r, 3)
	ctx.WithChainID("test-chain")

	cdc := app.AppCodec()
	appParams := make(simtypes.AppParams)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{{simappparams.DefaultWeightMsgUnjail, types.ModuleName, types.TypeMsgUnjail}}

	weightesOps := simulation.WeightedOperations(appParams, cdc, app.AccountKeeper, app.BankKeeper, app.SlashingKeeper, app.StakingKeeper)
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

// TestSimulateMsgUnjail tests the normal scenario of a valid message of type types.MsgUnjail.
// Abonormal scenarios, where the message is created by an errors, are not tested here.
func TestSimulateMsgUnjail(t *testing.T) {
	// setup 3 accounts
	s := rand.NewSource(5)
	r := rand.New(s)
	app, ctx, accounts := createTestApp(t, false, r, 3)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// remove genesis validator account
	accounts = accounts[1:]

	// setup accounts[0] as validator0
	validator0 := getTestingValidator0(t, app, ctx, accounts)

	// setup validator0 by consensus address
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validator0)
	val0ConsAddress, err := validator0.GetConsAddr()
	require.NoError(t, err)
	info := types.NewValidatorSigningInfo(val0ConsAddress, int64(4), int64(3),
		time.Unix(2, 0), false, int64(10))
	app.SlashingKeeper.SetValidatorSigningInfo(ctx, val0ConsAddress, info)

	// put validator0 in jail
	app.StakingKeeper.Jail(ctx, val0ConsAddress)

	// setup self delegation
	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
	validator0, issuedShares := validator0.AddTokensFromDel(delTokens)
	val0AccAddress, err := sdk.ValAddressFromBech32(validator0.OperatorAddress)
	require.NoError(t, err)
	selfDelegation := stakingtypes.NewDelegation(val0AccAddress.Bytes(), validator0.GetOperator(), issuedShares)
	app.StakingKeeper.SetDelegation(ctx, selfDelegation)
	app.DistrKeeper.SetDelegatorStartingInfo(ctx, validator0.GetOperator(), val0AccAddress.Bytes(), distrtypes.NewDelegatorStartingInfo(2, sdk.OneDec(), 200))

	// begin a new block
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: app.LastBlockHeight() + 1, AppHash: app.LastCommitID().Hash, Time: blockTime}})

	// execute operation
	op := simulation.SimulateMsgUnjail(app.AccountKeeper, app.BankKeeper, app.SlashingKeeper, app.StakingKeeper)
	operationMsg, futureOperations, err := op(r, app.BaseApp, ctx, accounts, "")
	require.NoError(t, err)

	var msg types.MsgUnjail
	legacy.Cdc.UnmarshalJSON(operationMsg.Msg, &msg)

	require.True(t, operationMsg.OK)
	require.Equal(t, types.TypeMsgUnjail, msg.Type())
	require.Equal(t, "cosmosvaloper17s94pzwhsn4ah25tec27w70n65h5t2scgxzkv2", msg.ValidatorAddr)
	require.Len(t, futureOperations, 0)
}

// returns context and an app with updated mint keeper
func createTestApp(t *testing.T, isCheckTx bool, r *rand.Rand, n int) (*simapp.SimApp, sdk.Context, []simtypes.Account) {
	accounts := simtypes.RandomAccounts(r, n)
	// create validator set with single validator
	account := accounts[0]
	tmPk, err := cryptocodec.ToTmPubKeyInterface(account.PubKey)
	require.NoError(t, err)
	validator := tmtypes.NewValidator(tmPk, 1)

	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	app := simapp.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)

	ctx := app.BaseApp.NewContext(isCheckTx, tmproto.Header{})
	initAmt := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// remove genesis validator account
	accs := accounts[1:]

	// add coins to the accounts
	for _, account := range accs {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
		require.NoError(t, testutil.FundAccount(app.BankKeeper, ctx, account.Address, initCoins))
	}

	app.MintKeeper.SetParams(ctx, minttypes.DefaultParams())
	app.MintKeeper.SetMinter(ctx, minttypes.DefaultInitialMinter())

	return app, ctx, accounts
}

func getTestingValidator0(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account) stakingtypes.Validator {
	commission0 := stakingtypes.NewCommission(sdk.ZeroDec(), sdk.OneDec(), sdk.OneDec())
	return getTestingValidator(t, app, ctx, accounts, commission0, 0)
}

func getTestingValidator(t *testing.T, app *simapp.SimApp, ctx sdk.Context, accounts []simtypes.Account, commission stakingtypes.Commission, n int) stakingtypes.Validator {
	account := accounts[n]
	valPubKey := account.ConsKey.PubKey()
	valAddr := sdk.ValAddress(account.PubKey.Address().Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, valPubKey, stakingtypes.Description{})
	require.NoError(t, err)
	validator, err = validator.SetInitialCommission(commission)
	require.NoError(t, err)

	validator.DelegatorShares = sdk.NewDec(100)
	validator.Tokens = sdk.NewInt(1000000)

	app.StakingKeeper.SetValidator(ctx, validator)

	return validator
}
