package auth

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(accountMapper sdk.AccountMapper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// Deduct the fee from the fee payer.
		// This is done first because it only
		// requires fetching 1 account.
		payerAddr := tx.GetFeePayer()
		if payerAddr != nil {
			payerAcc := accountMapper.GetAccount(ctx, payerAddr)
			if payerAcc == nil {
				return ctx,
					sdk.ErrUnrecognizedAddress("").Result(),
					true
			}
			// TODO: Charge fee from payerAcc.
			// TODO: accountMapper.SetAccount(ctx, payerAddr)
		} else {
			// TODO: Ensure that some other spam prevention is used.
			// NOTE: bam.TestApp.RunDeliverMsg/RunCheckMsg will
			// create a Tx with no payer.
		}

		// Ensure that signatures are correct.
		var signerAddrs = tx.GetSigners()
		var signerAccs = make([]sdk.Account, len(signerAddrs))
		var signatures = tx.GetSignatures()

		// Assert that there are signers.
		if len(signerAddrs) == 0 {
			if !bam.IsTestAppTx(tx) {
				return ctx,
					sdk.ErrUnauthorized("no signers").Result(),
					true
			}
		}

		// Assert that number of signatures is correct.
		if len(signatures) != len(signerAddrs) {
			return ctx,
				sdk.ErrUnauthorized("wrong number of signers").Result(),
				true
		}

		// Check each nonce and sig.
		for i, sig := range signatures {

			var signerAcc = accountMapper.GetAccount(ctx, signerAddrs[i])
			signerAccs[i] = signerAcc

			// If no pubkey, set pubkey.
			if signerAcc.GetPubKey() == nil {
				err := signerAcc.SetPubKey(sig.PubKey)
				if err != nil {
					return ctx,
						sdk.ErrInternal("setting PubKey on signer").Result(),
						true
				}
			}

			// Check and increment sequence number.
			seq := signerAcc.GetSequence()
			if seq != sig.Sequence {
				return ctx,
					sdk.ErrInvalidSequence("").Result(),
					true
			}
			signerAcc.SetSequence(seq + 1)

			// Check sig.
			if !sig.PubKey.VerifyBytes(tx.GetSignBytes(), sig.Signature) {
				return ctx,
					sdk.ErrUnauthorized("").Result(),
					true
			}

			// Save the account.
			accountMapper.SetAccount(ctx, signerAcc)
		}

		ctx = WithSigners(ctx, signerAccs)
		return ctx, sdk.Result{}, false // continue...
	}
}
