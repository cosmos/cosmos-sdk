package ibc

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
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

type remoteSaveFailPayload struct {
	remoteSavePayload
}

func (p remoteSaveFailPayload) Type() string {
	return "remote"
}

func (p remoteSaveFailPayload) ValidateBasic() sdk.Error {
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

func makeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(remoteSaveMsg{}, "test/remoteSave", nil)
	cdc.RegisterConcrete(ReceiveMsg{}, "test/Receive", nil)

	// Register AppAccount
	cdc.RegisterInterface((*sdk.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "test/Account", nil)
	wire.RegisterCrypto(cdc)

	// Register Payloads
	cdc.RegisterInterface((*Payload)(nil), nil)
	cdc.RegisterConcrete(remoteSavePayload{}, "test/payload/remoteSave", nil)
	cdc.RegisterConcrete(remoteSaveFailPayload{}, "test/payload/remoteSaveFail", nil)

	return cdc

}

func remoteSaveHandler(ibcc Channel, key sdk.StoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case remoteSaveMsg:
			return handleRemoteSaveMsg(ctx, ibcc, msg)
		case ReceiveMsg:
			return ibcc.Receive(func(ctx sdk.Context, p Payload) (Payload, sdk.Error) {
				switch p := p.(type) {
				case remoteSavePayload:
					return handleRemoteSavePayload(ctx, key, p)
				default:
					return nil, sdk.ErrUnknownRequest("")
				}
			}, ctx, msg)
		case ReceiptMsg:
			return ibcc.Receipt(func(ctx sdk.Context, p Payload) {
				switch p := p.(type) {
				case remoteSaveFailPayload:
					handleRemoteSaveFailPayload(ctx, key, p)
				default:
					sdk.ErrUnknownRequest("")
				}
			}, ctx, msg)

		default:
			return sdk.ErrUnknownRequest("").Result()
		}
	}
}

func handleRemoteSaveMsg(ctx sdk.Context, ibcc Channel, msg remoteSaveMsg) sdk.Result {
	ibcc.Send(ctx, msg.payload, msg.destChain)
	return sdk.Result{}
}

func handleRemoteSavePayload(ctx sdk.Context, key sdk.StoreKey, p remoteSavePayload) (Payload, sdk.Error) {
	store := ctx.KVStore(key)
	if store.Has(p.key) {
		return remoteSaveFailPayload{p}, sdk.NewError(1000, "Key already exists")
	}
	store.Set(p.key, p.value)
	return nil, nil
}

func handleRemoteSaveFailPayload(ctx sdk.Context, key sdk.StoreKey, p remoteSaveFailPayload) {
	return
}

func TestIBC(t *testing.T) {
	cdc := makeCodec()

	ibckey := sdk.NewKVStoreKey("ibc")
	key := sdk.NewKVStoreKey("remote")
	ctx := defaultContext(ibckey, key)
	chainid := ctx.ChainID()

	keeper := NewKeeper(cdc, key)
	ibch := NewHandler(keeper)
	h := remoteSaveHandler(keeper.Channel("remote"), key)

	payload := remoteSavePayload{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	saveMsg := remoteSaveMsg{
		payload:   payload,
		destChain: chainid,
	}

	var res sdk.Result

	res = h(ctx, saveMsg)
	assert.True(t, res.IsOK())

	packet := Packet{
		Payload:   payload,
		SrcChain:  chainid,
		DestChain: chainid,
	}

	receiveMsg := ReceiveMsg{
		Packet: packet,
		PacketProof: PacketProof{
			Sequence: 0,
		},
		Relayer: newAddress(),
	}

	res = h(ctx, receiveMsg)
	assert.True(t, res.IsOK())

	store := ctx.KVStore(key)
	val := store.Get(payload.key)
	assert.Equal(t, payload.value, val)

	res = h(ctx, receiveMsg)
	assert.False(t, res.IsOK())

	res = h(ctx, saveMsg)
	assert.True(t, res.IsOK())

	receiveMsg = ReceiveMsg{
		Packet: packet,
		PacketProof: PacketProof{
			Sequence: 1,
		},
		Relayer: newAddress(),
	}

	res = h(ctx, receiveMsg)
	assert.True(t, res.IsOK())

	packet.Payload = remoteSaveFailPayload{payload}

	receiptMsg := ReceiptMsg{
		Packet: packet,
		PacketProof: PacketProof{
			Sequence: 0,
		},
		Relayer: newAddress(),
	}

	res = h(ctx, receiptMsg)
	assert.True(t, res.IsOK())

	receiveCleanupMsg := ReceiveCleanupMsg{
		Sequence:     2,
		SrcChain:     chainid,
		CleanupProof: CleanupProof{},
		Cleaner:      newAddress(),
	}

	res = ibch(ctx, receiveCleanupMsg)
	assert.True(t, res.IsOK())

	receiptCleanupMsg := ReceiptCleanupMsg{
		Sequence:     1,
		SrcChain:     chainid,
		CleanupProof: CleanupProof{},
		Cleaner:      newAddress(),
	}

	res = ibch(ctx, receiptCleanupMsg)
	assert.True(t, res.IsOK())

	/*
		unknownMsg := sdk.NewTestMsg(newAddress())
		res = ibch(ctx, unknownMsg)
		assert.False(t, res.IsOK())
	*/
}
