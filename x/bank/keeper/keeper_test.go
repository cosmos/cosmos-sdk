package keeper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const (
	fooDenom     = "foo"
	barDenom     = "bar"
	initialPower = int64(100)
	holder       = "holder"
	multiPerm    = "multiple permissions account"
	randomPerm   = "random permission"
)

var (
	holderAcc     = authtypes.NewEmptyModuleAccount(holder)
	burnerAcc     = authtypes.NewEmptyModuleAccount(authtypes.Burner, authtypes.Burner)
	minterAcc     = authtypes.NewEmptyModuleAccount(authtypes.Minter, authtypes.Minter)
	multiPermAcc  = authtypes.NewEmptyModuleAccount(multiPerm, authtypes.Burner, authtypes.Minter, authtypes.Staking)
	randomPermAcc = authtypes.NewEmptyModuleAccount(randomPerm, "random")

	// The default power validators are initialized to have within tests
	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

func newFooCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(fooDenom, amt)
}

func newBarCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(barDenom, amt)
}

// nolint: interfacer
func getCoinsByName(ctx sdk.Context, bk keeper.Keeper, ak types.AccountKeeper, moduleName string) sdk.Coins {
	moduleAddress := ak.GetModuleAddress(moduleName)
	macc := ak.GetAccount(ctx, moduleAddress)
	if macc == nil {
		return sdk.Coins(nil)
	}

	return bk.GetAllBalances(ctx, macc.GetAddress())
}

type IntegrationTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *IntegrationTestSuite) initKeepersWithmAccPerms(blockedAddrs map[string]bool) (authkeeper.AccountKeeper, *keeper.BaseKeeper) {
	app := suite.app
	maccPerms := simapp.GetMaccPerms()
	appCodec := simapp.MakeTestEncodingConfig().Codec

	maccPerms[holder] = nil
	maccPerms[authtypes.Burner] = []string{authtypes.Burner}
	maccPerms[authtypes.Minter] = []string{authtypes.Minter}
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}
	maccPerms[randomPerm] = []string{"random"}
	authKeeper := authkeeper.NewAccountKeeper(
		appCodec, app.GetKey(types.StoreKey), app.GetSubspace(types.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)
	keeper := keeper.NewBaseKeeper(
		appCodec, app.GetKey(types.StoreKey), authKeeper,
		app.GetSubspace(types.ModuleName), blockedAddrs,
	)

	return authKeeper, keeper
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *IntegrationTestSuite) TestSupply() {
	ctx := suite.ctx

	require := suite.Require()

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	genesisSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)

	initialPower := int64(100)
	initTokens := suite.app.StakingKeeper.TokensFromConsensusPower(ctx, initialPower)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	require.NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	require.NoError(keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	total, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)

	expTotalSupply := initCoins.Add(genesisSupply...)
	require.Equal(expTotalSupply, total)

	// burning all supplied tokens
	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	require.NoError(err)

	total, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)
	require.Equal(total, genesisSupply)
}

func (suite *IntegrationTestSuite) TestSendCoinsFromModuleToAccount_Blocklist() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	_, keeper := suite.initKeepersWithmAccPerms(map[string]bool{addr1.String(): true})

	suite.Require().NoError(keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().Error(keeper.SendCoinsFromModuleToAccount(
		ctx, minttypes.ModuleName, addr1, initCoins,
	))
}

func (suite *IntegrationTestSuite) TestSupply_SendCoins() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	baseAcc := authKeeper.NewAccountWithAddress(ctx, authtypes.NewModuleAddress("baseAcc"))

	// set initial balances
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))

	suite.
		Require().
		NoError(keeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, holderAcc.GetAddress(), initCoins))

	authKeeper.SetModuleAccount(ctx, holderAcc)
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetAccount(ctx, baseAcc)

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, "", holderAcc.GetName(), initCoins) // nolint:errcheck
	})

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, authtypes.Burner, "", initCoins) // nolint:errcheck
	})

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins) // nolint:errcheck
	})

	suite.Require().Error(
		keeper.SendCoinsFromModuleToAccount(ctx, holderAcc.GetName(), baseAcc.GetAddress(), initCoins.Add(initCoins...)),
	)

	suite.Require().NoError(
		keeper.SendCoinsFromModuleToModule(ctx, holderAcc.GetName(), authtypes.Burner, initCoins),
	)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, holderAcc.GetName()).String())
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))

	suite.Require().NoError(
		keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Burner, baseAcc.GetAddress(), initCoins),
	)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	suite.Require().Equal(initCoins, keeper.GetAllBalances(ctx, baseAcc.GetAddress()))

	suite.Require().NoError(keeper.SendCoinsFromAccountToModule(ctx, baseAcc.GetAddress(), authtypes.Burner, initCoins))
	suite.Require().Equal(sdk.NewCoins().String(), keeper.GetAllBalances(ctx, baseAcc.GetAddress()).String())
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))
}

func (suite *IntegrationTestSuite) TestSupply_MintCoins() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetModuleAccount(ctx, minterAcc)
	authKeeper.SetModuleAccount(ctx, multiPermAcc)
	authKeeper.SetModuleAccount(ctx, randomPermAcc)

	initialSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	suite.Require().Panics(func() { keeper.MintCoins(ctx, "", initCoins) }, "no module account")                // nolint:errcheck
	suite.Require().Panics(func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }, "invalid permission") // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, sdk.Coins{sdk.Coin{Denom: "denom", Amount: sdk.NewInt(-10)}})
	suite.Require().Error(err, "insufficient coins")

	suite.Require().Panics(func() { keeper.MintCoins(ctx, randomPerm, initCoins) }) // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, initCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Minter))
	totalSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	suite.Require().Equal(initialSupply.Add(initCoins...), totalSupply)

	// test same functionality on module account with multiple permissions
	initialSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	err = keeper.MintCoins(ctx, multiPermAcc.GetName(), initCoins)
	suite.Require().NoError(err)

	totalSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()))
	suite.Require().Equal(initialSupply.Add(initCoins...), totalSupply)
	suite.Require().Panics(func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }) // nolint:errcheck
}

func (suite *IntegrationTestSuite) TestSupply_BurnCoins() {
	ctx := suite.ctx
	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	suite.
		Require().
		NoError(keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	// inflate supply
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	supplyAfterInflation, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})

	suite.Require().Panics(func() { keeper.BurnCoins(ctx, "", initCoins) }, "no module account")                    // nolint:errcheck
	suite.Require().Panics(func() { keeper.BurnCoins(ctx, authtypes.Minter, initCoins) }, "invalid permission")     // nolint:errcheck
	suite.Require().Panics(func() { keeper.BurnCoins(ctx, randomPerm, supplyAfterInflation) }, "random permission") // nolint:errcheck
	err = keeper.BurnCoins(ctx, authtypes.Burner, supplyAfterInflation)
	suite.Require().Error(err, "insufficient coins")

	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	suite.Require().NoError(err)
	supplyAfterBurn, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	suite.Require().Equal(supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)

	// test same functionality on module account with multiple permissions
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))

	supplyAfterInflation, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().NoError(keeper.SendCoins(ctx, authtypes.NewModuleAddress(authtypes.Minter), multiPermAcc.GetAddress(), initCoins))
	authKeeper.SetModuleAccount(ctx, multiPermAcc)

	err = keeper.BurnCoins(ctx, multiPermAcc.GetName(), initCoins)
	supplyAfterBurn, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()).String())
	suite.Require().Equal(supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)
}

