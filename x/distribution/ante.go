package distribution

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AnteHandler deducts fees from first signer
// It assumes tx is of type auth.StdTx
func AnteHandler(accountKeeper auth.AccountKeeper, supplyKeeper types.SupplyKeeper, ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {

	if addr := supplyKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
	}

	newCtx = ctx

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
	if newCtx.IsCheckTx() && !simulate {
		res := EnsureSufficientMempoolFees(newCtx, stdTx.Fee)
		if !res.IsOK() {
			return newCtx, res, true
		}
	}

	if !stdTx.Fee.Amount.IsZero() {
		signerAddrs := stdTx.GetSigners()
		feeAcc, res := auth.GetSignerAcc(newCtx, accountKeeper, signerAddrs[0])

		res = DeductFees(supplyKeeper, newCtx, feeAcc, stdTx.Fee.Amount)
		if !res.IsOK() {
			return newCtx, res, true
		}
	}
	
	return newCtx, sdk.Result{
		GasWanted: tx.Gas()}, false
}

// EnsureSufficientMempoolFees verifies that the given transaction has supplied
// enough fees to cover a proposer's minimum fees. A result object is returned
// indicating success or failure.
//
// Contract: This should only be called during CheckTx as it cannot be part of
// consensus.
func EnsureSufficientMempoolFees(ctx sdk.Context, fee auth.StdFee) sdk.Result {
	minGasPrices := ctx.MinGasPrices()
	if !minGasPrices.IsZero() {
		requiredFees := make(sdk.Coins, len(minGasPrices))

		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(fee.Gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}

		if !fee.Amount.IsAnyGTE(requiredFees) {
			return sdk.ErrInsufficientFee(
				fmt.Sprintf(
					"insufficient fees; got: %q required: %q", fee.Amount, requiredFees,
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
func DeductFees(supplyKeeper types.SupplyKeeper, ctx sdk.Context, acc exported.Account, fees sdk.Coins) sdk.Result {
	blockTime := ctx.BlockHeader().Time
	coins := acc.GetCoins()

	if !fees.IsValid() {
		return sdk.ErrInsufficientFee(fmt.Sprintf("invalid fee amount: %s", fees)).Result()
	}

	// verify the account has enough funds to pay for fees
	_, hasNeg := coins.SafeSub(fees)
	if hasNeg {
		return sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds to pay for fees; %s < %s", coins, fees),
		).Result()
	}

	// Validate the account has enough "spendable" coins as this will cover cases
	// such as vesting accounts.
	spendableCoins := acc.SpendableCoins(blockTime)
	if _, hasNeg := spendableCoins.SafeSub(fees); hasNeg {
		return sdk.ErrInsufficientFunds(
			fmt.Sprintf("insufficient funds to pay for fees; %s < %s", spendableCoins, fees),
		).Result()
	}

	err := supplyKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}