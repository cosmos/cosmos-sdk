package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"reflect"
)

// Handle all "bank" type messages.
func NewHandler(accStore sdk.AccountStore) sdk.Handler {

	return func(ctx sdk.Context, tx sdk.Tx) sdk.Result {
		cs := CoinStore{accStore}
		msg := tx.(sdk.Msg)
		switch msg := msg.(type) {
		case SendMsg:
			return handleSendMsg(ctx, cs, msg)
		case IssueMsg:
			return handleIssueMsg(ctx, cs, msg)
		default:
			return sdk.Result{
				Code: 1, // TODO
				Log:  "Unrecognized bank Tx type: " + reflect.TypeOf(tx).Name(),
			}
		}
	}

}

// Handle SendMsg.
func handleSendMsg(ctx sdk.Context, cs CoinStore, msg SendMsg) sdk.Result {
	// NOTE: totalIn == totalOut should already have been checked

	for _, in := range msg.Inputs {
		_, err := cs.SubtractCoins(ctx, in.Address, in.Coins)
		if err != nil {
			return sdk.Result{
				Code: 1, // TODO
			}
		}
	}

	for _, out := range msg.Outputs {
		_, err := cs.AddCoins(ctx, out.Address, out.Coins)
		if err != nil {
			return sdk.Result{
				Code: 1, // TODO
			}
		}
	}

	return sdk.Result{} // TODO
}

// Handle IssueMsg.
func handleIssueMsg(ctx sdk.Context, cs CoinStore, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}