func (suite *IntegrationTestSuite) TestSendCoinsNewAccount() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(app.AccountKeeper.GetAccount(ctx, addr2))
	app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Empty(app.BankKeeper.GetAllBalances(ctx, addr2))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(50))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	acc1Balances = app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(sendAmt, acc2Balances)
	updatedAcc1Bal := balances.Sub(sendAmt...)
	suite.Require().Len(acc1Balances, len(updatedAcc1Bal))
	suite.Require().Equal(acc1Balances, updatedAcc1Bal)
	suite.Require().NotNil(app.AccountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputNewAccount() {
	app, ctx := suite.app, suite.ctx

	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(app.AccountKeeper.GetAccount(ctx, addr2))
	suite.Require().Empty(app.BankKeeper.GetAllBalances(ctx, addr2))

	inputs := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)
	suite.Require().NotNil(app.AccountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(90), newBarCoin(30))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)

	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	app.AccountKeeper.SetAccount(ctx, acc3)

	inputs := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, []types.Output{}))
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	insufficientInputs := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
	}
	insufficientOutputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
	}
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, insufficientInputs, insufficientOutputs))
	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)

	acc3Balances := app.BankKeeper.GetAllBalances(ctx, addr3)
	suite.Require().Equal(expected, acc3Balances)
}

