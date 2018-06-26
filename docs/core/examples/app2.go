package app

import (
	"reflect"

	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	app2Name = "App2"
)

func NewCodec() *wire.Codec {
	// TODO register
	return nil
}

func NewApp2(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	cdc := NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app2Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")
	keyIssuer := sdk.NewKVStoreKey("issuer")

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp2Handler(keyAccount, keyIssuer))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount, keyIssuer)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

//------------------------------------------------------------------
// Msgs

// TODO: MsgIssue

//------------------------------------------------------------------
// Handler for the message

func NewApp1Handler(keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, keyAcc, msg)
		case MsgIssue:
			// TODO
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

//------------------------------------------------------------------
// Tx

// Simple tx to wrap the Msg.
type app2Tx struct {
	sdk.Msg
}

// This tx only has one Msg.
func (tx app2Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.Msg}
}

// TODO: remove the need for this
func (tx app2Tx) GetMemo() string {
	return ""
}
