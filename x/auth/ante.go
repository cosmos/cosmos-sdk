package auth

import (
	"fmt"
	"reflect"

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

		// Get the sign bytes by collecting all sequence numbers
		sequences := make([]int64, len(signerAddrs))
		for i := 0; i < len(signerAddrs); i++ {
			sequences[i] = sigs[i].Sequence
		}
		signBytes := sdk.StdSignBytes(ctx.ChainID(), sequences, msg)

		// Check fee payer sig and nonce, and deduct fee.
		// This is done first because it only
		// requires fetching 1 account.
		payerAddr, payerSig := signerAddrs[0], sigs[0]
		payerAcc, res := processSig(ctx, accountMapper, payerAddr, payerSig, signBytes)
		if !res.IsOK() {
			return ctx, res, true
		}
		signerAccs[0] = payerAcc
		// TODO: Charge fee from payerAcc.
		// TODO: accountMapper.SetAccount(ctx, payerAddr)

		// Check sig and nonce for the rest.
		for i := 1; i < len(sigs); i++ {
			signerAddr, sig := signerAddrs[i], sigs[i]
			signerAcc, res := processSig(ctx, accountMapper, signerAddr, sig, signBytes)
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
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	acc.SetSequence(seq + 1)

	// Check and possibly set pubkey.
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
		err := acc.SetPubKey(pubKey)
		if err != nil {
			return nil, sdk.ErrInternal("setting PubKey on signer").Result()
		}
	}
	// Check sig.
	if !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	// Save the account.
	am.SetAccount(ctx, acc)
	return
}