func (suite *IntegrationTestSuite) TestInputOutputCoinsWithQuarantine() {
	app, ctx := suite.app, suite.ctx

	// makeAndFundAccount makes and (if balance isn't zero) funds an account.
	makeAndFundAccount := func(i uint8, balance sdk.Coins) sdk.AccAddress {
		addr := sdk.AccAddress(fmt.Sprintf("testaddr%03d_________", i))
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
		if !balance.IsZero() {
			suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balance), "funding account %d with %q", i, balance.String())
		}
		return addr
	}

	// cz converts the string to a Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// makeInput makes an Input.
	makeInput := func(addr sdk.AccAddress, coins string) types.Input {
		return types.Input{
			Address: addr.String(),
			Coins:   cz(coins),
		}
	}

	// makeOutput makes an Output.
	makeOutput := func(addr sdk.AccAddress, coins string) types.Output {
		return types.Output{
			Address: addr.String(),
			Coins:   cz(coins),
		}
	}

	// makeInEvent makes an event expected from an input.
	makeInEvent := func(addr sdk.AccAddress) sdk.Event {
		return sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, addr.String()),
		)
	}

	// makeOutEvent makes an event expected from an output.
	makeOutEvent := func(addr sdk.AccAddress, amt string) sdk.Event {
		return sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeyRecipient, addr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt),
		)
	}

	type expectedBalance struct {
		addr    sdk.AccAddress
		balance sdk.Coins
	}

	makeExpectedBalance := func(addr sdk.AccAddress, coins string) *expectedBalance {
		return &expectedBalance{
			addr:    addr,
			balance: cz(coins),
		}
	}

	makeQc := func(coins string, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) *QuarantinedCoins {
		return &QuarantinedCoins{
			coins:     cz(coins),
			toAddr:    toAddr,
			fromAddrs: fromAddrs,
		}
	}

	fundsHolder := sdk.AccAddress("quarantinefundholder")
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, fundsHolder))

	addr1 := makeAndFundAccount(1, cz("500acoin,500bcoin,500ccoin"))
	addr2 := makeAndFundAccount(2, cz("500acoin,500bcoin,500ccoin"))
	addr3 := makeAndFundAccount(3, cz("500acoin,500bcoin,500ccoin"))

	tests := []struct {
		name      string
		inputs    []types.Input
		outputs   []types.Output
		qk        types.QuarantineKeeper
		expInErr  []string
		expEvents sdk.Events
		expBals   []*expectedBalance
		expQ      []*QuarantinedCoins
	}{
		{
			name: "nil quarantine keeper",
			inputs: []types.Input{
				makeInput(addr1, "5acoin"),
				makeInput(addr2, "5acoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10acoin"),
			},
			qk:       nil,
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("5acoin")),
				makeInEvent(addr1),
				types.NewCoinSpentEvent(addr2, cz("5acoin")),
				makeInEvent(addr2),
				types.NewCoinReceivedEvent(addr3, cz("10acoin")),
				makeOutEvent(addr3, "10acoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,500bcoin,500ccoin"),
			},
			expQ: nil,
		},
		{
			name:    "no funds holder",
			inputs:  []types.Input{makeInput(addr1, "5acoin")},
			outputs: []types.Output{makeOutput(addr2, "5acoin")},
			qk: NewMockMockQuarantineKeeper(nil).
				WithIsQuarantinedAddrResponse(addr2, true),
			expInErr:  []string{"no quarantine holder account defined", "unknown address"},
			expEvents: nil,
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "490acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "495acoin,500bcoin,500ccoin"),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name: "with quarantine keeper but not quarantined",
			inputs: []types.Input{
				makeInput(addr3, "10acoin"),
			},
			outputs: []types.Output{
				makeOutput(addr1, "6acoin"),
				makeOutput(addr2, "4acoin"),
			},
			qk:       NewMockMockQuarantineKeeper(fundsHolder),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr3, cz("10acoin")),
				makeInEvent(addr3),
				types.NewCoinReceivedEvent(addr1, cz("6acoin")),
				makeOutEvent(addr1, "6acoin"),
				types.NewCoinReceivedEvent(addr2, cz("4acoin")),
				makeOutEvent(addr2, "4acoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "496acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "499acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr3, "500acoin,500bcoin,500ccoin"),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name: "quarantined with auto-accept",
			inputs: []types.Input{
				makeInput(addr1, "16acoin"),
				makeInput(addr2, "9acoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "25acoin"),
			},
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr3, true).
				WithIsAutoAcceptResponse(addr3, []sdk.AccAddress{addr1, addr2}, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("16acoin")),
				makeInEvent(addr1),
				types.NewCoinSpentEvent(addr2, cz("9acoin")),
				makeInEvent(addr2),
				types.NewCoinReceivedEvent(addr3, cz("25acoin")),
				makeOutEvent(addr3, "25acoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "480acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "490acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr3, "525acoin,500bcoin,500ccoin"),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name: "one output that is quarantined",
			inputs: []types.Input{
				makeInput(addr1, "5bcoin"),
				makeInput(addr2, "5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "5bcoin,5ccoin"),
			},
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr3, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("5bcoin")),
				makeInEvent(addr1),
				types.NewCoinSpentEvent(addr2, cz("5ccoin")),
				makeInEvent(addr2),
				types.NewCoinReceivedEvent(fundsHolder, cz("5bcoin,5ccoin")),
				makeOutEvent(fundsHolder, "5bcoin,5ccoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "480acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "490acoin,500bcoin,495ccoin"),
				makeExpectedBalance(addr3, "525acoin,500bcoin,500ccoin"),
				makeExpectedBalance(fundsHolder, "5bcoin,5ccoin"),
			},
			expQ: []*QuarantinedCoins{makeQc("5bcoin,5ccoin", addr3, addr1, addr2)},
		},
		{
			name: "two outputs one that is quarantined",
			inputs: []types.Input{
				makeInput(addr3, "5bcoin,5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr1, "5bcoin"),
				makeOutput(addr2, "5ccoin"),
			},
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr1, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr3, cz("5bcoin,5ccoin")),
				makeInEvent(addr3),
				types.NewCoinReceivedEvent(fundsHolder, cz("5bcoin")),
				makeOutEvent(fundsHolder, "5bcoin"),
				types.NewCoinReceivedEvent(addr2, cz("5ccoin")),
				makeOutEvent(addr2, "5ccoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "480acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "490acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr3, "525acoin,495bcoin,495ccoin"),
				makeExpectedBalance(fundsHolder, "10bcoin,5ccoin"),
			},
			expQ: []*QuarantinedCoins{makeQc("5bcoin", addr1, addr3)},
		},
		{
			name: "two outputs both quarantined",
			inputs: []types.Input{
				makeInput(addr2, "11acoin,22ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr1, "11acoin"),
				makeOutput(addr3, "22ccoin"),
			},
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr1, true).
				WithIsQuarantinedAddrResponse(addr3, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr2, cz("11acoin,22ccoin")),
				makeInEvent(addr2),
				types.NewCoinReceivedEvent(fundsHolder, cz("11acoin")),
				makeOutEvent(fundsHolder, "11acoin"),
				types.NewCoinReceivedEvent(fundsHolder, cz("22ccoin")),
				makeOutEvent(fundsHolder, "22ccoin"),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "480acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "479acoin,500bcoin,478ccoin"),
				makeExpectedBalance(addr3, "525acoin,495bcoin,495ccoin"),
				makeExpectedBalance(fundsHolder, "11acoin,10bcoin,27ccoin"),
			},
			expQ: []*QuarantinedCoins{
				makeQc("11acoin", addr1, addr2),
				makeQc("22ccoin", addr3, addr2),
			},
		},
		{
			name:    "add quarantined coins returns error",
			inputs:  []types.Input{makeInput(addr1, "5acoin")},
			outputs: []types.Output{makeOutput(addr2, "5acoin")},
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr2, true).
				WithQueuedAddQuarantinedCoinsErrors(fmt.Errorf("this is a mocked error")),
			expInErr:  []string{"this is a mocked error"},
			expEvents: nil,
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "475acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "479acoin,500bcoin,478ccoin"),
				makeExpectedBalance(addr3, "525acoin,495bcoin,495ccoin"),
				makeExpectedBalance(fundsHolder, "11acoin,10bcoin,27ccoin"),
			},
			expQ: []*QuarantinedCoins{},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			defer app.BankKeeper.SetQuarantineKeeper(nil)
			bk := app.BankKeeper
			bk.SetQuarantineKeeper(tc.qk)
			em := sdk.NewEventManager()
			tctx := ctx.WithEventManager(em)
			err := bk.InputOutputCoins(tctx, tc.inputs, tc.outputs)
			if len(tc.expInErr) > 0 {
				if suite.Assert().Error(err, "InputOutputCoins error") {
					for _, exp := range tc.expInErr {
						suite.Assert().ErrorContains(err, exp, "InputOutputCoins error")
					}
				}
			} else {
				suite.Assert().NoError(err, "InputOutputCoins error")
			}

			if tc.expEvents != nil {
				events := em.Events()
				suite.Assert().Equal(tc.expEvents, events, "InputOutputCoins emitted events")
			}

			for i, expBal := range tc.expBals {
				actual := bk.GetAllBalances(ctx, expBal.addr)
				suite.Assert().Equal(expBal.balance.String(), actual.String(), "expected balance %d for %s", i, string(expBal.addr))
			}

			if tc.expQ != nil {
				qk := tc.qk.(*MockQuarantineKeeper)
				actual := qk.AddedQuarantinedCoins
				suite.Assert().Equal(tc.expQ, actual, "calls made to AddQuarantinedCoins")
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSendCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress("addr1_______________")
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress("addr2_______________")
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, balances))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	expected = sdk.NewCoins(newFooCoin(150), newBarCoin(75))
	suite.Require().Equal(expected, acc2Balances)

	// we sent all foo coins to acc2, so foo balance should be deleted for acc1 and bar should be still there
	var coins []sdk.Coin
	app.BankKeeper.IterateAccountBalances(ctx, addr1, func(c sdk.Coin) (stop bool) {
		coins = append(coins, c)
		return true
	})
	suite.Require().Len(coins, 1)
	suite.Require().Equal(newBarCoin(25), coins[0], "expected only bar coins in the account balance, got: %v", coins)
}

