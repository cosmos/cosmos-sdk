<<<<<<< HEAD
package bank

import (
=======
package simplestake

import (
	"fmt"

>>>>>>> asdf
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
<<<<<<< HEAD
=======
	crypto "github.com/tendermint/go-crypto"
>>>>>>> asdf
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	oldwire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/x/auth"
=======
	"github.com/cosmos/cosmos-sdk/x/bank"
>>>>>>> asdf
)

func setupMultiStore() (sdk.MultiStore, *sdk.KVStoreKey) {
	db := dbm.NewMemDB()
	authKey := sdk.NewKVStoreKey("authkey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms, authKey
}

func TestCoinKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

<<<<<<< HEAD
	// wire registration while we're at it ... TODO
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, 0x1},
	)

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	accountMapper := auth.NewAccountMapperSealed(authKey, &auth.BaseAccount{})
	coinKeeper := NewCoinKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	addr2 := sdk.Address([]byte("addr2"))
	addr3 := sdk.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 5}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 15}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 5}}))

	// Test AddCoins
	coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"foocoin", 15}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 25}}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"barcoin", 15}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 15}, {"foocoin", 25}}))

	// Test SubtractCoins
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 15}}))

	_, err := coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 11}})
	assert.Implements(t, (*sdk.Error)(nil), err)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 15}}))

	coinKeeper.SubtractCoins(ctx, addr, sdk.Coins{{"barcoin", 10}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 15}}))
	assert.False(t, coinKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 1}}))

	// Test SendCoins
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

	err2 := coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 50}})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"barcoin", 30}})
	coinKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"barcoin", 10}, {"foocoin", 5}})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 5}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 10}}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{{"foocoin", 2}})
	output1 := NewOutput(addr, sdk.Coins{{"foocoin", 2}})
	coinKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 7}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 8}}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{{"foocoin", 3}}),
		NewInput(addr2, sdk.Coins{{"barcoin", 3}, {"foocoin", 2}}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{{"barcoin", 1}}),
		NewOutput(addr3, sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}),
	}
	coinKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 21}, {"foocoin", 4}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 7}, {"foocoin", 6}}))
	assert.True(t, coinKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}))

}

func TestSendKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	// wire registration while we're at it ... TODO
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, 0x1},
	)

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	accountMapper := auth.NewAccountMapperSealed(authKey, &auth.BaseAccount{})
	coinKeeper := NewCoinKeeper(accountMapper)
	sendKeeper := NewSendKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	addr2 := sdk.Address([]byte("addr2"))
	addr3 := sdk.Address([]byte("addr3"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 5}}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 15}}))
	assert.False(t, sendKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 5}}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{{"foocoin", 15}})

	// Test SendCoins
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 5}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

	err2 := sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"foocoin", 50}})
	assert.Implements(t, (*sdk.Error)(nil), err2)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"foocoin", 5}}))

	coinKeeper.AddCoins(ctx, addr, sdk.Coins{{"barcoin", 30}})
	sendKeeper.SendCoins(ctx, addr, addr2, sdk.Coins{{"barcoin", 10}, {"foocoin", 5}})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 5}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 10}}))

	// Test InputOutputCoins
	input1 := NewInput(addr2, sdk.Coins{{"foocoin", 2}})
	output1 := NewOutput(addr, sdk.Coins{{"foocoin", 2}})
	sendKeeper.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 20}, {"foocoin", 7}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 10}, {"foocoin", 8}}))

	inputs := []Input{
		NewInput(addr, sdk.Coins{{"foocoin", 3}}),
		NewInput(addr2, sdk.Coins{{"barcoin", 3}, {"foocoin", 2}}),
	}

	outputs := []Output{
		NewOutput(addr, sdk.Coins{{"barcoin", 1}}),
		NewOutput(addr3, sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}),
	}
	sendKeeper.InputOutputCoins(ctx, inputs, outputs)
	assert.True(t, sendKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"barcoin", 21}, {"foocoin", 4}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr2).IsEqual(sdk.Coins{{"barcoin", 7}, {"foocoin", 6}}))
	assert.True(t, sendKeeper.GetCoins(ctx, addr3).IsEqual(sdk.Coins{{"barcoin", 2}, {"foocoin", 5}}))

}

func TestViewKeeper(t *testing.T) {
	ms, authKey := setupMultiStore()

	// wire registration while we're at it ... TODO
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, 0x1},
	)

	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	accountMapper := auth.NewAccountMapperSealed(authKey, &auth.BaseAccount{})
	coinKeeper := NewCoinKeeper(accountMapper)
	viewKeeper := NewViewKeeper(accountMapper)

	addr := sdk.Address([]byte("addr1"))
	acc := accountMapper.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	accountMapper.SetAccount(ctx, acc)
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{}))

	coinKeeper.SetCoins(ctx, addr, sdk.Coins{{"foocoin", 10}})
	assert.True(t, viewKeeper.GetCoins(ctx, addr).IsEqual(sdk.Coins{{"foocoin", 10}}))

	// Test HasCoins
	assert.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 10}}))
	assert.True(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 5}}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{{"foocoin", 15}}))
	assert.False(t, viewKeeper.HasCoins(ctx, addr, sdk.Coins{{"barcoin", 5}}))
=======
	ctx := sdk.NewContext(ms, abci.Header{}, false, nil)
	stakeKeeper := NewKeeper(capKey, bank.NewCoinKeeper(nil))
	addr := sdk.Address([]byte("some-address"))

	bi := stakeKeeper.getBondInfo(ctx, addr)
	assert.Equal(t, bi, bondInfo{})

	privKey := crypto.GenPrivKeyEd25519()

	bi = bondInfo{
		PubKey: privKey.PubKey(),
		Power:  int64(10),
	}
	fmt.Printf("Pubkey: %v\n", privKey.PubKey())
	stakeKeeper.setBondInfo(ctx, addr, bi)

	savedBi := stakeKeeper.getBondInfo(ctx, addr)
	assert.NotNil(t, savedBi)
	fmt.Printf("Bond Info: %v\n", savedBi)
	assert.Equal(t, int64(10), savedBi.Power)
>>>>>>> asdf
}
