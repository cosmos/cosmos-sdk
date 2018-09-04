package bank

import (
	"time"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

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

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))

	// Test AddCoins
	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 25)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 15)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 15), sdk.NewInt64Coin("foocoin", 25)}))

	// Test SubtractCoins
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15)}))

	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 11)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 15)}))

	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 10)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 1)}))

	// Test SendCoins
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	_, err2 := coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 50)})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 30)})
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 5)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 10)}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	output1 := NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 2)})
	coinKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 20), sdk.NewInt64Coin("foocoin", 7)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 10), sdk.NewInt64Coin("foocoin", 8)}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 3)}),
		NewInput(addr2, sdk.Coins{sdk.NewInt64Coin("barcoin", 3), sdk.NewInt64Coin("foocoin", 2)}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 1)}),
		NewOutput(addr3, sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}),
	}
	coinKeeper.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 21), sdk.NewInt64Coin("foocoin", 4)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 7), sdk.NewInt64Coin("foocoin", 6)}))
	require.True(t, coinKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{sdk.NewInt64Coin("barcoin", 2), sdk.NewInt64Coin("foocoin", 5)}))

}

func TestSendKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)
	sendKeeper := NewSendKeeper(accountMapper)

	addr := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)})

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)})
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	_, err2 := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.NewInt64Coin("foocoin", 50)})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	require.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 30)})
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

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)
	viewKeeper := NewViewKeeper(accountMapper)

	addr := sdk.AccAddress([]byte("addr1"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)})
	require.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))

	// Test HasCoins
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}))
	require.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 5)}))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("foocoin", 15)}))
	require.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{sdk.NewInt64Coin("barcoin", 5)}))
}

func TestVesting(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(500, 0)}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewContinuousVestingAccount(addr1, sdk.Coins{{"steak", sdk.NewInt(100)}}, time.Unix(0, 0), time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &vacc)

	// Try sending more than sendable coins
    _, err := coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(70)}})

	require.NotNil(t, err, "Keeper did not error")
	require.Equal(t, sdk.CodeType(10), err.Code(), "Did not error with insufficient coins")

	// Send less than sendable coins
	_, err = coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(40)}})

	require.Nil(t, err, "Keeper errored on valid transfer")
	acc := accountMapper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)
	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(-40)}}, acc.TransferredCoins, "Did not track transfers")

	// Receive coins
	addr3 := sdk.AccAddress([]byte("addr3"))
	acc3 := auth.NewBaseAccountWithAddress(addr3)
	acc3.SetCoins(sdk.Coins{{"steak", sdk.NewInt(50)}})
	accountMapper.SetAccount(ctx, &acc3)

	_, err = coinKeeper.SendCoins(ctx, addr3, addr1, sdk.Coins{{"steak", sdk.NewInt(50)}})

	acc = accountMapper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)

	require.Nil(t, err, "Send to a vesting account failed")
	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(10)}}, acc.TransferredCoins, "Transferred coins did not change")

	// Send transferred coins
	_, err = coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(60)}})

	require.Nil(t, err, "Sending transferred coins failed")

	acc = accountMapper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)

	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(-50)}}, acc.TransferredCoins, "Transferred coins did not update correctly")
}

func TestVestingInputOutput(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(500, 0)}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewContinuousVestingAccount(addr1, sdk.Coins{{"steak", sdk.NewInt(100)}}, time.Unix(0, 0), time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &vacc)

	// Send some coins back to self to check if transferredCoins updates correctly
	inputs := []Input{{addr1, sdk.Coins{{"steak", sdk.NewInt(50)}}}}
	outputs := []Output{{addr1, sdk.Coins{{"steak", sdk.NewInt(20)}}}, {addr2, sdk.Coins{{"steak", sdk.NewInt(30)}}}}
	_, err := coinKeeper.InputOutputCoins(ctx, inputs, outputs)

	require.Nil(t, err, "InputOutput failed on valid vested spend")

	acc := accountMapper.GetAccount(ctx, addr1).(*auth.ContinuousVestingAccount)

	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(-30)}}, acc.TransferredCoins, "Transferred coins did not update correctly")
}