func (suite *IntegrationTestSuite) TestInputOutputCoinsWithSanction() {
	app, ctx := suite.app, suite.ctx

	// makeAndFundAccount makes and (if balance isn't zero) funds an account.
	makeAndFundAccount := func(i uint8, balance sdk.Coins) sdk.AccAddress {
		addr := sdk.AccAddress(fmt.Sprintf("addr%03d_____________", i))
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
		if !balance.IsZero() {
			suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balance), "funding account %d with %q", i, balance.String())
		}
		return addr
	}

	// cz converts the string to a Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// makeInput makes an Input.
	makeInput := func(addr sdk.AccAddress, coins string) types.Input {
		return types.Input{
			Address: addr.String(),
			Coins:   cz(coins),
		}
	}

	// makeOutput makes an Output.
	makeOutput := func(addr sdk.AccAddress, coins string) types.Output {
		return types.Output{
			Address: addr.String(),
			Coins:   cz(coins),
		}
	}

	type expectedBalance struct {
		addr    sdk.AccAddress
		balance sdk.Coins
	}

	makeExpectedBalance := func(addr sdk.AccAddress, coins string) *expectedBalance {
		return &expectedBalance{
			addr:    addr,
			balance: cz(coins),
		}
	}

	addr1 := makeAndFundAccount(1, cz("500acoin,500bcoin,500ccoin"))
	addr2 := makeAndFundAccount(2, cz("500acoin,500bcoin,500ccoin"))
	addr3 := makeAndFundAccount(3, cz("500acoin,500bcoin,500ccoin"))

	tests := []struct {
		name     string
		inputs   []types.Input
		outputs  []types.Output
		sk       *MockSanctionKeeper
		expInErr []string
		expCalls []sdk.AccAddress
		expBals  []*expectedBalance
	}{
		{
			name: "nil sanction keeper",
			inputs: []types.Input{
				makeInput(addr1, "5acoin"),
				makeInput(addr2, "5acoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10acoin"),
			},
			sk: nil,
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,500bcoin,500ccoin"),
			},
		},
		{
			name: "no sanctioned addresses",
			inputs: []types.Input{
				makeInput(addr1, "5bcoin"),
				makeInput(addr2, "5bcoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10bcoin"),
			},
			sk:       NewMockSanctionKeeper(),
			expCalls: []sdk.AccAddress{addr1, addr2},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,510bcoin,500ccoin"),
			},
		},
		{
			name: "first input addr sanctioned",
			inputs: []types.Input{
				makeInput(addr1, "5ccoin"),
				makeInput(addr2, "5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10ccoin"),
			},
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr1),
			expInErr: []string{addr1.String(), "account is sanctioned"},
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr2, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,510bcoin,500ccoin"),
			},
		},
		{
			name: "second input addr sanctioned",
			inputs: []types.Input{
				makeInput(addr1, "5ccoin"),
				makeInput(addr2, "5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10ccoin"),
			},
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr2),
			expInErr: []string{addr2.String(), "account is sanctioned"},
			expCalls: []sdk.AccAddress{addr1, addr2},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,495ccoin"),
				makeExpectedBalance(addr2, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,510bcoin,500ccoin"),
			},
		},
		{
			name: "both input addrs sanctioned",
			inputs: []types.Input{
				makeInput(addr1, "5ccoin"),
				makeInput(addr2, "5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10ccoin"),
			},
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr1, addr2),
			expInErr: []string{addr1.String(), "account is sanctioned"},
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,495ccoin"),
				makeExpectedBalance(addr2, "495acoin,495bcoin,500ccoin"),
				makeExpectedBalance(addr3, "510acoin,510bcoin,500ccoin"),
			},
		},
		{
			name: "output addr sanctioned",
			inputs: []types.Input{
				makeInput(addr1, "5ccoin"),
				makeInput(addr2, "5ccoin"),
			},
			outputs: []types.Output{
				makeOutput(addr3, "10ccoin"),
			},
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr3),
			expCalls: []sdk.AccAddress{addr1, addr2},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,490ccoin"),
				makeExpectedBalance(addr2, "495acoin,495bcoin,495ccoin"),
				makeExpectedBalance(addr3, "510acoin,510bcoin,510ccoin"),
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			defer app.BankKeeper.SetSanctionKeeper(nil)
			bk := app.BankKeeper
			bk.SetSanctionKeeper(nil)
			if tc.sk != nil {
				bk.SetSanctionKeeper(tc.sk)
			}
			var err error
			testFunc := func() {
				err = bk.InputOutputCoins(ctx, tc.inputs, tc.outputs)
			}
			suite.Require().NotPanics(testFunc, "InputOutputCoins")

			if len(tc.expInErr) > 0 {
				if suite.Assert().Error(err, "InputOutputCoins error") {
					for _, exp := range tc.expInErr {
						suite.Assert().ErrorContains(err, exp, "InputOutputCoins error")
					}
				}
			} else {
				suite.Assert().NoError(err, "InputOutputCoins error")
			}

			if tc.sk != nil {
				calls := tc.sk.IsSanctionedAddrCalls
				suite.Assert().Equal(tc.expCalls, calls, "addresses provided to IsSanctionedAddr")
			}

			for i, expBal := range tc.expBals {
				actual := bk.GetAllBalances(ctx, expBal.addr)
				suite.Assert().Equal(expBal.balance.String(), actual.String(), "expected balance %d for %s", i, string(expBal.addr))
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSendCoinsWithQuarantine() {
	app, ctx := suite.app, suite.ctx

	// makeAndFundAccount makes and (if balance isn't zero) funds an account.
	makeAndFundAccount := func(i uint8, balance sdk.Coins) sdk.AccAddress {
		addr := sdk.AccAddress(fmt.Sprintf("addr%03d_____________", i))
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
		if !balance.IsZero() {
			suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balance), "funding account %d with %q", i, balance.String())
		}
		return addr
	}

	// cz converts the string to a Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// makeInEvent makes an event expected from an input.
	makeInEvent := func(addr sdk.AccAddress) sdk.Event {
		return sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(types.AttributeKeySender, addr.String()),
		)
	}

	// makeTransferEvent makes an event expeced for a transfer.
	makeTransferEvent := func(to, from sdk.AccAddress, amt string) sdk.Event {
		return sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.AttributeKeyRecipient, to.String()),
			sdk.NewAttribute(types.AttributeKeySender, from.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, amt),
		)
	}

	type expectedBalance struct {
		addr    sdk.AccAddress
		balance sdk.Coins
	}

	makeExpectedBalance := func(addr sdk.AccAddress, coins string) *expectedBalance {
		return &expectedBalance{
			addr:    addr,
			balance: cz(coins),
		}
	}

	makeQc := func(coins string, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) *QuarantinedCoins {
		return &QuarantinedCoins{
			coins:     cz(coins),
			toAddr:    toAddr,
			fromAddrs: fromAddrs,
		}
	}

	fundsHolder := sdk.AccAddress("holdsquarantinefunds")
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, fundsHolder))

	addr1 := makeAndFundAccount(1, cz("500acoin,500bcoin,500ccoin"))
	addr2 := makeAndFundAccount(2, cz("500acoin,500bcoin,500ccoin"))

	tests := []struct {
		name      string
		fromAddr  sdk.AccAddress
		toAddr    sdk.AccAddress
		amt       sdk.Coins
		qk        types.QuarantineKeeper
		expInErr  []string
		expEvents sdk.Events
		expBals   []*expectedBalance
		expQ      []*QuarantinedCoins
	}{
		{
			name:     "nil quarantine keeper",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5acoin"),
			qk:       nil,
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("5acoin")),
				types.NewCoinReceivedEvent(addr2, cz("5acoin")),
				makeTransferEvent(addr2, addr1, "5acoin"),
				makeInEvent(addr1),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,500ccoin"),
				makeExpectedBalance(fundsHolder, ""),
			},
			expQ: nil,
		},
		{
			name:     "no funds holder",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("10acoin"),
			qk: NewMockMockQuarantineKeeper(nil).
				WithIsQuarantinedAddrResponse(addr2, true),
			expInErr:  []string{"no quarantine holder account defined", "unknown address"},
			expEvents: sdk.Events{},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,500ccoin"),
				makeExpectedBalance(fundsHolder, ""),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name:     "with quarantine keeper but not quarantined",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5ccoin"),
			qk:       NewMockMockQuarantineKeeper(fundsHolder),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("5ccoin")),
				types.NewCoinReceivedEvent(addr2, cz("5ccoin")),
				makeTransferEvent(addr2, addr1, "5ccoin"),
				makeInEvent(addr1),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,505ccoin"),
				makeExpectedBalance(fundsHolder, ""),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name:     "quarantined with auto-accept",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5bcoin"),
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr2, true).
				WithIsAutoAcceptResponse(addr2, []sdk.AccAddress{addr1}, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("5bcoin")),
				types.NewCoinReceivedEvent(addr2, cz("5bcoin")),
				makeTransferEvent(addr2, addr1, "5bcoin"),
				makeInEvent(addr1),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,495bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,505bcoin,505ccoin"),
				makeExpectedBalance(fundsHolder, ""),
			},
			expQ: []*QuarantinedCoins{},
		},
		{
			name:     "quarantined",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("7acoin"),
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr2, true),
			expInErr: nil,
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("7acoin")),
				types.NewCoinReceivedEvent(fundsHolder, cz("7acoin")),
				makeTransferEvent(fundsHolder, addr1, "7acoin"),
				makeInEvent(addr1),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "488acoin,495bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,505bcoin,505ccoin"),
				makeExpectedBalance(fundsHolder, "7acoin"),
			},
			expQ: []*QuarantinedCoins{
				makeQc("7acoin", addr2, addr1),
			},
		},
		{
			name:     "add quarantined coins returns error",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("8bcoin"),
			qk: NewMockMockQuarantineKeeper(fundsHolder).
				WithIsQuarantinedAddrResponse(addr2, true).
				WithQueuedAddQuarantinedCoinsErrors(fmt.Errorf("this is a mocked test error")),
			expInErr: []string{"this is a mocked test error"},
			expEvents: sdk.Events{
				types.NewCoinSpentEvent(addr1, cz("8bcoin")),
				types.NewCoinReceivedEvent(fundsHolder, cz("8bcoin")),
				makeTransferEvent(fundsHolder, addr1, "8bcoin"),
				makeInEvent(addr1),
			},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "488acoin,487bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,505bcoin,505ccoin"),
				makeExpectedBalance(fundsHolder, "7acoin,8bcoin"),
			},
			expQ: []*QuarantinedCoins{},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			defer app.BankKeeper.SetQuarantineKeeper(nil)
			bk := app.BankKeeper
			bk.SetQuarantineKeeper(tc.qk)
			em := sdk.NewEventManager()
			tctx := ctx.WithEventManager(em)
			err := bk.SendCoins(tctx, tc.fromAddr, tc.toAddr, tc.amt)
			if len(tc.expInErr) > 0 {
				if suite.Assert().Error(err, "SendCoins error") {
					for _, exp := range tc.expInErr {
						suite.Assert().ErrorContains(err, exp, "SendCoins error")
					}
				}
			} else {
				suite.Assert().NoError(err, "SendCoins error")
			}

			if tc.expEvents != nil {
				events := em.Events()
				suite.Assert().Equal(tc.expEvents, events, "SendCoins emitted events")
			}

			for i, expBal := range tc.expBals {
				actual := bk.GetAllBalances(ctx, expBal.addr)
				suite.Assert().Equal(expBal.balance.String(), actual.String(), "expected balance %d for %s", i, string(expBal.addr))
			}

			if tc.expQ != nil {
				qk := tc.qk.(*MockQuarantineKeeper)
				actual := qk.AddedQuarantinedCoins
				suite.Assert().Equal(tc.expQ, actual, "calls made to AddQuarantinedCoins")
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSendCoinsWithSanction() {
	app, ctx := suite.app, suite.ctx

	// makeAndFundAccount makes and (if balance isn't zero) funds an account.
	makeAndFundAccount := func(i uint8, balance sdk.Coins) sdk.AccAddress {
		addr := sdk.AccAddress(fmt.Sprintf("addr%03d_____________", i))
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
		app.AccountKeeper.SetAccount(ctx, acc)
		if !balance.IsZero() {
			suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balance), "funding account %d with %q", i, balance.String())
		}
		return addr
	}

	// cz converts the string to a Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	type expectedBalance struct {
		addr    sdk.AccAddress
		balance sdk.Coins
	}

	makeExpectedBalance := func(addr sdk.AccAddress, coins string) *expectedBalance {
		return &expectedBalance{
			addr:    addr,
			balance: cz(coins),
		}
	}

	addr1 := makeAndFundAccount(1, cz("500acoin,500bcoin,500ccoin"))
	addr2 := makeAndFundAccount(2, cz("500acoin,500bcoin,500ccoin"))

	tests := []struct {
		name     string
		fromAddr sdk.AccAddress
		toAddr   sdk.AccAddress
		amt      sdk.Coins
		sk       *MockSanctionKeeper
		expInErr []string
		expCalls []sdk.AccAddress
		expBals  []*expectedBalance
	}{
		{
			name:     "neither address sanctioned",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5acoin"),
			sk:       NewMockSanctionKeeper(),
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,500ccoin"),
			},
		},
		{
			name:     "from address sanctioned",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5bcoin"),
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr1),
			expInErr: []string{addr1.String(), "account is sanctioned"},
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,500ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,500ccoin"),
			},
		},
		{
			name:     "to address sanctioned",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5ccoin"),
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr2),
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,505ccoin"),
			},
		},
		{
			name:     "both addresses sanctioned",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("5bcoin"),
			sk:       NewMockSanctionKeeper().WithSanctionedAddrs(addr1, addr2),
			expInErr: []string{addr1.String(), "account is sanctioned"},
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "495acoin,500bcoin,495ccoin"),
				makeExpectedBalance(addr2, "505acoin,500bcoin,505ccoin"),
			},
		},
		{
			name:     "nil sanction keeper",
			fromAddr: addr1,
			toAddr:   addr2,
			amt:      cz("3acoin,4bcoin,5ccoin"),
			sk:       nil,
			expCalls: []sdk.AccAddress{addr1},
			expBals: []*expectedBalance{
				makeExpectedBalance(addr1, "492acoin,496bcoin,490ccoin"),
				makeExpectedBalance(addr2, "508acoin,504bcoin,510ccoin"),
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			defer app.BankKeeper.SetSanctionKeeper(nil)
			bk := app.BankKeeper
			bk.SetSanctionKeeper(nil)
			if tc.sk != nil {
				bk.SetSanctionKeeper(tc.sk)
			}
			var err error
			testFunc := func() {
				err = bk.SendCoins(ctx, tc.fromAddr, tc.toAddr, tc.amt)
			}
			suite.Require().NotPanics(testFunc, "SendCoins")

			if len(tc.expInErr) > 0 {
				if suite.Assert().Error(err, "SendCoins error") {
					for _, exp := range tc.expInErr {
						suite.Assert().ErrorContains(err, exp, "SendCoins error")
					}
				}
			} else {
				suite.Assert().NoError(err, "SendCoins error")
			}

			if tc.sk != nil {
				calls := tc.sk.IsSanctionedAddrCalls
				suite.Assert().Equal(tc.expCalls, calls, "addresses provided to IsSanctionedAddr")
			}

			for i, expBal := range tc.expBals {
				actual := bk.GetAllBalances(ctx, expBal.addr)
				suite.Assert().Equal(expBal.balance.String(), actual.String(), "expected balance %d for %s", i, string(expBal.addr))
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestValidateBalance() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr1))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.ValidateBalance(ctx, addr1))

	bacc := authtypes.NewBaseAccountWithAddress(addr2)
	vacc := vesting.NewContinuousVestingAccount(bacc, balances.Add(balances...), now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, balances))
	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestSendEnabled() {
	app, ctx := suite.app, suite.ctx
	enabled := true
	params := types.DefaultParams()
	suite.Require().Equal(enabled, params.DefaultSendEnabled)

	app.BankKeeper.SetParams(ctx, params)

	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())
	fooCoin := sdk.NewCoin("foocoin", sdk.OneInt())
	barCoin := sdk.NewCoin("barcoin", sdk.OneInt())

	// assert with default (all denom) send enabled both Bar and Bond Denom are enabled
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Both coins should be send enabled.
	err := app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().NoError(err)

	// Set default send_enabled to !enabled, add a foodenom that overrides default as enabled
	params.DefaultSendEnabled = !enabled
	app.BankKeeper.SetParams(ctx, params)
	app.BankKeeper.SetSendEnabled(ctx, fooCoin.Denom, enabled)

	// Expect our specific override to be enabled, others to be !enabled.
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, fooCoin))
	suite.Require().Equal(!enabled, app.BankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(!enabled, app.BankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Foo coin should be send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin)
	suite.Require().NoError(err)

	// Expect an error when one coin is not send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().Error(err)

	// Expect an error when all coins are not send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, bondCoin, barCoin)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestHasBalance() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(99)))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balances))
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(101)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(100)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(1)))
}

