package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"

	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	oldwire "github.com/tendermint/go-wire"
)

func defaultContext(keys ...sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	}
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	return ctx
}

func makeCodec() {
	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, accTypeApp},
	)
}

func newAddress() sdk.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func newCoinKeeper(key sdk.StoreKey) CoinKeeper {
	am := auth.NewAccountMapper(key, &auth.BaseAccount{})
	return NewCoinKeeper(am)
}

func TestCoinKeeper(t *testing.T) {
	makeCodec()
	key := sdk.NewKVStoreKey("bank")
	ck := newCoinKeeper(key)
	ctx := defaultContext(key)

	zero := sdk.Coins{}
	one := sdk.Coins{{"atom", 1}}
	addr := newAddress()

	coins := ck.GetCoins(ctx, addr)
	assert.Equal(t, zero, coins)

	coins = ck.AddCoins(ctx, addr, one)
	assert.Equal(t, one, coins)

	coins, err := ck.SubtractCoins(ctx, addr, one)
	assert.Nil(t, err)
	assert.Equal(t, zero, coins)

	coins, err = ck.SubtractCoins(ctx, addr, one)
	assert.NotNil(t, err)
	assert.Equal(t, one, coins)
}
