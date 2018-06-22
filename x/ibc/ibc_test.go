package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// AccountMapper(/Keeper) and IBCMapper should use different StoreKey later

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

func newAddress() crypto.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func getCoins(ck bank.Keeper, ctx sdk.Context, addr crypto.Address) (sdk.Coins, sdk.Error) {
	zero := sdk.Coins(nil)
	coins, _, err := ck.AddCoins(ctx, addr, zero)
	return coins, err
}

func makeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.MsgSend{}, "test/ibc/Send", nil)
	cdc.RegisterConcrete(bank.MsgIssue{}, "test/ibc/Issue", nil)
	cdc.RegisterConcrete(IBCTransferMsg{}, "test/ibc/IBCTransferMsg", nil)
	cdc.RegisterConcrete(IBCReceiveMsg{}, "test/ibc/IBCReceiveMsg", nil)

	// Register AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/ibc/Account", nil)
	wire.RegisterCrypto(cdc)

	return cdc
}

func TestIBC(t *testing.T) {
	cdc := makeCodec()

	key := sdk.NewKVStoreKey("ibc")
	ctx := defaultContext(key)

	am := auth.NewAccountMapper(cdc, key, &auth.BaseAccount{})
	ck := bank.NewKeeper(am)

	src := newAddress()
	dest := newAddress()
	chainid := "ibcchain"
	zero := sdk.Coins(nil)
	mycoins := sdk.Coins{sdk.NewCoin("mycoin", 10)}

	coins, _, err := ck.AddCoins(ctx, src, mycoins)
	assert.Nil(t, err)
	assert.Equal(t, mycoins, coins)

	ibcm := NewMapper(cdc, key, DefaultCodespace)
	h := NewHandler(ibcm, ck)
	packet := IBCPacket{
		SrcAddr:   src,
		DestAddr:  dest,
		Coins:     mycoins,
		SrcChain:  chainid,
		DestChain: chainid,
	}

	store := ctx.KVStore(key)

	var msg sdk.Msg
	var res sdk.Result
	var egl int64
	var igs int64

	egl = ibcm.getEgressLength(store, chainid)
	assert.Equal(t, egl, int64(0))

	msg = IBCTransferMsg{
		IBCPacket: packet,
	}
	res = h(ctx, msg)
	assert.True(t, res.IsOK())

	coins, err = getCoins(ck, ctx, src)
	assert.Nil(t, err)
	assert.Equal(t, zero, coins)

	egl = ibcm.getEgressLength(store, chainid)
	assert.Equal(t, egl, int64(1))

	igs = ibcm.GetIngressSequence(ctx, chainid)
	assert.Equal(t, igs, int64(0))

	msg = IBCReceiveMsg{
		IBCPacket: packet,
		Relayer:   src,
		Sequence:  0,
	}
	res = h(ctx, msg)
	assert.True(t, res.IsOK())

	coins, err = getCoins(ck, ctx, dest)
	assert.Nil(t, err)
	assert.Equal(t, mycoins, coins)

	igs = ibcm.GetIngressSequence(ctx, chainid)
	assert.Equal(t, igs, int64(1))

	res = h(ctx, msg)
	assert.False(t, res.IsOK())

	igs = ibcm.GetIngressSequence(ctx, chainid)
	assert.Equal(t, igs, int64(1))
}
