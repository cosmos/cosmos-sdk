package bank

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtime "github.com/tendermint/tendermint/types/time"

	codec "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/auth"
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey
}

func TestKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := accountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountKeeper.SetAccount(ctx, acc)
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
	assert.Implements(t, (*sdk.Error)(nil), err2)
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
	ms, authKey := setupMultiStore()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)
	sendKeeper := NewBaseSendKeeper(accountKeeper)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := accountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountKeeper.SetAccount(ctx, acc)
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

	_, err2 := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 50)})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	bankKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 30)})
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10)}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	output1 := NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	sendKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8)}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 3)}),
		NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2)}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 1)}),
		NewOutput(addr3, sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}),
	}
	sendKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}))

}

func TestViewKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)
	viewKeeper := NewBaseViewKeeper(accountKeeper)

	addr := sdk.AccAddress([]byte("addr1"))
	acc := accountKeeper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountKeeper.SetAccount(ctx, acc)
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
	ms, authKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	ctx := sdk.NewContext(ms, abci.Header{Time: now}, false, log.NewNopLogger())

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	sendCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}

	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	vacc := auth.NewContinuousVestingAccount(addr1, origCoins, ctx.BlockHeader().Time, endTime)
	accountKeeper.SetAccount(ctx, vacc)

	// require that no coins be sendable at the beginning of the vesting schedule
	_, err := bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	require.Error(t, err)

	// receive some coins
	vacc.SetCoins(origCoins.Plus(sendCoins))
	accountKeeper.SetAccount(ctx, vacc)

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	_, err = bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins)
	vacc = accountKeeper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, origCoins, vacc.GetCoins())
}

func TestVestingAccountReceive(t *testing.T) {
	ms, authKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	ctx := sdk.NewContext(ms, abci.Header{Time: now}, false, log.NewNopLogger())

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	sendCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}

	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewContinuousVestingAccount(addr1, origCoins, ctx.BlockHeader().Time, endTime)
	acc := accountKeeper.NewAccountWithAddress(ctx, addr2)
	accountKeeper.SetAccount(ctx, vacc)
	accountKeeper.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	// send some coins to the vesting account
	bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins)

	// require the coins are spendable
	vacc = accountKeeper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, origCoins.Plus(sendCoins), vacc.GetCoins())
	require.Equal(t, vacc.SpendableCoins(now), sendCoins)

	// require coins are spendable plus any that have vested
	require.Equal(t, vacc.SpendableCoins(now.Add(12*time.Hour)), origCoins)
}

func TestDelegateCoins(t *testing.T) {
	ms, authKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	ctx := sdk.NewContext(ms, abci.Header{Time: now}, false, log.NewNopLogger())

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	delCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}

	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewContinuousVestingAccount(addr1, origCoins, ctx.BlockHeader().Time, endTime)
	acc := accountKeeper.NewAccountWithAddress(ctx, addr2)
	accountKeeper.SetAccount(ctx, vacc)
	accountKeeper.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	_, err := bankKeeper.DelegateCoins(ctx, addr2, delCoins)
	acc = accountKeeper.GetAccount(ctx, addr2)
	require.NoError(t, err)
	require.Equal(t, delCoins, acc.GetCoins())

	// require the ability for a vesting account to delegate
	_, err = bankKeeper.DelegateCoins(ctx, addr1, delCoins)
	vacc = accountKeeper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.NoError(t, err)
	require.Equal(t, delCoins, vacc.GetCoins())
}

func TestUndelegateCoins(t *testing.T) {
	ms, authKey := setupMultiStore()
	cdc := codec.New()
	auth.RegisterBaseAccount(cdc)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	ctx := sdk.NewContext(ms, abci.Header{Time: now}, false, log.NewNopLogger())

	origCoins := sdk.Coins{sdk.NewInt64Coin("steak", 100)}
	delCoins := sdk.Coins{sdk.NewInt64Coin("steak", 50)}

	accountKeeper := auth.NewAccountKeeper(cdc, authKey, auth.ProtoBaseAccount)
	bankKeeper := NewBaseKeeper(accountKeeper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewContinuousVestingAccount(addr1, origCoins, ctx.BlockHeader().Time, endTime)
	acc := accountKeeper.NewAccountWithAddress(ctx, addr2)
	accountKeeper.SetAccount(ctx, vacc)
	accountKeeper.SetAccount(ctx, acc)
	bankKeeper.SetCoins(ctx, addr2, origCoins)

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to undelegate
	_, err := bankKeeper.DelegateCoins(ctx, addr2, delCoins)
	require.NoError(t, err)

	_, err = bankKeeper.UndelegateCoins(ctx, addr2, delCoins)
	require.NoError(t, err)
	acc = accountKeeper.GetAccount(ctx, addr2)
	require.Equal(t, origCoins, acc.GetCoins())

	// require the ability for a vesting account to delegate
	_, err = bankKeeper.DelegateCoins(ctx, addr1, delCoins)
	require.NoError(t, err)

	_, err = bankKeeper.UndelegateCoins(ctx, addr1, delCoins)
	require.NoError(t, err)
	vacc = accountKeeper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, origCoins, vacc.GetCoins())
}
