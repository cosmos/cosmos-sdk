package sendtx

import (
	"github.com/cosmos/cosmos-sdk/types"
	coinstore "github.com/cosmos/cosmos-sdk/x/coinstore"
)

func TransferHandlerFn(newAccStore func(types.KVStore) types.AccountStore) types.Handler {
	return func(ctx types.Context, ms types.MultiStore, tx types.Tx) types.Result {

		accStore := newAccStore(ms.GetKVStore("main"))
		cs := coinstore.CoinStore{accStore}

		sendTx, ok := tx.(SendTx)
		if !ok {
			panic("tx is not SendTx") // ?
		}

		// NOTE: totalIn == totalOut should already have been checked

		for _, in := range sendTx.Inputs {
			_, err := cs.SubtractCoins(in.Address, in.Coins)
			if err != nil {
				return types.Result{
					Code: 1, // TODO
				}
			}
		}

		for _, out := range sendTx.Outputs {
			_, err := cs.AddCoins(out.Address, out.Coins)
			if err != nil {
				return types.Result{
					Code: 1, // TODO
				}
			}
		}

		return types.Result{} // TODO
	}
}