func (suite *IntegrationTestSuite) TestMsgSendEvents() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, newCoins))

	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins))
	event1 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr2.String())},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())},
	)

	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)

	// events are shifted due to the funding account events
	events := ctx.EventManager().ABCIEvents()
	suite.Require().Equal(10, len(events))
	suite.Require().Equal(abci.Event(event1), events[8])
	suite.Require().Equal(abci.Event(event2), events[9])
}

func (suite *IntegrationTestSuite) TestMsgMultiSendEvents() {
	app, ctx := suite.app, suite.ctx

	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, acc2)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	newCoins2 := sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))
	inputs := []types.Input{
		{Address: addr.String(), Coins: newCoins},
		{Address: addr2.String(), Coins: newCoins2},
	}
	outputs := []types.Output{
		{Address: addr3.String(), Coins: newCoins},
		{Address: addr4.String(), Coins: newCoins2},
	}

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events := ctx.EventManager().ABCIEvents()
	suite.Require().Equal(0, len(events))

	// Set addr's coins but not addr2's coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))))
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events = ctx.EventManager().ABCIEvents()
	suite.Require().Equal(8, len(events)) // 7 events because account funding causes extra minting + coin_spent + coin_recv events

	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)
	suite.Require().Equal(abci.Event(event1), events[7])

	// Set addr's coins and addr2's coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))))
	newCoins2 = sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))

	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events = ctx.EventManager().ABCIEvents()
	suite.Require().Equal(28, len(events)) // 25 due to account funding + coin_spent + coin_recv events

	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr2.String())},
	)
	event3 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event3.Attributes = append(
		event3.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr3.String())},
	)
	event3.Attributes = append(
		event3.Attributes,
		abci.EventAttribute{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())})
	event4 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event4.Attributes = append(
		event4.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr4.String())},
	)
	event4.Attributes = append(
		event4.Attributes,
		abci.EventAttribute{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins2.String())},
	)
	// events are shifted due to the funding account events
	suite.Require().Equal(abci.Event(event1), events[21])
	suite.Require().Equal(abci.Event(event2), events[23])
	suite.Require().Equal(abci.Event(event3), events[25])
	suite.Require().Equal(abci.Event(event4), events[27])
}

