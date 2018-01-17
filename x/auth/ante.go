package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(store sdk.AccountStore) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// Deduct the fee from the fee payer.
		// This is done first because it only
		// requires fetching 1 account.
		payerAddr := tx.GetFeePayer()
		payerAcc := store.GetAccount(ctx, payerAddr)
		if payerAcc == nil {
			return ctx, sdk.Result{
				Code: 1, // TODO
			}, true
		}

		// payerAcc.Subtract ?

		// Ensure that signatures are correct.
		var signerAddrs = tx.GetSigners()
		var signerAccs = make([]sdk.Account, len(signerAddrs))
		var signatures = tx.GetSignatures()

		// Assert that there are signers.
		if len(signatures) == 0 {
			return ctx, sdk.Result{
				Code: 1, // TODO
			}, true
		}
		if len(signatures) != len(signerAddrs) {
			return ctx, sdk.Result{
				Code: 1, // TODO
			}, true
		}

		// Check each nonce and sig.
		for i, sig := range signatures {

			var signerAcc = store.GetAccount(ctx, signerAddrs[i])
			signerAccs[i] = signerAcc

			// If no pubkey, set pubkey.
			if signerAcc.GetPubKey() == nil {
				err := signerAcc.SetPubKey(sig.PubKey)
				if err != nil {
					return ctx, sdk.Result{
						Code: 1, // TODO
					}, true
				}
			}

			// Check and increment sequence number.
			seq := signerAcc.GetSequence()
			if seq != sig.Sequence {
				return ctx, sdk.Result{
					Code: 1, // TODO
				}, true
			}
			signerAcc.SetSequence(seq + 1)

			// Check sig.
			if !sig.PubKey.VerifyBytes(tx.GetSignBytes(), sig.Signature) {
				return ctx, sdk.Result{
					Code: 1, // TODO
				}, true
			}

			// Save the account.
			store.SetAccount(ctx, signerAcc)
		}

		ctx = WithSigners(ctx, signerAccs)
		return ctx, sdk.Result{}, false // continue...
	}
}
