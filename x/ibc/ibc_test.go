package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"

	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wire "github.com/cosmos/cosmos-sdk/wire"

	ibc "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// AccountMapper(/CoinKeeper) and IBCMapper should use different StoreKey later

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

func newAddress() crypto.Address {
	return crypto.GenPrivKeyEd25519().PubKey().Address()
}

type remoteSavePayload struct {
	key   []byte
	value []byte
}

func (p remoteSavePayload) Type() string {
	return "remote"
}

func (p remoteSavePayload) ValidateBasic() sdk.Error {
	return nil
}

type remoteSaveMsg struct {
	payload   remoteSavePayload
	destChain string
}

func (msg remoteSaveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg remoteSaveMsg) GetSignBytes() []byte {
	return nil
}

func (msg remoteSaveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{}
}

func (msg remoteSaveMsg) Type() string {
	return "remote"
}

func (msg remoteSaveMsg) ValidateBasic() sdk.Error {
	return nil
}

func remoteSaveHandler(sender ibc.Sender) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		rsmsg := msg.(remoteSaveMsg)
		sender.Push(ctx, rsmsg.payload, rsmsg.destChain)
		return sdk.Result{}
	}
}

func remoteSaveIBCHandler(key sdk.StoreKey) ibc.Handler {
	return func(ctx sdk.Context, p ibc.Payload) sdk.Error {
		rsp := p.(remoteSavePayload)
		store := ctx.KVStore(key)
		store.Set(rsp.key, rsp.value)
		return nil
	}
}

func TestIBC(t *testing.T) {
	cdc := wire.NewCodec()

	key := sdk.NewKVStoreKey("ibc")
	rskey := sdk.NewKVStoreKey("remote")
	ctx := defaultContext(key, rskey)
	chainid := ctx.ChainID()

	keeper := ibc.NewKeeper(cdc, key)
	keeper.Dispatcher().
		AddDispatch("remote", remoteSaveIBCHandler(rskey))

	rsh := remoteSaveHandler(keeper.Sender())
	ibch := NewHandler(keeper)

	payload := remoteSavePayload{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	saveMsg := remoteSaveMsg{
		payload:   payload,
		destChain: chainid,
	}

	packet := ibc.Packet{
		Payload:   payload,
		SrcChain:  chainid,
		DestChain: chainid,
	}

	receiveMsg := ReceiveMsg{
		Packet:   packet,
		Relayer:  newAddress(),
		Sequence: 0,
	}

	var res sdk.Result

	res = rsh(ctx, saveMsg)
	assert.True(t, res.IsOK())

	res = ibch(ctx, receiveMsg)
	assert.True(t, res.IsOK())

	store := ctx.KVStore(rskey)
	val := store.Get(payload.key)
	assert.Equal(t, payload.value, val)

	res = ibch(ctx, receiveMsg)
	assert.False(t, res.IsOK())

	unknownMsg := sdk.NewTestMsg(newAddress())
	res = ibch(ctx, unknownMsg)
	assert.False(t, res.IsOK())
}
