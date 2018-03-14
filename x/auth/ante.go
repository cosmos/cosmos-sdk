package auth

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewAnteHandler returns an AnteHandler that checks
// and increments sequence numbers, checks signatures,
// and deducts fees from the first signer.
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

		// Get the sign bytes (requires all sequence numbers)
		sequences := make([]int64, len(signerAddrs))
		for i := 0; i < len(signerAddrs); i++ {
			sequences[i] = sigs[i].Sequence
		}
		signBytes := sdk.StdSignBytes(ctx.ChainID(), sequences, msg)

		// Check sig and nonce and collect signer accounts.
		var signerAccs = make([]sdk.Account, len(signerAddrs))
		for i := 0; i < len(sigs); i++ {
			isFeePayer := i == 0 // first sig pays the fees

			signerAddr, sig := signerAddrs[i], sigs[i]
			signerAcc, res := processSig(ctx, accountMapper, signerAddr, sig, signBytes, isFeePayer)
			if !res.IsOK() {
				return ctx, res, true
			}
			signerAccs[i] = signerAcc
		}

		ctx = WithSigners(ctx, signerAccs)
		// TODO: tx tags (?)
		return ctx, sdk.Result{}, false // continue...
	}
}

// verify the signature and increment the sequence.
// if the account doesn't have a pubkey, set it.
// deduct fee from fee payer.
func processSig(ctx sdk.Context, am sdk.AccountMapper,
	addr sdk.Address, sig sdk.StdSignature, signBytes []byte,
	isFeePayer bool) (acc sdk.Account, res sdk.Result) {

	// Get the account
	acc = am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnrecognizedAddress(addr).Result()
	}

	// Check and increment sequence number.
	seq := acc.GetSequence()
	if seq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	acc.SetSequence(seq + 1)

	// If pubkey is not known for account,
	// set it from the StdSignature
	pubKey := acc.GetPubKey()
	if pubKey.Empty() {
		if sig.PubKey.Empty() {
			return nil, sdk.ErrInternal("public Key not found").Result()
		}
		if !reflect.DeepEqual(sig.PubKey.Address(), addr) {
			return nil, sdk.ErrInternal(
				fmt.Sprintf("invalid PubKey for address %v", addr)).Result()
		}
		pubKey = sig.PubKey
		if pubKey.Empty() {
			return nil, sdk.ErrMissingPubKey(addr).Result()
		}
		err := acc.SetPubKey(pubKey)
		if err != nil {
			return nil, sdk.ErrInternal("setting PubKey on signer").Result()
		}
	}
	// Check sig.
	if !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	if isFeePayer {
		// TODO: pay fees
	}

	// Save the account.
	am.SetAccount(ctx, acc)
	return
}
