package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Set via `app.App.SetTxDecoder(app.decodeTx)`
func (app *BasecoinApp) decodeTx(txBytes []byte) (sdk.Tx, error) {
	var tx = sdk.StdTx{}
	err := app.cdc.UnmarshalBinary(txBytes, &tx)
	return tx, err
}

// Wire requires registration of interfaces & concrete types.
func (app *BasecoinApp) registerMsgs() {
	cdc := app.cdc

	// Register the Msg interface.
	cdc.RegisterInterface((*sdk.Msg)(nil), nil)
	cdc.RegisterConcrete(bank.SendMsg{}, "cosmos-sdk/SendMsg", nil)   // XXX refactor out
	cdc.RegisterConcrete(bank.IssueMsg{}, "cosmos-sdk/IssueMsg", nil) // XXX refactor out to bank/msgs.go
	// more msgs here...

	// All interfaces to be encoded/decoded in a Msg must be
	// registered here, along with all the concrete types that
	// implement them.
}