func (suite *IntegrationTestSuite) TestSpendableCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, macc)
	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.SpendableCoins(ctx, addr2))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.SpendableCoins(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, sendCoins))
	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, sendCoins))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestDelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addr1))

	// require that delegated vesting amount is equal to what was delegated with DelegateCoins
	acc = app.AccountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Equal(delCoins, vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestDelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}}
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, invalidCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, origCoins.Add(origCoins...)))
}

func (suite *IntegrationTestSuite) TestUndelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a non-vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))

	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require that delegated vesting amount is completely empty, since they were completely undelegated
	acc = app.AccountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Empty(vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestUndelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
}

func (suite *IntegrationTestSuite) TestSetDenomMetaData() {
	app, ctx := suite.app, suite.ctx

	metadata := suite.getTestMetadata()

	for i := range []int{1, 2} {
		app.BankKeeper.SetDenomMetaData(ctx, metadata[i])
	}

	actualMetadata, found := app.BankKeeper.GetDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	found = app.BankKeeper.HasDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	suite.Require().Equal(metadata[1].GetBase(), actualMetadata.GetBase())
	suite.Require().Equal(metadata[1].GetDisplay(), actualMetadata.GetDisplay())
	suite.Require().Equal(metadata[1].GetDescription(), actualMetadata.GetDescription())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetDenom(), actualMetadata.GetDenomUnits()[1].GetDenom())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetExponent(), actualMetadata.GetDenomUnits()[1].GetExponent())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetAliases(), actualMetadata.GetDenomUnits()[1].GetAliases())
}

func (suite *IntegrationTestSuite) TestIterateAllDenomMetaData() {
	app, ctx := suite.app, suite.ctx

	expectedMetadata := suite.getTestMetadata()
	// set metadata
	for i := range []int{1, 2} {
		app.BankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
	}
	// retrieve metadata
	actualMetadata := make([]types.Metadata, 0)
	app.BankKeeper.IterateAllDenomMetaData(ctx, func(metadata types.Metadata) bool {
		actualMetadata = append(actualMetadata, metadata)
		return false
	})
	// execute checks
	for i := range []int{1, 2} {
		suite.Require().Equal(expectedMetadata[i].GetBase(), actualMetadata[i].GetBase())
		suite.Require().Equal(expectedMetadata[i].GetDisplay(), actualMetadata[i].GetDisplay())
		suite.Require().Equal(expectedMetadata[i].GetDescription(), actualMetadata[i].GetDescription())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetDenom(), actualMetadata[i].GetDenomUnits()[1].GetDenom())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetExponent(), actualMetadata[i].GetDenomUnits()[1].GetExponent())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetAliases(), actualMetadata[i].GetDenomUnits()[1].GetAliases())
	}
}

