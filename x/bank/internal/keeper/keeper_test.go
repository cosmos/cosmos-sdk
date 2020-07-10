package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmkv "github.com/tendermint/tendermint/libs/kv"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	keep "github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

func TestKeeper(t *testing.T) {
	app, ctx := createTestApp(false)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	app.AccountKeeper.SetAccount(ctx, acc)
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, app.BankKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, app.BankKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, app.BankKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, app.BankKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))

	// Test AddCoins
	app.BankKeeper.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 25))))

	app.BankKeeper.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15), sdk.NewInt64Coin("foocoin", 25))))

	// Test SubtractCoins
	app.BankKeeper.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	app.BankKeeper.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	app.BankKeeper.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 11)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	app.BankKeeper.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, app.BankKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))))

	// Test SendCoins
	app.BankKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	app.BankKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	app.BankKeeper.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	app.BankKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	output1 := types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	app.BankKeeper.InputOutputCoins(ctx, []types.Input{input1}, []types.Output{output1})
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8))))

	inputs := []types.Input{
		types.NewInput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 3))),
		types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2))),
	}

	outputs := []types.Output{
		types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))),
		types.NewOutput(addr3, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))),
	}
	app.BankKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, app.BankKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6))))
	require.True(t, app.BankKeeper.GetCoins(ctx, addr3).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))))

	// Test retrieving black listed accounts
	for acc := range simapp.GetMaccPerms() {
		addr := supply.NewModuleAddress(acc)
		require.Equal(t, app.BlacklistedAccAddrs()[addr.String()], app.BankKeeper.BlacklistedAddr(addr))
	}
}

func TestSendKeeper(t *testing.T) {
	app, ctx := createTestApp(false)

	blacklistedAddrs := make(map[string]bool)

	paramSpace := app.ParamsKeeper.Subspace("newspace")
	sendKeeper := keep.NewBaseSendKeeper(app.AccountKeeper, paramSpace, blacklistedAddrs)
	app.BankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	app.AccountKeeper.SetAccount(ctx, acc)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))

	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	app.BankKeeper.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	err := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{Denom: "FOOCOIN", Amount: sdk.NewInt(-5)}})
	require.Error(t, err)
}

func TestInputOutputNewAccount(t *testing.T) {
	app, ctx := createTestApp(false)
	balances := sdk.NewCoins(sdk.NewInt64Coin("foo", 100), sdk.NewInt64Coin("bar", 50))
	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	app.AccountKeeper.SetAccount(ctx, acc1)
	require.NoError(t, app.BankKeeper.SetCoins(ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetCoins(ctx, addr1)
	require.Equal(t, balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2"))

	require.Nil(t, app.AccountKeeper.GetAccount(ctx, addr2))
	require.Empty(t, app.BankKeeper.GetCoins(ctx, addr2))

	inputs := []types.Input{
		{Address: addr1, Coins: sdk.NewCoins(sdk.NewInt64Coin("foo", 30), sdk.NewInt64Coin("bar", 10))},
	}
	outputs := []types.Output{
		{Address: addr2, Coins: sdk.NewCoins(sdk.NewInt64Coin("foo", 30), sdk.NewInt64Coin("bar", 10))},
	}

	require.NoError(t, app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	expected := sdk.NewCoins(sdk.NewInt64Coin("foo", 30), sdk.NewInt64Coin("bar", 10))
	acc2Balances := app.BankKeeper.GetCoins(ctx, addr2)
	require.Equal(t, expected, acc2Balances)
	require.NotNil(t, app.AccountKeeper.GetAccount(ctx, addr2))
}

func TestMsgSendEvents(t *testing.T) {
	app, ctx := createTestApp(false)

	app.BankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50))
	err := app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins)
	require.Error(t, err)
	events := ctx.EventManager().Events()
	require.Equal(t, 2, len(events))
	event1 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event1.Attributes = append(
		event1.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr2.String())},
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())})
	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event2.Attributes = append(
		event2.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())})
	require.Equal(t, event1, events[0])
	require.Equal(t, event2, events[1])

	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50))
	err = app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins)
	require.NoError(t, err)
	events = ctx.EventManager().Events()
	require.Equal(t, 4, len(events))
	require.Equal(t, event1, events[2])
	require.Equal(t, event2, events[3])
}

