package auth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(accountMapper sdk.AccountMapper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (_ sdk.Context, _ sdk.Result, abort bool) {

		// Check sequence number based on the fee payer.
		// Deduct the fee from the fee payer.
		// This is done first because it only
		// requires fetching 1 account.
		payerAddr := tx.GetFeePayer()
		if payerAddr != nil {
			payerAcc := accountMapper.GetAccount(ctx, payerAddr)
			if payerAcc == nil {
				return ctx,
					sdk.ErrUnrecognizedAddress(payerAddr).Result(),
					true
			}

			// Check the nonces on the fee payer
			accNonce := payerAcc.GetAccNonce()
			txNonce := payerAcc.GetTxNonce()

			if accNonce != tx.GetAccNonce() {
				return ctx,
					sdk.ErrBadNonce(fmt.Sprintf("Got %d, expected %d", tx.GetAccNonce(), accNonce)).Result(),
					true
			}

			if txNonce != tx.GetAccNonce() {
				return ctx,
					sdk.ErrBadNonce(fmt.Sprintf("Got %d, expected %d", tx.GetTxNonce(), txNonce)).Result(),
					true
			}

			// Increment Tx Nonce
			payerAcc.SetTxNonce(txNonce + 1)

			// TODO: Charge fee from payerAcc.

			accountMapper.SetAccount(ctx, payerAcc)
		} else {
			// XXX:
			// TODO: Ensure that some other spam and replay prevention is used.
			// XXX:
		}

		var sigs = tx.GetSignatures()

		// Assert that there are signatures.
		if len(sigs) == 0 {
			return ctx,
				sdk.ErrUnauthorized("no signers").Result(),
				true
		}

		// Ensure that sigs are correct.
		var msg = tx.GetMsg()
		var signerAddrs = msg.GetSigners()
		var signerAccs = make([]sdk.Account, len(signerAddrs))

		// Assert that number of signatures is correct.
		if len(sigs) != len(signerAddrs) {
			return ctx,
				sdk.ErrUnauthorized("wrong number of signers").Result(),
				true
		}

		// Check each nonce and sig.
		// TODO Refactor out.
		for i, sig := range sigs {

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

			// Check sig.
			if !sig.PubKey.VerifyBytes(msg.GetSignBytes(ctx), sig.Signature) {
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
