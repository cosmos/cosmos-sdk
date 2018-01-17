package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handles "bank" type messages.
func NewHandler(accStore sdk.AccountStore) sdk.Handler {

	return func(ctx sdk.Context, tx sdk.Tx) sdk.Result {
		cs := CoinStore{accStore}
		msg := tx.Msg
		switch msg := msg.(type) {
		case SendMsg:
			handleSendMsg(ctx, msg)
		case IssueMsg:
			handleIssueMsg(ctx, msg)
		}
	}
}

func handleSendMsg(ctx sdk.Context, msg SendMsg) sdk.Result {

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

func handleIssueMsg(ctx sdk.Context, msg IssueMsg) sdk.Result {
	panic("not implemented yet")
}