func TestDelayTransferSend(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(500, 0)}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	dtacc := auth.NewDelayTransferAccount(addr1, sdk.Coins{{"steak", sdk.NewInt(100)}}, time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &dtacc)

	acc := auth.NewBaseAccountWithAddress(addr2)
	acc.SetCoins(sdk.Coins{{"steak", sdk.NewInt(50)}})
	accountMapper.SetAccount(ctx, &acc)

	// Send coins before EndTime fails
	_, err := coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(1)}})

	require.NotNil(t, err, "Keeper did not error trying to send locked coins")
	require.Equal(t, sdk.CodeType(10), err.Code(), "Did not error with insufficient coins")

	// Receive coins
	coinKeeper.SendCoins(ctx, addr2, addr1, sdk.Coins{{"steak", sdk.NewInt(50)}})

	recoverAcc := accountMapper.GetAccount(ctx, addr1).(*auth.DelayTransferAccount)
	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(50)}}, recoverAcc.TransferredCoins, "Transferred coins did not update correctly")

	// Spend some of Received Coins
	_, err = coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(25)}})

	recoverAcc = accountMapper.GetAccount(ctx, addr1).(*auth.DelayTransferAccount)
	require.Nil(t, err, "Keeper errorred on valid spend")
	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(25)}}, recoverAcc.TransferredCoins, "Transferred coins did not update correctly")

	// Fast-forward to EndTime
	ctx = ctx.WithBlockHeader(abci.Header{Time: time.Unix(1000, 0)})

	// Spend all unlocked coins
	_, err = coinKeeper.SendCoins(ctx, addr1, addr2, sdk.Coins{{"steak", sdk.NewInt(125)}})

	recoverAcc = accountMapper.GetAccount(ctx, addr1).(*auth.DelayTransferAccount)
	require.Nil(t, err, "Keeper errorred on valid spend")
	require.Equal(t, sdk.Coins(nil), recoverAcc.GetCoins(), "SendableCoins is incorrect")
	require.False(t, recoverAcc.IsVesting(ctx.BlockHeader().Time), "Account still vesting after EndTime")
}

func TestDelayTransferInputOutput(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{Time: time.Unix(500, 0)}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	vacc := auth.NewDelayTransferAccount(addr1, sdk.Coins{{"steak", sdk.NewInt(100)}}, time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &vacc)

	acc := auth.NewBaseAccountWithAddress(addr2)
	acc.SetCoins(sdk.Coins{{"steak", sdk.NewInt(50)}})
	accountMapper.SetAccount(ctx, &acc)

	// Transfer coins to delay transfer account
	coinKeeper.SendCoins(ctx, addr2, addr1, sdk.Coins{{"steak", sdk.NewInt(50)}})

	// Send some coins back to self to check if transferredCoins updates correctly
	inputs := []Input{{addr1, sdk.Coins{{"steak", sdk.NewInt(50)}}}}
	outputs := []Output{{addr1, sdk.Coins{{"steak", sdk.NewInt(20)}}}, {addr2, sdk.Coins{{"steak", sdk.NewInt(30)}}}}
	_, err := coinKeeper.InputOutputCoins(ctx, inputs, outputs)

	require.Nil(t, err, "InputOutput failed on valid vested spend")

	recoverAcc := accountMapper.GetAccount(ctx, addr1).(*auth.DelayTransferAccount)

	require.Equal(t, sdk.Coins{{"steak", sdk.NewInt(20)}}, recoverAcc.TransferredCoins, "Transferred coins did not update correctly")
}

func TestSubtractVesting(t *testing.T) {
	// SubtractCoins must still work without vesting restriction so that they can delegate locked coins.
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterAccount(cdc)

	ctx := sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, auth.ProtoBaseAccount)
	coinKeeper := NewKeeper(accountMapper)

	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))
	amt := sdk.Coins{{"steak", sdk.NewInt(100)}}

	vacc := auth.NewContinuousVestingAccount(addr1, amt, time.Unix(0, 0), time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &vacc)

	dtacc := auth.NewDelayTransferAccount(addr2, amt, time.Unix(1000, 0))
	accountMapper.SetAccount(ctx, &dtacc)

	res, _, err := coinKeeper.SubtractCoins(ctx, addr1, amt)

	require.Nil(t, err, "ContinuousVestingAccount fails on SubtractCoins")
	require.Equal(t, sdk.Coins(nil), res, "Coins did not update correctly")

	res, _, err = coinKeeper.SubtractCoins(ctx, addr2, amt)

	require.Nil(t, err, "DelayTransferAccount fails on SubtractCoins")
	require.Equal(t, sdk.Coins(nil), res, "Coins did not update correctly")
}
