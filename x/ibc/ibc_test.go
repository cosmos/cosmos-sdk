package ibc

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/lite"
	tmtypes "github.com/tendermint/tendermint/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var testCodespace = sdk.CodespaceUndefined

// AccountMapper(/Keeper) and IBCMapper should use different StoreKey later

func defaultContext(keys ...sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	for _, key := range keys {
		cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	}
	cms.LoadLatestVersion()
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

func newAddress() sdk.AccAddress {
	return sdk.AccAddress(crypto.GenPrivKeyEd25519().PubKey().Address())
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

func (p remoteSavePayload) DatagramType() DatagramType {
	return PacketType
}

func (p remoteSavePayload) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
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

func (p remoteSaveFailPayload) DatagramType() DatagramType {
	return ReceiptType
}

func (p remoteSaveFailPayload) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

func makeCodec() *wire.Codec {
	var cdc = wire.NewCodec()

	// Register Msgs
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(MsgSend{}, "test/ibc/Send", nil)
	cdc.RegisterConcrete(MsgReceive{}, "test/ibc/Receive", nil)

	// Register Payloads
	cdc.RegisterInterface((*Payload)(nil), nil)
	cdc.RegisterConcrete(remoteSavePayload{}, "test/payload/remoteSave", nil)
	cdc.RegisterConcrete(remoteSaveFailPayload{}, "test/payload/remoteSaveFail", nil)

	cdc.Seal()

	return cdc

}

func newIBCTestApp(logger log.Logger, db dbm.DB) *bam.BaseApp {
	cdc := makeCodec()
	app := bam.NewBaseApp("test", cdc, logger, db)

	key := sdk.NewKVStoreKey("remote")
	ibcKey := sdk.NewKVStoreKey("ibc")
	keeper := NewKeeper(cdc, ibcKey, app.RegisterCodespace(DefaultCodespace))

	app.Router().
		AddRoute("remote", remoteSaveHandler(key, keeper)).
		AddRoute("ibc", NewHandler(keeper))

	app.MountStoresIAVL(key, ibcKey)
	err := app.LoadLatestVersion(key)
	if err != nil {
		panic(err)
	}

	app.InitChain(abci.RequestInitChain{})
	return app
}

func remoteSaveHandler(key sdk.StoreKey, ibck Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ibcc := ibck.Channel(sdk.NewPrefixStoreGetter(key, []byte("ibctest")))
		switch msg := msg.(type) {
		case MsgSend:
			return ibcc.Send(func(p Payload) sdk.Result {
				switch p := p.(type) {
				case remoteSavePayload:
					return handleRemoteSavePayloadSend(p)
				default:
					return sdk.ErrUnknownRequest("").Result()
				}
			}, ctx, msg)
		case MsgReceive:
			return ibcc.Receive(func(ctx sdk.Context, p Payload) (Payload, sdk.Result) {
				switch p := p.(type) {
				case remoteSavePayload:
					return handleRemoteSavePayloadReceive(ctx, key, p)
				case remoteSaveFailPayload:
					return handleRemoteSaveFailPayloadReceive(ctx, key, p)
				default:
					return nil, sdk.ErrUnknownRequest("").Result()
				}
			}, ctx, msg)
			/*
				case MsgCleanup:
					return ibcc.Cleanup()
			*/
		default:
			return sdk.ErrUnknownRequest("").Result()
		}
	}
}

func handleRemoteSavePayloadSend(p Payload) sdk.Result {
	return sdk.Result{}
}

func handleRemoteSavePayloadReceive(ctx sdk.Context, key sdk.StoreKey, p remoteSavePayload) (Payload, sdk.Result) {
	store := ctx.KVStore(key)
	if store.Has(p.key) {
		return remoteSaveFailPayload{p}, sdk.NewError(testCodespace, 1000, "Key already exists").Result()
	}
	store.Set(p.key, p.value)
	return nil, sdk.Result{}
}

func handleRemoteSaveFailPayloadReceive(ctx sdk.Context, key sdk.StoreKey, p remoteSaveFailPayload) (Payload, sdk.Result) {
	return nil, sdk.Result{}
}

func TestIBC(t *testing.T) {
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app")
	db := dbm.NewMemDB()
	app := newIBCTestApp(logger, db)

	ctx := app.NewContext(false, abci.Header{})
	chainid := ctx.ChainID()

	var res sdk.Result

	// Open connection
	openConnMsg := MsgOpenConnection{
		ROT: lite.FullCommit{
			Commit: lite.Commit{
				Header: &tmtypes.Header{
					Height: 1,
				},
			},
		},
	}

	tx := auth.NewStdTx([]sdk.Msg{openConnMsg}, auth.NewStdFee(0), []auth.StdSignature{}, "")

	res = app.Deliver(tx)
	require.True(t, res.IsOK(), "%+v", res)

	// Open channel

	// Send IBC message
	payload := remoteSavePayload{
		key:   []byte("hello"),
		value: []byte("world"),
	}

	saveMsg := MsgSend{
		Payload:   payload,
		DestChain: chainid,
	}

	tx.Msgs[0] = saveMsg

	res = app.Deliver(tx)
	require.True(t, res.IsOK(), "%+v", res)

	// Receive IBC message
	data := Datagram{
		Header: Header{
			SrcChain:  chainid,
			DestChain: chainid,
		},
		Payload: payload,
	}

	receiveMsg := MsgReceive{
		Datagram: data,
		/*	PacketProof: PacketProof{
			Sequence: 0,
		},*/
		Relayer: newAddress(),
	}

	tx.Msgs[0] = receiveMsg

	res = app.Deliver(tx)
	require.True(t, res.IsOK(), "%+v\n", res)
	/*
		store := ctx.KVStore(key)
		val := store.Get(payload.key)
		require.Equal(t, payload.value, val)
	*/

	tx.Msgs[0] = receiveMsg
	res = app.Deliver(tx)
	require.False(t, UnwrapResult(res).IsOK(), "%+v\n", res)

	// Send another IBC message and receive it
	// It has duplicated key bytes so fails
	tx.Msgs[0] = saveMsg
	res = app.Deliver(tx)
	require.True(t, res.IsOK())

	receiveMsg = MsgReceive{
		Datagram: data,
		/*PacketProof: PacketProof{
			Sequence: 1,
		},*/
		Relayer: newAddress(),
	}

	tx.Msgs[0] = receiveMsg
	res = app.Deliver(tx)
	require.False(t, UnwrapResult(res).IsOK())

	// Return fail receipt
	data.Payload = remoteSaveFailPayload{payload}

	receiptMsg := MsgReceive{
		Datagram: data,
		/*PacketProof: PacketProof{
			Sequence: 0,
		},*/
		Relayer: newAddress(),
	}

	tx.Msgs[0] = receiptMsg
	res = app.Deliver(tx)
	require.True(t, res.IsOK())
	/*
		// Cleanup receive queue
		receiveCleanupMsg := MsgReceiveCleanup{
			Sequence: 2,
			SrcChain: chainid,
			//CleanupProof: CleanupProof{},
			Cleaner: newAddress(),
		}

		tx.Msgs[0] = receiveCleanupMsg
		res = app.Deliver(tx)
		require.True(t, res.IsOK(), "%+v", res)

		// Cleanup receipt queue
		receiptCleanupMsg := MsgReceiptCleanup{
			Sequence: 1,
			SrcChain: chainid,
			//CleanupProof: CleanupProof{},
			Cleaner: newAddress(),
		}

		tx.Msgs[0] = receiptCleanupMsg
		res = app.Deliver(tx)
		require.True(t, UnwrapResult(res).IsOK())

		unknownMsg := sdk.NewTestMsg(newAddress())
		tx.Msgs[0] = unknownMsg
		res = app.Deliver(tx)
		require.False(t, UnwrapResult(res).IsOK())
	*/
}
