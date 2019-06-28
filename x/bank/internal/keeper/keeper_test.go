package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	k   Keeper
	ak  auth.AccountKeeper
	pk  params.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	pk := params.NewKeeper(cdc, keyParams, tkeyParams, params.DefaultCodespace)

	ak := auth.NewAccountKeeper(
		cdc, authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())

	bankKeeper := NewBaseKeeper(ak, pk.Subspace(types.DefaultParamspace), types.DefaultCodespace)
	bankKeeper.SetSendEnabled(ctx, true)

	return testInput{cdc: cdc, ctx: ctx, k: bankKeeper, ak: ak, pk: pk}
}

func TestKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	input.k.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))

	// Test AddCoins
	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 25))))

	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 15), sdk.NewInt64Coin("foocoin", 25))))

	// Test SubtractCoins
	input.k.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	input.k.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	input.k.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 11)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15))))

	input.k.SubtractCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, input.k.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))))

	// Test SendCoins
	input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, input.k.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	err2 := input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.Implements(t, (*sdk.Error)(nil), err2)
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, input.k.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	input.k.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, input.k.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	output1 := types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 2)))
	input.k.InputOutputCoins(ctx, []types.Input{input1}, []types.Output{output1})
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7))))
	require.True(t, input.k.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8))))

	inputs := []types.Input{
		types.NewInput(addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 3))),
		types.NewInput(addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2))),
	}

	outputs := []types.Output{
		types.NewOutput(addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 1))),
		types.NewOutput(addr3, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))),
	}
	input.k.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, input.k.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4))))
	require.True(t, input.k.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6))))
	require.True(t, input.k.GetCoins(ctx, addr3).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5))))
}

func TestSendKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	paramSpace := input.pk.Subspace("newspace")
	sendKeeper := NewBaseSendKeeper(input.ak, paramSpace, types.DefaultCodespace)
	input.k.SetSendEnabled(ctx, true)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	input.k.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))

	input.k.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15)))

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	err := sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 50)))
	require.Implements(t, (*sdk.Error)(nil), err)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))

	input.k.AddCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 30)))
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)))
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5))))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	err = sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{"FOOCOIN", sdk.NewInt(-5)}})
	require.Error(t, err)
}

func TestViewKeeper(t *testing.T) {
	input := setupTestInput()
	ctx := input.ctx
	//paramSpace := input.pk.Subspace(types.DefaultParamspace)
	viewKeeper := NewBaseViewKeeper(input.ak, types.DefaultCodespace)

	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.ak.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	input.ak.SetAccount(ctx, acc)
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))

	input.k.SetCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10)))
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))

	// Test HasCoins
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 10))))
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 5))))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("foocoin", 15))))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.NewCoins(sdk.NewInt64Coin("barcoin", 5))))
}

func TestVestingAccountSend(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	input.ak.SetAccount(ctx, vacc)

	// require that no coins be sendable at the beginning of the vesting schedule
	err := input.k.SendCoins(ctx, addr1, addr2, sendCoins)
	require.Error(t, err)

	// receive some coins
	vacc.SetCoins(origCoins.Add(sendCoins))
	input.ak.SetAccount(ctx, vacc)

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	err = input.k.SendCoins(ctx, addr1, addr2, sendCoins)
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, origCoins, vacc.GetCoins())
}

func TestVestingAccountReceive(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	input.k.SetCoins(ctx, addr2, origCoins)

	// send some coins to the vesting account
	input.k.SendCoins(ctx, addr2, addr1, sendCoins)

	// require the coins are spendable
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, origCoins.Add(sendCoins), vacc.GetCoins())
	require.Equal(t, vacc.SpendableCoins(now), sendCoins)

	// require coins are spendable plus any that have vested
	require.Equal(t, vacc.SpendableCoins(now.Add(12*time.Hour)), origCoins)
}

func TestDelegateCoins(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	macc := input.ak.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	input.ak.SetAccount(ctx, macc)
	input.k.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := input.k.DelegateCoins(ctx, addr2, addrModule, delCoins)
	acc = input.ak.GetAccount(ctx, addr2)
	macc = input.ak.GetAccount(ctx, addrModule)
	require.NoError(t, err)
	require.Equal(t, origCoins.Sub(delCoins), acc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a vesting account to delegate
	err = input.k.DelegateCoins(ctx, addr1, addrModule, delCoins)
	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, delCoins, vacc.GetCoins())
}

func TestUndelegateCoins(t *testing.T) {
	input := setupTestInput()
	now := tmtime.Now()
	ctx := input.ctx.WithBlockHeader(abci.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addrModule := sdk.AccAddress([]byte("moduleAcc"))

	bacc := auth.NewBaseAccountWithAddress(addr1)
	bacc.SetCoins(origCoins)
	macc := input.ak.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	vacc := auth.NewContinuousVestingAccount(&bacc, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := input.ak.NewAccountWithAddress(ctx, addr2)
	input.ak.SetAccount(ctx, vacc)
	input.ak.SetAccount(ctx, acc)
	input.ak.SetAccount(ctx, macc)
	input.k.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := input.k.DelegateCoins(ctx, addr2, addrModule, delCoins)
	require.NoError(t, err)

	acc = input.ak.GetAccount(ctx, addr2)
	macc = input.ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins.Sub(delCoins), acc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a non-vesting account to undelegate
	err = input.k.UndelegateCoins(ctx, addrModule, addr2, delCoins)
	require.NoError(t, err)

	acc = input.ak.GetAccount(ctx, addr2)
	macc = input.ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins, acc.GetCoins())
	require.True(t, macc.GetCoins().Empty())

	// require the ability for a vesting account to delegate
	err = input.k.DelegateCoins(ctx, addr1, addrModule, delCoins)
	require.NoError(t, err)

	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	macc = input.ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins.Sub(delCoins), vacc.GetCoins())
	require.Equal(t, delCoins, macc.GetCoins())

	// require the ability for a vesting account to undelegate
	err = input.k.UndelegateCoins(ctx, addrModule, addr1, delCoins)
	require.NoError(t, err)

	vacc = input.ak.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	macc = input.ak.GetAccount(ctx, addrModule)
	require.Equal(t, origCoins, vacc.GetCoins())
	require.True(t, macc.GetCoins().Empty())
}
