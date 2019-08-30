package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	errs "github.com/cosmos/cosmos-sdk/types/errors"
)

func MempoolFeeDecorator(ctx Context, tx Tx, simulate bool, next AnteHandler) (newCtx Context, err error) {
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		return ctx, errs.Wrap(errs.ErrInternal, "Tx must be a StdTx")
	}
	stdFee := stdTx.Fee

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
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

			return ctx, errs.Wrapf(errs.ErrInsufficientFee, "insufficient fees; got: %q required: %q", stdFee.Amount, requiredFees)
		}
	}

	return next(ctx, tx, simulate)
}

func NewDeductFeeDecorator(ak keeper.AccountKeeper, supplyKeeper types.SupplyKeeper) {
	return func(ctx Context, tx Tx, simulate bool, next AnteHandler) (newCtx Context, err error) {
		stdTx, ok := tx.(types.StdTx)
		if !ok {
			return ctx, errs.Wrap(errs.ErrInternal, "Tx must be a StdTx")
		}

		if addr := supplyKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
			panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
		}

		// stdSigs contains the sequence number, account number, and signatures.
		// When simulating, this would just be a 0-length slice.
		signerAddrs := stdTx.GetSigners()

		// fetch first signer, who's going to pay the fees
		feePayer, err = GetSignerAcc(newCtx, ak, signerAddrs[0])
		if err != nil {
			return ctx, err
		}

		// deduct the fees
		if !stdTx.Fee.Amount.IsZero() {
			err = DeductFees(supplyKeeper, newCtx, feePayer, stdTx.Fee.Amount)
			if err != nil {
				return ctx, err
			}
		}

		return next(ctx, tx, simulate)
	}
}

// DeductFees deducts fees from the given account.
//
// NOTE: We could use the CoinKeeper (in addition to the AccountKeeper, because
// the CoinKeeper doesn't give us accounts), but it seems easier to do this.
func DeductFees(supplyKeeper types.SupplyKeeper, ctx sdk.Context, acc exported.Account, fees sdk.Coins) error {
	blockTime := ctx.BlockHeader().Time
	coins := acc.GetCoins()

	if !fees.IsValid() {
		return errs.Wrapf(errs.ErrInsufficientFee, "invalid fee amount: %s", fees)
	}

	// verify the account has enough funds to pay for fees
	_, hasNeg := coins.SafeSub(fees)
	if hasNeg {
		return errs.Wrapf(errs.ErrInsufficientFunds,
			"insufficient funds to pay for fees; %s < %s", coins, fees)
	}

	// Validate the account has enough "spendable" coins as this will cover cases
	// such as vesting accounts.
	spendableCoins := acc.SpendableCoins(blockTime)
	if _, hasNeg := spendableCoins.SafeSub(fees); hasNeg {
		return errs.Wrapf(errs.ErrInsufficientFunds,
			"insufficient funds to pay for fees; %s < %s", spendableCoins, fees)
	}

	err := supplyKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), types.FeeCollectorName, fees)
	if err != nil {
		return err.Result()
	}

	return sdk.Result{}
}
