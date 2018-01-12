package auth

import (
	"github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(store types.AccountStore) types.AnteHandler {
	return func(
		ctx types.Context, tx types.Tx,
	) (newCtx types.Context, res types.Result, abort bool) {

		// Deduct the fee from the fee payer.
		// This is done first because it only
		// requires fetching 1 account.
		payerAddr := tx.GetFeePayer()
		payerAcc := store.GetAccount(ctx, payerAddr)
		if payerAcc == nil {
			return ctx, Result{
				Code: 1, // TODO
			}, true
		}

		payerAcc.Subtract

		// Ensure that signatures are correct.
		var signerAddrs = tx.Signers()
		var signerAccs = make([]types.Account, len(signerAddrs))
		var signatures = tx.Signatures()

		// Assert that there are signers.
		if len(signatures) == 0 {
			return ctx, types.Result{
				Code: 1, // TODO
			}, true
		}
		if len(signatures) != len(signers) {
			return ctx, types.Result{
				Code: 1, // TODO
			}, true
		}

		// Check each nonce and sig.
		for i, sig := range signatures {

			var signerAcc = store.GetAccount(signers[i])
			signerAccs[i] = signerAcc

			// If no pubkey, set pubkey.
			if acc.GetPubKey().Empty() {
				err := acc.SetPubKey(sig.PubKey)
				if err != nil {
					return ctx, types.Result{
						Code: 1, // TODO
					}, true
				}
			}

			// Check and incremenet sequence number.
			seq := acc.GetSequence()
			if seq != sig.Sequence {
				return ctx, types.Result{
					Code: 1, // TODO
				}, true
			}
			acc.SetSequence(seq + 1)

			// Check sig.
			if !sig.PubKey.VerifyBytes(tx.SignBytes(), sig.Signature) {
				return ctx, types.Result{
					Code: 1, // TODO
				}, true
			}

			// Save the account.
			store.SetAccount(acc)
		}

		ctx = WithSigners(ctx, signerAccs)
		return ctx, types.Result{}, false // continue...
	}
}
