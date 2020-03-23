package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmkv "github.com/tendermint/tendermint/libs/kv"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	fooDenom = "foo"
	barDenom = "bar"
)

func newFooCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(fooDenom, amt)
}

func newBarCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(barDenom, amt)
}

type IntegrationTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	app.AccountKeeper.SetParams(ctx, auth.DefaultParams())
	app.BankKeeper.SetSendEnabled(ctx, true)

	suite.app = app
	suite.ctx = ctx
}

func (suite *IntegrationTestSuite) TestSendCoinsNewAccount() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2"))

	suite.Require().Nil(app.AccountKeeper.GetAccount(ctx, addr2))
	suite.Require().Empty(app.BankKeeper.GetAllBalances(ctx, addr2))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(sendAmt, acc2Balances)
	suite.Require().NotNil(app.AccountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(90), newBarCoin(30))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2"))
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)

	addr3 := sdk.AccAddress([]byte("addr3"))
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	app.AccountKeeper.SetAccount(ctx, acc3)

	inputs := []types.Input{
		{Address: addr1, Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr1, Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}
	outputs := []types.Output{
		{Address: addr2, Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr3, Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, []types.Output{}))
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, balances))

	insufficientInputs := []types.Input{
		{Address: addr1, Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr1, Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
	}
	insufficientOutputs := []types.Output{
		{Address: addr2, Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr3, Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
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

func (suite *IntegrationTestSuite) TestSendCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2"))
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, balances))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	expected = sdk.NewCoins(newFooCoin(150), newBarCoin(75))
	suite.Require().Equal(expected, acc2Balances)
}

func (suite *IntegrationTestSuite) TestValidateBalance() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr1))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.ValidateBalance(ctx, addr1))

	bacc := auth.NewBaseAccountWithAddress(addr2)
	vacc := vesting.NewContinuousVestingAccount(bacc, balances.Add(balances...), now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, balances))
	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestBalance() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1"))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	suite.Require().Equal(sdk.NewCoin(fooDenom, sdk.ZeroInt()), app.BankKeeper.GetBalance(ctx, addr, fooDenom))
	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr, balances))

	suite.Require().Equal(balances.AmountOf(fooDenom), app.BankKeeper.GetBalance(ctx, addr, fooDenom).Amount)
	suite.Require().Equal(balances, app.BankKeeper.GetAllBalances(ctx, addr))

	newFooBalance := newFooCoin(99)
	suite.Require().NoError(app.BankKeeper.SetBalance(ctx, addr, newFooBalance))
	suite.Require().Equal(newFooBalance, app.BankKeeper.GetBalance(ctx, addr, fooDenom))

	balances = sdk.NewCoins(newBarCoin(500))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr, balances))
	suite.Require().Equal(sdk.NewCoin(fooDenom, sdk.ZeroInt()), app.BankKeeper.GetBalance(ctx, addr, fooDenom))
	suite.Require().Equal(balances.AmountOf(barDenom), app.BankKeeper.GetBalance(ctx, addr, barDenom).Amount)
	suite.Require().Equal(balances, app.BankKeeper.GetAllBalances(ctx, addr))

	invalidBalance := sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}
	suite.Require().Error(app.BankKeeper.SetBalance(ctx, addr, invalidBalance))
}

func (suite *IntegrationTestSuite) TestSendEnabled() {
	app, ctx := suite.app, suite.ctx
	enabled := false
	app.BankKeeper.SetSendEnabled(ctx, enabled)
	suite.Require().Equal(enabled, app.BankKeeper.GetSendEnabled(ctx))
}

func (suite *IntegrationTestSuite) TestHasBalance() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1"))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(99)))

	app.BankKeeper.SetBalances(ctx, addr, balances)
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(101)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(100)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(1)))
}

func (suite *IntegrationTestSuite) TestMsgSendEvents() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins))

	events := ctx.EventManager().Events()
	suite.Require().Equal(2, len(events))

	event1 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event1.Attributes = append(
		event1.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr2.String())},
	)
	event1.Attributes = append(
		event1.Attributes,
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())},
	)
	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event2.Attributes = append(
		event2.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)

	suite.Require().Equal(event1, events[0])
	suite.Require().Equal(event2, events[1])

	app.BankKeeper.SetBalances(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50)))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins))

	events = ctx.EventManager().Events()
	suite.Require().Equal(4, len(events))
	suite.Require().Equal(event1, events[2])
	suite.Require().Equal(event2, events[3])
}

func (suite *IntegrationTestSuite) TestMsgMultiSendEvents() {
	app, ctx := suite.app, suite.ctx

	app.BankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	addr4 := sdk.AccAddress([]byte("addr4"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, acc2)

	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	newCoins2 := sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))
	inputs := []types.Input{
		{Address: addr, Coins: newCoins},
		{Address: addr2, Coins: newCoins2},
	}
	outputs := []types.Output{
		{Address: addr3, Coins: newCoins},
		{Address: addr4, Coins: newCoins2},
	}

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events := ctx.EventManager().Events()
	suite.Require().Equal(0, len(events))

	// Set addr's coins but not addr2's coins
	app.BankKeeper.SetBalances(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50)))

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events = ctx.EventManager().Events()
	suite.Require().Equal(1, len(events))

	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event1.Attributes = append(
		event1.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)
	suite.Require().Equal(event1, events[0])

	// Set addr's coins and addr2's coins
	app.BankKeeper.SetBalances(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50)))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100)))
	newCoins2 = sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))

	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	events = ctx.EventManager().Events()
	suite.Require().Equal(5, len(events))

	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event2.Attributes = append(
		event2.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr2.String())},
	)
	event3 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event3.Attributes = append(
		event3.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr3.String())},
	)
	event3.Attributes = append(
		event3.Attributes,
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())})
	event4 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event4.Attributes = append(
		event4.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr4.String())},
	)
	event4.Attributes = append(
		event4.Attributes,
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins2.String())},
	)

	suite.Require().Equal(event1, events[1])
	suite.Require().Equal(event2, events[2])
	suite.Require().Equal(event3, events[3])
	suite.Require().Equal(event4, events[4])
}

func (suite *IntegrationTestSuite) TestSpendableCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule)
	bacc := auth.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, macc)
	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, origCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.SpendableCoins(ctx, addr2))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins), app.BankKeeper.SpendableCoins(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins.Add(sendCoins...)))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	bacc := auth.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins.Add(sendCoins...)))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))), origCoins)
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))), origCoins)
}

func (suite *IntegrationTestSuite) TestDelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	bacc := auth.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestDelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}}
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, invalidCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))

	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)

	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, origCoins.Add(origCoins...)))
}

func (suite *IntegrationTestSuite) TestUndelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(origCoins.Sub(delCoins), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a non-vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))

	suite.Require().Equal(origCoins.Sub(delCoins), app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())
}

func (suite *IntegrationTestSuite) TestUndelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(app.BankKeeper.SetBalances(ctx, addr1, origCoins))

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
