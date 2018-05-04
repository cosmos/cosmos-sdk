package bank

import (
	//	"testing"

	//	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"

	dbm "github.com/tendermint/tmlibs/db"

	oldwire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	//	"github.com/cosmos/cosmos-sdk/x/ibc"
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

/*
func TestHandler(t *testing.T) {
	cdc := makeCodec()
	key := sdk.NewKVStoreKey("bank")
	am := auth.NewAccountMapper(key, &auth.BaseAccount{})
	ck := NewCoinKeeper(am)
	ibckey := sdk.NewKVStoreKey("ibc")
	ibcf := ibc.NewKeeperFactory(cdc, ibckey)
	h := NewHandler(ck, ibcf.Port("bank"))

	ctx := defaultContext(key, ibckey)

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
	assert.Nil(t, res)

	coins, err = getCoins(ck, ctx, addr)
	assert.Nil(t, err)
	assert.Equal(t, mycoins, coins)
}
*/
