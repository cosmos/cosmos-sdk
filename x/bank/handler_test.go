package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"

	dbm "github.com/tendermint/tmlibs/db"

	oldwire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	return ctx
}

func newAddress() sdk.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func getCoins(ck CoinKeeper, ctx sdk.Context, addr crypto.Address) (sdk.Coins, sdk.Error) {
	zero := sdk.Coins{}
	return ck.AddCoins(ctx, addr, zero)
}

// custom tx codec
// TODO: use new go-wire
func makeCodec() *wire.Codec {
	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, accTypeApp},
	)
	cdc := wire.NewCodec()
	return cdc
}

func TestIBCHandler(t *testing.T) {
	key := sdk.NewKVStoreKey("bank")
	ctx := defaultContext(key)
	am := auth.NewAccountMapper(key, &auth.BaseAccount{})
	ck := NewCoinKeeper(am)
	h := NewIBCHandler(ck)
	var _ = makeCodec()

	addr := newAddress()
	mycoins := sdk.Coins{{"atom", 10}}

	payload := SendPayload{
		SrcAddr:  addr,
		DestAddr: addr,
		Coins:    mycoins,
	}

	coins, err := getCoins(ck, ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, sdk.Coins{}, coins)

	res := h(ctx, payload)
	assert.True(t, res.IsOK())

	coins, err = getCoins(ck, ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, mycoins, coins)
}
