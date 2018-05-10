package auth

import (
	"bytes"
	"fmt"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
)

// NewAnteHandler returns an AnteHandler that checks
// and increments sequence numbers, checks signatures,
// and deducts fees from the first signer.
func NewAnteHandler(accountMapper AccountMapper, feeHandler bam.FeeHandler) bam.AnteHandler {
	return func(
		ctx sdk.Context, tx bam.Tx,
	) (_ sdk.Context, _ sdk.Result, abort bool) {

		stdTx, ok := tx.(StdTx)
		if !ok {
			return ctx, bam.ErrInternal("tx must be bam.StdTx").Result(), true
		}

		// Assert that there are signatures.
		var sigs = stdTx.GetSignatures()
		if len(sigs) == 0 {
			return ctx,
				bam.ErrUnauthorized("no signers").Result(),
				true
		}

		msg := stdTx.GetMsg()

		// Assert that number of signatures is correct.
		var signerAddrs = msg.GetSigners()
		if len(sigs) != len(signerAddrs) {
			return ctx,
				bam.ErrUnauthorized("wrong number of signers").Result(),
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
		signBytes := StdSignBytes(ctx.ChainID(), sequences, fee, msg)

		// Check sig and nonce and collect signer accounts.
		var signerAccs = make([]Account, len(signerAddrs))
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
					feeHandler(ctx, tx, fee.Amount)
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
	ctx sdk.Context, am AccountMapper,
	addr bam.Address, sig StdSignature, signBytes []byte) (
	acc Account, res sdk.Result) {

	// Get the account.
	acc = am.GetAccount(ctx, addr)
	if acc == nil {
		return nil, bam.ErrUnknownAddress(addr.String()).Result()
	}

	// Check and increment sequence number.
	seq := acc.GetSequence()
	if seq != sig.Sequence {
		return nil, bam.ErrInvalidSequence(
			fmt.Sprintf("Invalid sequence. Got %d, expected %d", sig.Sequence, seq)).Result()
	}
	acc.SetSequence(seq + 1)

	// If pubkey is not known for account,
	// set it from the StdSignature.
	pubKey := acc.GetPubKey()
	if pubKey == nil {
		pubKey = sig.PubKey
		if pubKey == nil {
			return nil, bam.ErrInvalidPubKey("PubKey not found").Result()
		}
		if !bytes.Equal(pubKey.Address(), addr) {
			return nil, bam.ErrInvalidPubKey(
				fmt.Sprintf("PubKey does not match Signer address %v", addr)).Result()
		}
		err := acc.SetPubKey(pubKey)
		if err != nil {
			return nil, bam.ErrInternal("setting PubKey on signer's account").Result()
		}
	}

	// Check sig.
	if !pubKey.VerifyBytes(signBytes, sig.Signature) {
		return nil, bam.ErrUnauthorized("signature verification failed").Result()
	}

	return
}

// Deduct the fee from the account.
// We could use the CoinKeeper (in addition to the AccountMapper,
// because the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func deductFees(acc Account, fee StdFee) (Account, sdk.Result) {
	coins := acc.GetCoins()
	feeAmount := fee.Amount

	newCoins := coins.Minus(feeAmount)
	if !newCoins.IsNotNegative() {
		errMsg := fmt.Sprintf("%s < %s", coins, feeAmount)
		return nil, bam.ErrInsufficientFunds(errMsg).Result()
	}
	acc.SetCoins(newCoins)
	return acc, sdk.Result{}
}

// BurnFeeHandler burns all fees (decreasing total supply)
func BurnFeeHandler(ctx sdk.Context, tx bam.Tx, fee bam.Coins) {
}