func (suite *IntegrationTestSuite) TestBalanceTrackingEvents() {
	// replace account keeper and bank keeper otherwise the account keeper won't be aware of the
	// existence of the new module account because GetModuleAccount checks for the existence via
	// permissions map and not via state... weird
	maccPerms := simapp.GetMaccPerms()
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	suite.app.AccountKeeper = authkeeper.NewAccountKeeper(
		suite.app.AppCodec(), suite.app.GetKey(authtypes.StoreKey), suite.app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)

	suite.app.BankKeeper = keeper.NewBaseKeeper(suite.app.AppCodec(), suite.app.GetKey(types.StoreKey),
		suite.app.AccountKeeper, suite.app.GetSubspace(types.ModuleName), nil,
	)

	// set account with multiple permissions
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)
	// mint coins
	suite.Require().NoError(
		suite.app.BankKeeper.MintCoins(
			suite.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(100000)))),
	)
	// send coins to address
	addr1 := sdk.AccAddress("addr1_______________")
	suite.Require().NoError(
		suite.app.BankKeeper.SendCoinsFromModuleToAccount(
			suite.ctx,
			multiPermAcc.Name,
			addr1,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(50000))),
		),
	)

	// burn coins from module account
	suite.Require().NoError(
		suite.app.BankKeeper.BurnCoins(
			suite.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewInt64Coin("utxo", 1000)),
		),
	)

	// process balances and supply from events
	supply := sdk.NewCoins()

	balances := make(map[string]sdk.Coins)

	for _, e := range suite.ctx.EventManager().ABCIEvents() {
		switch e.Type {
		case types.EventTypeCoinBurn:
			burnedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			supply = supply.Sub(burnedCoins...)

		case types.EventTypeCoinMint:
			mintedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			supply = supply.Add(mintedCoins...)

		case types.EventTypeCoinSpent:
			coinsSpent, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			spender, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			suite.Require().NoError(err)
			balances[spender.String()] = balances[spender.String()].Sub(coinsSpent...)

		case types.EventTypeCoinReceived:
			coinsRecv, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			receiver, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			suite.Require().NoError(err)
			balances[receiver.String()] = balances[receiver.String()].Add(coinsRecv...)
		}
	}

	// check balance and supply tracking
	suite.Require().True(suite.app.BankKeeper.HasSupply(suite.ctx, "utxo"))
	savedSupply := suite.app.BankKeeper.GetSupply(suite.ctx, "utxo")
	utxoSupply := savedSupply
	suite.Require().Equal(utxoSupply.Amount, supply.AmountOf("utxo"))
	// iterate accounts and check balances
	suite.app.BankKeeper.IterateAllBalances(suite.ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
		// if it's not utxo coin then skip
		if coin.Denom != "utxo" {
			return false
		}

		balance, exists := balances[address.String()]
		suite.Require().True(exists)

		expectedUtxo := sdk.NewCoin("utxo", balance.AmountOf(coin.Denom))
		suite.Require().Equal(expectedUtxo.String(), coin.String())
		return false
	})
}

func (suite *IntegrationTestSuite) getTestMetadata() []types.Metadata {
	return []types.Metadata{
		{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{"uatom", uint32(0), []string{"microatom"}},
				{"matom", uint32(3), []string{"milliatom"}},
				{"atom", uint32(6), nil},
			},
			Base:    "uatom",
			Display: "atom",
		},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*types.DenomUnit{
				{"1token", uint32(5), []string{"decitoken"}},
				{"2token", uint32(4), []string{"centitoken"}},
				{"3token", uint32(7), []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
	}
}

func (suite *IntegrationTestSuite) TestMintCoinRestrictions() {
	type BankMintingRestrictionFn func(ctx sdk.Context, coins sdk.Coins) error

	maccPerms := simapp.GetMaccPerms()
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	suite.app.AccountKeeper = authkeeper.NewAccountKeeper(
		suite.app.AppCodec(), suite.app.GetKey(authtypes.StoreKey), suite.app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)

	type testCase struct {
		coinsToTry sdk.Coin
		expectPass bool
	}

	tests := []struct {
		name          string
		restrictionFn BankMintingRestrictionFn
		testCases     []testCase
	}{
		{
			"restriction",
			func(ctx sdk.Context, coins sdk.Coins) error {
				for _, coin := range coins {
					if coin.Denom != fooDenom {
						return fmt.Errorf("Module %s only has perms for minting %s coins, tried minting %s coins", types.ModuleName, fooDenom, coin.Denom)
					}
				}
				return nil
			},
			[]testCase{
				{
					coinsToTry: newFooCoin(100),
					expectPass: true,
				},
				{
					coinsToTry: newBarCoin(100),
					expectPass: false,
				},
			},
		},
	}

	for _, test := range tests {
		suite.app.BankKeeper = keeper.NewBaseKeeper(suite.app.AppCodec(), suite.app.GetKey(types.StoreKey),
			suite.app.AccountKeeper, suite.app.GetSubspace(types.ModuleName), nil,
		).WithMintCoinsRestriction(keeper.MintingRestrictionFn(test.restrictionFn))
		for _, testCase := range test.testCases {
			if testCase.expectPass {
				suite.Require().NoError(
					suite.app.BankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
				)
			} else {
				suite.Require().Error(
					suite.app.BankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
				)
			}
		}
	}
}

func (suite *IntegrationTestSuite) TestIsSendEnabledDenom() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	defaultCoin := "defaultCoin"
	enabledCoin := "enabledCoin"
	disabledCoin := "disabledCoin"
	bankKeeper.DeleteSendEnabled(ctx, defaultCoin)
	bankKeeper.SetSendEnabled(ctx, enabledCoin, true)
	bankKeeper.SetSendEnabled(ctx, disabledCoin, false)

	tests := []struct {
		denom  string
		exp    bool
		expDef bool
	}{
		{
			denom:  "defaultCoin",
			expDef: true,
		},
		{
			denom: enabledCoin,
			exp:   true,
		},
		{
			denom: disabledCoin,
			exp:   false,
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.denom, def), func(t *testing.T) {
				actual := suite.app.BankKeeper.IsSendEnabledDenom(suite.ctx, tc.denom)
				exp := tc.exp
				if tc.expDef {
					exp = def
				}
				assert.Equal(t, exp, actual)
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestGetSendEnabledEntry() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	bankKeeper.SetAllSendEnabled(ctx, []*types.SendEnabled{
		{Denom: "gettruecoin", Enabled: true},
		{Denom: "getfalsecoin", Enabled: false},
	})

	tests := []struct {
		denom string
		expSE types.SendEnabled
		expF  bool
	}{
		{
			denom: "missing",
			expSE: types.SendEnabled{},
			expF:  false,
		},
		{
			denom: "gettruecoin",
			expSE: types.SendEnabled{Denom: "gettruecoin", Enabled: true},
			expF:  true,
		},
		{
			denom: "getfalsecoin",
			expSE: types.SendEnabled{Denom: "getfalsecoin", Enabled: false},
			expF:  true,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.denom, func(t *testing.T) {
			actualSE, actualF := bankKeeper.GetSendEnabledEntry(ctx, tc.denom)
			assert.Equal(t, tc.expF, actualF, "found")
			assert.Equal(t, tc.expSE, actualSE, "SendEnabled")
		})
	}
}

func (suite *IntegrationTestSuite) TestSetSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	tests := []struct {
		name  string
		denom string
		value bool
	}{
		{
			name:  "very short denom true",
			denom: "f",
			value: true,
		},
		{
			name:  "very short denom false",
			denom: "f",
			value: true,
		},
		{
			name:  "falseFirstCoin false",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "falseFirstCoin true",
			denom: "falseFirstCoin",
			value: true,
		},
		{
			name:  "falseFirstCoin true again",
			denom: "falseFirstCoin",
			value: true,
		},
		{
			name:  "trueFirstCoin true",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "trueFirstCoin false",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "trueFirstCoin false again",
			denom: "falseFirstCoin",
			value: false,
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetSendEnabled(ctx, tc.denom, tc.value)
				actual := bankKeeper.IsSendEnabledDenom(ctx, tc.denom)
				assert.Equal(t, tc.value, actual)
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestSetAllSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	tests := []struct {
		name         string
		sendEnableds []*types.SendEnabled
	}{
		{
			name:         "nil",
			sendEnableds: nil,
		},
		{
			name:         "empty",
			sendEnableds: []*types.SendEnabled{},
		},
		{
			name: "one true",
			sendEnableds: []*types.SendEnabled{
				{"aonecoin", true},
			},
		},
		{
			name: "one false",
			sendEnableds: []*types.SendEnabled{
				{"bonecoin", false},
			},
		},
		{
			name: "two true",
			sendEnableds: []*types.SendEnabled{
				{"conecoin", true},
				{"ctwocoin", true},
			},
		},
		{
			name: "two true false",
			sendEnableds: []*types.SendEnabled{
				{"donecoin", true},
				{"dtwocoin", false},
			},
		},
		{
			name: "two false true",
			sendEnableds: []*types.SendEnabled{
				{"eonecoin", false},
				{"etwocoin", true},
			},
		},
		{
			name: "two false",
			sendEnableds: []*types.SendEnabled{
				{"fonecoin", false},
				{"ftwocoin", false},
			},
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetAllSendEnabled(ctx, tc.sendEnableds)
				for _, se := range tc.sendEnableds {
					actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
					assert.Equal(t, se.Enabled, actual, se.Denom)
				}
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestDeleteSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		suite.T().Run(fmt.Sprintf("default %t", def), func(t *testing.T) {
			denom := fmt.Sprintf("somerand%tcoin", !def)
			bankKeeper.SetSendEnabled(ctx, denom, !def)
			require.Equal(t, !def, bankKeeper.IsSendEnabledDenom(ctx, denom))
			bankKeeper.DeleteSendEnabled(ctx, denom)
			require.Equal(t, def, bankKeeper.IsSendEnabledDenom(ctx, denom))
		})
	}
}

func (suite *IntegrationTestSuite) TestIterateSendEnabledEntries() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	suite.T().Run("no entries to iterate", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})

	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	denoms := make([]string, len(alpha)*2)
	for i, l := range alpha {
		denoms[i*2] = fmt.Sprintf("%sitercointrue", l)
		denoms[i*2+1] = fmt.Sprintf("%sitercoinfalse", l)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2], true)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2+1], false)
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		var seen []string
		suite.T().Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
			bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
				seen = append(seen, denom)
				exp := true
				if strings.HasSuffix(denom, "false") {
					exp = false
				}
				assert.Equal(t, exp, sendEnabled, denom)
				return false
			})
		})
		suite.T().Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			assert.ElementsMatch(t, denoms, seen)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	suite.T().Run("no entries to iterate again after deleting all of them", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})
}

