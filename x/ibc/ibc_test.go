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

// custom tx codec
// TODO: use new go-wire
func makeCodec() *wire.Codec {
	const msgTypeRemoteSave = 0x1
	const msgTypeReceive = 0x2
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Msg }{},
		oldwire.ConcreteType{remoteSaveMsg{}, msgTypeRemoteSave},
		oldwire.ConcreteType{ReceiveMsg{}, msgTypeReceive},
	)

	const accTypeApp = 0x1
	var _ = oldwire.RegisterInterface(
		struct{ sdk.Account }{},
		oldwire.ConcreteType{&auth.BaseAccount{}, accTypeApp},
	)
	cdc := wire.NewCodec()
	return cdc
}

func remoteSaveHandler(ibck Keeper, key sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case remoteSaveMsg:
			return handleRemoteSaveMsg(ctx, ibck, msg)
		case ReceiveMsg:
			return ibck.Handle(func(ctx sdk.Context, p Payload) sdk.Error {
				switch p := p.(type) {
				case remoteSavePayload:
					return handleRemoteSavePayload(ctx, key, p)
				default:
					return sdk.ErrUnknownRequest("")
				}
			}, ctx, msg)
		default:
			return sdk.ErrUnknownRequest("").Result()
		}
	}
}

func handleRemoteSaveMsg(ctx sdk.Context, ibck Keeper, msg remoteSaveMsg) sdk.Result {
	ibck.Send(ctx, msg.payload, msg.destChain)
	return sdk.Result{}
}

func handleRemoteSavePayload(ctx sdk.Context, key sdk.StoreKey, p remoteSavePayload) sdk.Error {
	store := ctx.KVStore(key)
	store.Set(p.key, p.value)
	return nil
}

func TestIBC(t *testing.T) {
	cdc := wire.NewCodec()

	key := sdk.NewKVStoreKey("ibc")
	rskey := sdk.NewKVStoreKey("remote")
	ctx := defaultContext(key, rskey)
	chainid := ctx.ChainID()

	factory := NewKeeperFactory(cdc, key)
	rsh := remoteSaveHandler(factory.Port("remote"), key)

	payload := remoteSavePayload{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	saveMsg := remoteSaveMsg{
		payload:   payload,
		destChain: chainid,
	}
	/*
		packet := Packet{
			Payload:   payload,
			SrcChain:  chainid,
			DestChain: chainid,
		}

			receiveMsg := ReceiveMsg{
				Packet:   packet,
				Relayer:  newAddress(),
				Sequence: 0,
			}
	*/
	var res sdk.Result

	res = rsh(ctx, saveMsg)
	assert.True(t, res.IsOK())
	/*
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
	*/
}