func TestMsgMultiSendEvents(t *testing.T) {
	app, ctx := createTestApp(false)

	app.BankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	addr4 := sdk.AccAddress([]byte("addr4"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, acc2)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50))
	newCoins2 := sdk.NewCoins(sdk.NewInt64Coin("barcoin", 100))
	inputs := []types.Input{
		{Address: addr, Coins: newCoins},
		{Address: addr2, Coins: newCoins2},
	}
	outputs := []types.Output{
		{Address: addr3, Coins: newCoins},
		{Address: addr4, Coins: newCoins2},
	}
	err := app.BankKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.Error(t, err)
	events := ctx.EventManager().Events()
	require.Equal(t, 0, len(events))

	// Set addr's coins but not addr2's coins
	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))

	err = app.BankKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.Error(t, err)
	events = ctx.EventManager().Events()
	require.Equal(t, 1, len(events))
	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event1.Attributes = append(
		event1.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())})
	require.Equal(t, event1, events[0])

	// Set addr's coins and addr2's coins
	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50))
	app.BankKeeper.SetCoins(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 100)))
	newCoins2 = sdk.NewCoins(sdk.NewInt64Coin("barcoin", 100))

	err = app.BankKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.NoError(t, err)
	events = ctx.EventManager().Events()
	require.Equal(t, 5, len(events))
	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []tmkv.Pair{},
	}
	event2.Attributes = append(
		event2.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeySender), Value: []byte(addr2.String())})
	event3 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event3.Attributes = append(
		event3.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr3.String())})
	event3.Attributes = append(
		event3.Attributes,
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())})
	event4 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []tmkv.Pair{},
	}
	event4.Attributes = append(
		event4.Attributes,
		tmkv.Pair{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr4.String())})
	event4.Attributes = append(
		event4.Attributes,
		tmkv.Pair{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins2.String())})
	require.Equal(t, event1, events[1])
	require.Equal(t, event2, events[2])
	require.Equal(t, event3, events[3])
	require.Equal(t, event4, events[4])
}

func TestViewKeeper(t *testing.T) {
	app, ctx := createTestApp(false)

	//paramSpace := app.ParamsKeeper.Subspace(types.DefaultParamspace)
	viewKeeper := keep.NewBaseViewKeeper(app.AccountKeeper)

	addr := sdk.AccAddress([]byte("addr1"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	app.AccountKeeper.SetAccount(ctx, acc)
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	app.BankKeeper.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))
}

func TestVestingAccountSend(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := vesting.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	app.AccountKeeper.SetAccount(ctx, vacc)

	// require that no coins be sendable at the beginning of the vesting schedule
	err := app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	require.Error(t, err)

	// receive some coins
	vacc.SetCoins(origCoins.Add(sendCoins...))
	app.AccountKeeper.SetAccount(ctx, vacc)

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	err = app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, origCoins, vacc.GetCoins())
}

func TestPeriodicVestingAccountSend(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}
	vacc := vesting.NewPeriodicVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), periods)
	app.AccountKeeper.SetAccount(ctx, vacc)

	// require that no coins be sendable at the beginning of the vesting schedule
	err := app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	require.Error(t, err)

	// receive some coins
	vacc.SetCoins(origCoins.Add(sendCoins...))
	app.AccountKeeper.SetAccount(ctx, vacc)

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	err = app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	require.NoError(t, err)
	require.Equal(t, origCoins, vacc.GetCoins())
}

func TestVestingAccountReceive(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := vesting.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.BankKeeper.SetCoins(ctx, addr2, origCoins)

	// send some coins to the vesting account
	app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins)

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	require.Equal(t, origCoins.Add(sendCoins...), vacc.GetCoins())
	require.Equal(t, vacc.SpendableCoins(now), sendCoins)

	// require coins are spendable plus any that have vested
	require.Equal(t, vacc.SpendableCoins(now.Add(12*time.Hour)), origCoins)
}

func TestPeriodicVestingAccountReceive(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}
	vacc := vesting.NewPeriodicVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), periods)
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.BankKeeper.SetCoins(ctx, addr2, origCoins)

	// send some coins to the vesting account
	app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins)

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	require.Equal(t, origCoins.Add(sendCoins...), vacc.GetCoins())
	require.Equal(t, vacc.SpendableCoins(now), sendCoins)

	// require coins are spendable plus any that have vested
	require.Equal(t, vacc.SpendableCoins(now.Add(12*time.Hour)), origCoins)
}

func TestDelegateCoins(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)
	ak := app.AccountKeeper

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	macc := ak.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	vacc := vesting.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := ak.NewAccountWithAddress(ctx, addr2)
	ak.SetAccount(ctx, vacc)
	ak.SetAccount(ctx, acc)
	ak.SetAccount(ctx, macc)
	app.BankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	acc = ak.GetAccount(ctx, addr2)
	macc = ak.GetAccount(ctx, addrModule)
	require.NoError(t, err)
	require.Equal(t, origCoins.Sub(delCoins), acc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a vesting account to delegate
	err = app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins)
	vacc = ak.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, delCoins, vacc.GetCoins())
}

func TestUndelegateCoins(t *testing.T) {
	app, ctx := createTestApp(false)
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)
	ak := app.AccountKeeper

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	macc := ak.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	vacc := vesting.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := ak.NewAccountWithAddress(ctx, addr2)
	ak.SetAccount(ctx, vacc)
	ak.SetAccount(ctx, acc)
	ak.SetAccount(ctx, macc)
	app.BankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	require.NoError(t, err)

	acc = ak.GetAccount(ctx, addr2)
	macc = ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins.Sub(delCoins), acc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a non-vesting account to undelegate
	err = app.BankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins)
	require.NoError(t, err)

	acc = ak.GetAccount(ctx, addr2)
	macc = ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins, acc.GetCoins())
	require.True(t, macc.GetCoins().Empty())

	// require the ability for a vesting account to delegate
	err = app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins)
	require.NoError(t, err)

	vacc = ak.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	macc = ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins.Sub(delCoins), vacc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a vesting account to undelegate
	err = app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins)
	require.NoError(t, err)

	vacc = ak.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	macc = ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins, vacc.GetCoins())
	require.True(t, macc.GetCoins().Empty())
}