func (suite *IntegrationTestSuite) TestGetAllSendEnabledEntries() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	suite.T().Run("no entries", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Len(t, actual, 0)
	})

	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	denoms := make([]string, len(alpha)*2)
	for i, l := range alpha {
		denoms[i*2] = fmt.Sprintf("%sitercointrue", l)
		denoms[i*2+1] = fmt.Sprintf("%sitercoinfalse", l)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2], true)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2+1], false)
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		var seen []string
		suite.T().Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
			actual := bankKeeper.GetAllSendEnabledEntries(ctx)
			for _, se := range actual {
				seen = append(seen, se.Denom)
				exp := true
				if strings.HasSuffix(se.Denom, "false") {
					exp = false
				}
				assert.Equal(t, exp, se.Enabled, se.Denom)
			}
		})
		suite.T().Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			assert.ElementsMatch(t, denoms, seen)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	suite.T().Run("no entries again after deleting all of them", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Len(t, actual, 0)
	})
}

func (suite *IntegrationTestSuite) TestMigrator_Migrate3to4() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		suite.T().Run(fmt.Sprintf("default %t does not change", def), func(t *testing.T) {
			kp := bankKeeper.(*keeper.BaseKeeper)
			migrator := keeper.NewMigrator(*kp)
			require.NoError(t, migrator.Migrate3to4(ctx))
			actual := bankKeeper.GetParams(ctx)
			assert.Equal(t, params.DefaultSendEnabled, actual.DefaultSendEnabled)
		})
	}

	for _, def := range []bool{true, false} {
		params := types.Params{
			SendEnabled: []*types.SendEnabled{
				{fmt.Sprintf("truecoin%t", def), true},
				{fmt.Sprintf("falsecoin%t", def), false},
			},
		}
		bankKeeper.SetParams(ctx, params)
		suite.T().Run(fmt.Sprintf("default %t send enabled info moved to store", def), func(t *testing.T) {
			kp := bankKeeper.(*keeper.BaseKeeper)
			migrator := keeper.NewMigrator(*kp)
			require.NoError(t, migrator.Migrate3to4(ctx))
			newParams := bankKeeper.GetParams(ctx)
			assert.Len(t, newParams.SendEnabled, 0)
			for _, se := range params.SendEnabled {
				actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
				assert.Equal(t, se.Enabled, actual, se.Denom)
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSetParams() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper
	params := types.NewParams(true)
	params.SendEnabled = []*types.SendEnabled{
		{"paramscointrue", true},
		{"paramscoinfalse", false},
	}
	bankKeeper.SetParams(ctx, params)

	suite.Run("stored params are as expected", func() {
		actual := bankKeeper.GetParams(ctx)
		suite.Assert().True(actual.DefaultSendEnabled, "DefaultSendEnabled")
		suite.Assert().Len(actual.SendEnabled, 0, "SendEnabled")
	})

	suite.Run("send enabled params converted to store", func() {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		if suite.Assert().Len(actual, 2) {
			suite.Equal("paramscoinfalse", actual[0].Denom, "actual[0].Denom")
			suite.False(actual[0].Enabled, "actual[0].Enabled")
			suite.Equal("paramscointrue", actual[1].Denom, "actual[1].Denom")
			suite.True(actual[1].Enabled, "actual[1].Enabled")
		}
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
