package auth

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
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

		// TODO: will this always be a stdtx? should that be used in the function signature?
		stdTx, ok := tx.(sdk.StdTx)
		if !ok {
			return ctx, sdk.ErrInternal("tx must be sdk.StdTx").Result(), true
		}

		// Assert that number of signatures is correct.
		var signerAddrs = msg.GetSigners()
		if len(sigs) != len(signerAddrs) {
			return ctx,
				sdk.ErrUnauthorized("wrong number of signers").Result(),
				true
		}

		// Get the sign bytes (requires all sequence numbers and the fee)
		sequences := make([]int64, len(signerAddrs))
		for i := 0; i < len(signerAddrs); i++ {
			sequences[i] = sigs[i].Sequence
		}
		fee := stdTx.Fee
		chainID := ctx.ChainID()
		// XXX: major hack; need to get ChainID
		// into the app right away (#565)
		if chainID == "" {
			chainID = viper.GetString("chain-id")
		}
		signBytes := sdk.StdSignBytes(ctx.ChainID(), sequences, fee, msg)

		// Check sig and nonce and collect signer accounts.
		var signerAccs = make([]sdk.Account, len(signerAddrs))
		for i := 0; i < len(sigs); i++ {
			signerAddr, sig := signerAddrs[i], sigs[i]

			// check signature, return account with incremented nonce
			signerAcc, res := processSig(
				ctx, accountMapper,
				signerAddr, sig, signBytes,
			)
			if !res.IsOK() {
				return ctx, res, true
			}

			// first sig pays the fees
			if i == 0 {
				// TODO: min fee
				if !fee.Amount.IsZero() {
					signerAcc, res = deductFees(signerAcc, fee)
					if !res.IsOK() {
						return ctx, res, true
					}
				}
			}

			// Save the account.
			accountMapper.SetAccount(ctx, signerAcc)
			signerAccs[i] = signerAcc
		}

		// cache the signer accounts in the context
		ctx = WithSigners(ctx, signerAccs)

		// TODO: tx tags (?)

		return ctx, sdk.Result{}, false // continue...
	}
}

// verify the signature and increment the sequence.
// if the account doesn't have a pubkey, set it.
func processSig(
	ctx sdk.Context, am sdk.AccountMapper,
	addr sdk.Address, sig sdk.StdSignature, signBytes []byte) (
	acc sdk.Account, res sdk.Result) {

	// Get the account.
	acc = am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(addr.String()).Result()
	}

	// Check and increment sequence number.
	seq := acc.GetSequence()
	if seq != sig.Sequence {
		return nil, sdk.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	acc.SetSequence(seq + 1)

	// If pubkey is not known for account,
	// set it from the StdSignature.
	pubKey := acc.GetPubKey()
	if pubKey.Empty() {
		pubKey = sig.PubKey
		if pubKey.Empty() {
			return nil, sdk.ErrInvalidPubKey("PubKey not found").Result()
		}
		if !bytes.Equal(pubKey.Address(), addr) {
			return nil, sdk.ErrInvalidPubKey(
				fmt.Sprintf("PubKey does not match Signer address %v", addr)).Result()
		}
		err := acc.SetPubKey(pubKey)
		if err != nil {
			return nil, sdk.ErrInternal("setting PubKey on signer's account").Result()
		}
	}

	// Check sig.
	if !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, sdk.ErrUnauthorized("signature verification failed").Result()
	}

	return
}

// Deduct the fee from the account.
// We could use the CoinKeeper (in addition to the AccountMapper,
// because the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func deductFees(acc sdk.Account, fee sdk.StdFee) (sdk.Account, sdk.Result) {
	coins := acc.GetCoins()
	feeAmount := fee.Amount

	newCoins := coins.Minus(feeAmount)
	if !newCoins.IsNotNegative() {
		errMsg := fmt.Sprintf("%s < %s", coins, feeAmount)
		return nil, sdk.ErrInsufficientFunds(errMsg).Result()
	}
	acc.SetCoins(newCoins)
	return acc, sdk.Result{}
}
