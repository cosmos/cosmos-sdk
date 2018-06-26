package app

import (
	"encoding/json"
	"reflect"

	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	bapp "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	app1Name = "App1"
)

func NewApp1(logger log.Logger, db dbm.DB) *bapp.BaseApp {

	// TODO: make this an interface or pass in
	// a TxDecoder instead.
	cdc := wire.NewCodec()

	// Create the base application object.
	app := bapp.NewBaseApp(app1Name, cdc, logger, db)

	// Create a key for accessing the account store.
	keyAccount := sdk.NewKVStoreKey("acc")

	// Determine how transactions are decoded.
	app.SetTxDecoder(txDecoder)

	// Register message routes.
	// Note the handler gets access to the account store.
	app.Router().
		AddRoute("bank", NewApp1Handler(keyAccount))

	// Mount stores and load the latest state.
	app.MountStoresIAVL(keyAccount)
	err := app.LoadLatestVersion(keyAccount)
	if err != nil {
		cmn.Exit(err.Error())
	}
	return app
}

//------------------------------------------------------------------
// Msg

// MsgSend implements sdk.Msg
var _ sdk.Msg = MsgSend{}

// MsgSend to send coins from Input to Output
type MsgSend struct {
	From   sdk.Address `json:"from"`
	To     sdk.Address `json:"to"`
	Amount sdk.Coins   `json:"amount"`
}

// NewMsgSend
func NewMsgSend(from, to sdk.Address, amt sdk.Coins) MsgSend {
	return MsgSend{from, to, amt}
}

// Implements Msg.
func (msg MsgSend) Type() string { return "bank" }

// Implements Msg. Ensure the addresses are good and the
// amount is positive.
func (msg MsgSend) ValidateBasic() sdk.Error {
	if len(msg.From) == 0 {
		return sdk.ErrInvalidAddress("From address is empty")
	}
	if len(msg.To) == 0 {
		return sdk.ErrInvalidAddress("To address is empty")
	}
	if !msg.Amount.IsPositive() {
		return sdk.ErrInvalidCoins("Amount is not positive")
	}
	return nil
}

// Implements Msg. JSON encode the message.
func (msg MsgSend) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// Implements Msg. Return the signer.
func (msg MsgSend) GetSigners() []sdk.Address {
	return []sdk.Address{msg.From}
}

//------------------------------------------------------------------
// Handler for the message

func NewApp1Handler(keyAcc *sdk.KVStoreKey) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSend:
			return handleMsgSend(ctx, keyAcc, msg)
		default:
			errMsg := "Unrecognized bank Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSend.
func handleMsgSend(ctx sdk.Context, key *sdk.KVStoreKey, msg MsgSend) sdk.Result {
	// NOTE: from, to, and amount were already validated

	store := ctx.KVStore(key)
	bz := store.Get(msg.From)
	if bz == nil {
		// TODO
	}

	var acc acc
	err := json.Unmarshal(bz, &acc)
	if err != nil {
		// InternalError
	}

	// TODO: finish the logic

	return sdk.Result{
	// TODO: Tags
	}
}

type acc struct {
	Coins sdk.Coins `json:"coins"`
}

//------------------------------------------------------------------
// Tx

// Simple tx to wrap the Msg.
type app1Tx struct {
	MsgSend
}

// This tx only has one Msg.
func (tx app1Tx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx.MsgSend}
}

// TODO: remove the need for this
func (tx app1Tx) GetMemo() string {
	return ""
}

// JSON decode MsgSend.
func txDecoder(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx app1Tx
	err := json.Unmarshal(txBytes, &tx)
	if err != nil {
		return nil, sdk.ErrTxDecode(err.Error())
	}
	return tx, nil
}
