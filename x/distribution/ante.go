package distribution

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func NewAnteHandler(fck auth.FeeCollectionKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

		// all transactions must be of type auth.StdTx
		stdTx, ok := tx.(auth.StdTx)
		if !ok {
			// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
			// during runTx.
			newCtx = sdk.SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		// Ensure that the provided fees meet a minimum threshold for the validator,
		// if this is a CheckTx. This is only for local mempool purposes, and thus
		// is only ran on check tx.
		if ctx.IsCheckTx() && !simulate {
			res := EnsureSufficientMempoolFees(ctx, stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		// NOTE: This should only get run by the first module
		// By placing this line here, we force users to place 
		// distribution module first in manager's AnteHandler list.
		newCtx = SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

		if !stdTx.Fee.Amount.IsZero() {
			signerAccs[0], res = DeductFees(ctx.BlockHeader().Time, signerAccs[0], stdTx.Fee)
			if !res.IsOK() {
				return newCtx, res, true
			}

			fck.AddCollectedFees(newCtx, stdTx.Fee.Amount)
		}	
	}
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, stdFee StdFee) sdk.Result {
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(stdFee.Gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !stdFee.Amount.IsAnyGTE(requiredFees) {
			return sdk.ErrInsufficientFee(
				fmt.Sprintf(
					"insufficient fees; got: %q required: %q", stdFee.Amount, requiredFees,
				),
			).Result()
		}
	}

	return sdk.Result{}
}

// DeductFees deducts fees from the given account.
//
// NOTE: We could use the CoinKeeper (in addition to the AccountKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(blockTime time.Time, acc Account, fee StdFee) (Account, sdk.Result) {
	coins := acc.GetCoins()
	feeAmount := fee.Amount

	if !feeAmount.IsValid() {
		return nil, sdk.ErrInsufficientFee(fmt.Sprintf("invalid fee amount: %s", feeAmount)).Result()
	}

	// get the resulting coins deducting the fees
	newCoins, ok := coins.SafeSub(feeAmount)
	if ok {
		return nil, sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds to pay for fees; %s < %s", coins, feeAmount),
		).Result()
	}

	// Validate the account has enough "spendable" coins as this will cover cases
	// such as vesting accounts.
	spendableCoins := acc.SpendableCoins(blockTime)
	if _, hasNeg := spendableCoins.SafeSub(feeAmount); hasNeg {
		return nil, sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds to pay for fees; %s < %s", spendableCoins, feeAmount),
		).Result()
	}

	if err := acc.SetCoins(newCoins); err != nil {
		return nil, sdk.ErrInternal(err.Error()).Result()
	}

	return acc, sdk.Result{}
}
