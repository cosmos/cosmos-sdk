package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/auth"
)

func setupMultiStore() (bam.MultiStore, *bam.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := bam.NewKVStoreKey("authkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authKey, bam.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey
}

func TestKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := bam.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)

	addr := bam.Address([]byte("addr1"))
	addr2 := bam.Address([]byte("addr2"))
	addr3 := bam.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{}))

	coinKeeper.SetCoins(ctx, addr, bam.Coins{{"foocoin", 10}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, coinKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 5}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 15}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, bam.Coins{{"barcoin", 5}}))

	// Test AddCoins
	coinKeeper.AddCoins(ctx, addr, bam.Coins{{"foocoin", 15}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 25}}))

	coinKeeper.AddCoins(ctx, addr, bam.Coins{{"barcoin", 15}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 15}, {"foocoin", 25}}))

	// Test SubtractCoins
	coinKeeper.SubtractCoins(ctx, addr, bam.Coins{{"foocoin", 10}})
	coinKeeper.SubtractCoins(ctx, addr, bam.Coins{{"barcoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 15}}))

	_, err := coinKeeper.SubtractCoins(ctx, addr, bam.Coins{{"barcoin", 11}})
	assert.Implements(t, (*sdk.Error)(nil), err)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 15}}))

	coinKeeper.SubtractCoins(ctx, addr, bam.Coins{{"barcoin", 10}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 15}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, bam.Coins{{"barcoin", 1}}))

	// Test SendCoins
	coinKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"foocoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"foocoin", 5}}))

	err2 := coinKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"foocoin", 50}})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"foocoin", 5}}))

	coinKeeper.AddCoins(ctx, addr, bam.Coins{{"barcoin", 30}})
	coinKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"barcoin", 10}, {"foocoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 20}, {"foocoin", 5}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 10}}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, bam.Coins{{"foocoin", 2}})
	output1 := NewOutput(addr, bam.Coins{{"foocoin", 2}})
	coinKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 20}, {"foocoin", 7}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 8}}))

	inputs := []Input{
		NewInput(addr, bam.Coins{{"foocoin", 3}}),
		NewInput(addr2, bam.Coins{{"barcoin", 3}, {"foocoin", 2}}),
	}

	outputs := []Output{
		NewOutput(addr, bam.Coins{{"barcoin", 1}}),
		NewOutput(addr3, bam.Coins{{"barcoin", 2}, {"foocoin", 5}}),
	}
	coinKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 21}, {"foocoin", 4}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 7}, {"foocoin", 6}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr3).IsEqual(bam.Coins{{"barcoin", 2}, {"foocoin", 5}}))

}

func TestSendKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := bam.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)
	sendKeeper := NewSendKeeper(accountMapper)

	addr := bam.Address([]byte("addr1"))
	addr2 := bam.Address([]byte("addr2"))
	addr3 := bam.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{}))

	coinKeeper.SetCoins(ctx, addr, bam.Coins{{"foocoin", 10}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, sendKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 5}}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 15}}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, bam.Coins{{"barcoin", 5}}))

	coinKeeper.SetCoins(ctx, addr, bam.Coins{{"foocoin", 15}})

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"foocoin", 5}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"foocoin", 5}}))

	err2 := sendKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"foocoin", 50}})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"foocoin", 5}}))

	coinKeeper.AddCoins(ctx, addr, bam.Coins{{"barcoin", 30}})
	sendKeeper.SendCoins(ctx, addr, addr2, bam.Coins{{"barcoin", 10}, {"foocoin", 5}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 20}, {"foocoin", 5}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 10}}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, bam.Coins{{"foocoin", 2}})
	output1 := NewOutput(addr, bam.Coins{{"foocoin", 2}})
	sendKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 20}, {"foocoin", 7}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 10}, {"foocoin", 8}}))

	inputs := []Input{
		NewInput(addr, bam.Coins{{"foocoin", 3}}),
		NewInput(addr2, bam.Coins{{"barcoin", 3}, {"foocoin", 2}}),
	}

	outputs := []Output{
		NewOutput(addr, bam.Coins{{"barcoin", 1}}),
		NewOutput(addr3, bam.Coins{{"barcoin", 2}, {"foocoin", 5}}),
	}
	sendKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"barcoin", 21}, {"foocoin", 4}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(bam.Coins{{"barcoin", 7}, {"foocoin", 6}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr3).IsEqual(bam.Coins{{"barcoin", 2}, {"foocoin", 5}}))

}

func TestViewKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)

	ctx := bam.NewContext(ms, abci.Header{}, false, nil, log.NewNopLogger())
	accountMapper := auth.NewAccountMapper(cdc, authKey, &auth.BaseAccount{})
	coinKeeper := NewKeeper(accountMapper)
	viewKeeper := NewViewKeeper(accountMapper)

	addr := bam.Address([]byte("addr1"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{}))

	coinKeeper.SetCoins(ctx, addr, bam.Coins{{"foocoin", 10}})
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(bam.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, viewKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 10}}))
	assert.True(t, viewKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 5}}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, bam.Coins{{"foocoin", 15}}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, bam.Coins{{"barcoin", 5}}))
}
