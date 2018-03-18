package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	oldwire "github.com/tendermint/go-wire"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/cool"
)

// AccountMapper(/CoinKeeper) and IBCMapper should use different StoreKey later

func defaultContext(key sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, nil)
	return ctx
}

func newAddress() crypto.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

func getCoins(ck bank.CoinKeeper, ctx sdk.Context, addr crypto.Address) (sdk.Coins, sdk.Error) {
	zero := sdk.Coins{}
	return ck.AddCoins(ctx, addr, zero)
}

// custom tx codec
// TODO: use new go-wire
func makeCodec() *wire.Codec {

	const msgTypeSend = 0x1
	const msgTypeIssue = 0x2
	const msgTypeQuiz = 0x3
	const msgTypeSetTrend = 0x4
	const msgTypeIBCTransferMsg = 0x5
	const msgTypeIBCReceiveMsg = 0x6
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{bank.SendMsg{}, msgTypeSend},
		oldwire.ConcreteType{bank.IssueMsg{}, msgTypeIssue},
		oldwire.ConcreteType{cool.QuizMsg{}, msgTypeQuiz},
		oldwire.ConcreteType{cool.SetTrendMsg{}, msgTypeSetTrend},
		oldwire.ConcreteType{IBCTransferMsg{}, msgTypeIBCTransferMsg},
		oldwire.ConcreteType{IBCReceiveMsg{}, msgTypeIBCReceiveMsg},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, accTypeApp},
	)
	cdc := wire.NewCodec()

	// cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	// bank.RegisterWire(cdc)   // Register bank.[SendMsg,IssueMsg] types.
	// crypto.RegisterWire(cdc) // Register crypto.[PubKey,PrivKey,Signature] types.
	// ibc.RegisterWire(cdc) // Register ibc.[IBCTransferMsg, IBCReceiveMsg] types.
	return cdc
}

func TestIBC(t *testing.T) {
	var _ = makeCodec()

	key := sdk.NewKVStoreKey("ibc")
	ctx := defaultContext(key)

	am := auth.NewAccountMapper(key, &auth.BaseAccount{})
	ck := bank.NewCoinKeeper(am)

	src := newAddress()
	dest := newAddress()
	chainid := "ibcchain"
	zero := sdk.Coins{}
	mycoins := sdk.Coins{sdk.Coin{"mycoin", 10}}

	coins, err := ck.AddCoins(ctx, src, mycoins)
	assert.Nil(t, err)
	assert.Equal(t, mycoins, coins)

	ibcm := NewIBCMapper(key)
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
