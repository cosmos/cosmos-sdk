package bank

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  auth.AccountKeeper
	pk  params.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	fckCapKey := sdk.NewKVStoreKey("fckCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(fckCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	ak := auth.NewAccountKeeper(
		cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, pk: pk}
}

func TestKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	bankKeeper := NewBaseKeeper(input.ak, input.pk.Subspace(DefaultParamspace), DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	bankKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, bankKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, bankKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, bankKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, bankKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))

	// Test AddCoins
	bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 25)}))

	bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 15)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 15), sdk.NewInt64Coin("foocoin", 25)}))

	// Test SubtractCoins
	bankKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	bankKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15)}))

	bankKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 11)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15)}))

	bankKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 10)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, bankKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 1)}))

	// Test SendCoins
	bankKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	_, err2 := bankKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 50)})
	require.Implements(t, (*sdk.Error)(nil), err2)
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 30)})
	bankKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10)}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	output1 := NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	bankKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8)}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 3)}),
		NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2)}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 1)}),
		NewOutput(addr3, sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}),
	}
	bankKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, bankKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6)}))
	require.True(t, bankKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}))
}

func TestSendKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	paramSpace := input.pk.Subspace(DefaultParamspace)
	bankKeeper := NewBaseKeeper(input.ak, paramSpace, DefaultCodespace)
	sendKeeper := NewBaseSendKeeper(input.ak, paramSpace, DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	bankKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))

	bankKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)})

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	_, err := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 50)})
	require.Implements(t, (*sdk.Error)(nil), err)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 30)})
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10)}))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	_, err = sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{"FOOCOIN", sdk.NewInt(-5)}})
	require.Error(t, err)
}

func TestViewKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	paramSpace := input.pk.Subspace(DefaultParamspace)
	bankKeeper := NewBaseKeeper(input.ak, paramSpace, DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)
	viewKeeper := NewBaseViewKeeper(input.ak, DefaultCodespace)

	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	bankKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))
}

func TestVestingAccountSend(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	sendCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}
	bankKeeper := NewBaseKeeper(input.ak, input.pk.Subspace(DefaultParamspace), DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	input.ak.SetAccount(ctx, vacc)

	// require that no coins be sendable at the beginning of the vesting schedule
	_, err := bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	require.Error(t, err)

	// receive some coins
	vacc.SetCoins(origCoins.Plus(sendCoins))
	input.ak.SetAccount(ctx, vacc)

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	_, err = bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, origCoins, vacc.GetCoins())
}

func TestVestingAccountReceive(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	sendCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}
	bankKeeper := NewBaseKeeper(input.ak, input.pk.Subspace(DefaultParamspace), DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	// send some coins to the vesting account
	bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins)

	// require the coins are spendable
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, origCoins.Plus(sendCoins), vacc.GetCoins())
	require.Equal(t, vacc.SpendableCoins(now), sendCoins)

	// require coins are spendable plus any that have vested
	require.Equal(t, vacc.SpendableCoins(now.Add(12*time.Hour)), origCoins)
}

func TestDelegateCoins(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	delCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}
	bankKeeper := NewBaseKeeper(input.ak, input.pk.Subspace(DefaultParamspace), DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	_, err := bankKeeper.DelegateCoins(ctx, addr2, delCoins)
	acc = input.ak.GetAccount(ctx, addr2)
	require.NoError(t, err)
	require.Equal(t, delCoins, acc.GetCoins())

	// require the ability for a vesting account to delegate
	_, err = bankKeeper.DelegateCoins(ctx, addr1, delCoins)
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, delCoins, vacc.GetCoins())
}

func TestUndelegateCoins(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	delCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}
	bankKeeper := NewBaseKeeper(input.ak, input.pk.Subspace(DefaultParamspace), DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	_, err := bankKeeper.DelegateCoins(ctx, addr2, delCoins)
	require.NoError(t, err)

	// require the ability for a non-vesting account to undelegate
	_, err = bankKeeper.UndelegateCoins(ctx, addr2, delCoins)
	require.NoError(t, err)

	acc = input.ak.GetAccount(ctx, addr2)
	require.Equal(t, origCoins, acc.GetCoins())

	// require the ability for a vesting account to delegate
	_, err = bankKeeper.DelegateCoins(ctx, addr1, delCoins)
	require.NoError(t, err)

	// require the ability for a vesting account to undelegate
	_, err = bankKeeper.UndelegateCoins(ctx, addr1, delCoins)
	require.NoError(t, err)

	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, origCoins, vacc.GetCoins())
}
