package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(accountMapper sdk.AccountMapper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx,
	) (_ sdk.Context, _ sdk.Result, abort bool) {

		// Assert that there are signatures.
		var sigs = tx.GetSignatures()
		if len(sigs) == 0 {
			return ctx,
				sdk.ErrUnauthorized("no signers").Result(),
				true
		}

		// TODO: can tx just implement message?
		msg := tx.GetMsg()

		// Assert that number of signatures is correct.
		var signerAddrs = msg.GetSigners()
		if len(sigs) != len(signerAddrs) {
			return ctx,
				sdk.ErrUnauthorized("wrong number of signers").Result(),
				true
		}

		// Collect accounts to set in the context
		var signerAccs = make([]sdk.Account, len(signerAddrs))

		signBytes := msg.GetSignBytes()

		// First sig is the fee payer.
		// Check sig and nonce, deduct fee.
		// This is done first because it only
		// requires fetching 1 account.
		payerAcc, res := processSig(ctx, accountMapper, signerAddrs[0], sigs[0], signBytes)
		if !res.IsOK() {
			return ctx, res, true
		}
		signerAccs[0] = payerAcc
		// TODO: Charge fee from payerAcc.
		// TODO: accountMapper.SetAccount(ctx, payerAddr)

		// Check sig and nonce for the rest.
		for i := 1; i < len(sigs); i++ {
			signerAcc, res := processSig(ctx, accountMapper, signerAddrs[i], sigs[i], signBytes)
			if !res.IsOK() {
				return ctx, res, true
			}
			signerAccs[i] = signerAcc
		}

		ctx = WithSigners(ctx, signerAccs)
		return ctx, sdk.Result{}, false // continue...
	}
}

// verify the signature and increment the sequence.
// if the account doesn't have a pubkey, set it as well.
func processSig(ctx sdk.Context, am sdk.AccountMapper, addr sdk.Address, sig sdk.StdSignature, signBytes []byte) (acc sdk.Account, res sdk.Result) {

	// Get the account
	acc = am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnrecognizedAddress(addr).Result()
	}

	// Check and increment sequence number.
	seq := acc.GetSequence()
	if seq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence("").Result()
	}
	acc.SetSequence(seq + 1)

	// Check and possibly set pubkey.
	pubKey := acc.GetPubKey()
	if pubKey.Empty() {
		pubKey = sig.PubKey
		err := acc.SetPubKey(pubKey)
		if err != nil {
			return nil, sdk.ErrInternal("setting PubKey on signer").Result()
		}
	}
	// TODO: should we enforce pubKey == sig.PubKey ?
	// If not, ppl can send useless PubKeys after first tx

	// Check sig.
	if !sig.PubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("").Result()
	}

	// Save the account.
	am.SetAccount(ctx, acc)
	return
}
