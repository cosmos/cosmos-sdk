package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(accStore sdk.AccountStore) sdk.Handler {

	return func(ctx sdk.Context, tx sdk.Tx) sdk.Result {
		cs := CoinStore{accStore}

		sendTx, ok := tx.(sdk.Msg).(SendMsg)
		if !ok {
			panic("tx is not SendTx") // ?
		}

		// NOTE: totalIn == totalOut should already have been checked

		for _, in := range sendTx.Inputs {
			_, err := cs.SubtractCoins(ctx, in.Address, in.Coins)
			if err != nil {
				return sdk.Result{
					Code: 1, // TODO
				}
			}
		}

		for _, out := range sendTx.Outputs {
			_, err := cs.AddCoins(ctx, out.Address, out.Coins)
			if err != nil {
				return sdk.Result{
					Code: 1, // TODO
				}
			}
		}

		return sdk.Result{} // TODO
	}
}
