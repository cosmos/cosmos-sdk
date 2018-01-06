package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/store"
)

func DecoratorFn(newAccountStore func(types.KVStore) store.AccountStore) types.Decorator {
	return func(ctx types.Context, ms types.MultiStore, tx types.Tx, next types.Handler) types.Result {

		accountStore := newAccountStore(ms.GetKVStore("main"))

		signers := tx.Signers()
		signatures := tx.Signatures()

		// assert len
		if len(signatures) == 0 {
			return types.Result{
				Code: 1, // TODO
			}
		}
		if len(signatures) != len(signers) {
			return types.Result{
				Code: 1, // TODO
			}
		}

		// check each nonce and sig
		for i, sig := range signatures {

			// get account
			_acc := accountStore.GetAccount(signers[i])

			// assert it has the right methods
			acc, ok := _acc.(Auther)
			if !ok {
				return types.Result{
					Code: 1, // TODO
				}
			}

			// if no pubkey, set pubkey
			if acc.GetPubKey().Empty() {
				err := acc.SetPubKey(sig.PubKey)
				if err != nil {
					return types.Result{
						Code: 1, // TODO
					}
				}
			}

			// check sequence number
			seq := acc.GetSequence()
			if seq != sig.Sequence {
				return types.Result{
					Code: 1, // TODO
				}
			}

			// check sig
			if !sig.PubKey.VerifyBytes(tx.SignBytes(), sig.Signature) {
				return types.Result{
					Code: 1, // TODO
				}
			}
		}
		return next(ctx, ms, tx)
	}
}
