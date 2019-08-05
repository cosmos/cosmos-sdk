package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewAnteHandler(ak AccountKeeper, supplyKeeper SupplyKeeper) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, simulate bool,
	) (newCtx sdk.Context, res sdk.Result, abort bool) {
		if addr := supplyKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
			panic(fmt.Sprintf("%s module account has not been set", types.FeeCollectorName))
		}

		// all transactions must be of type auth.StdTx
		stdTx, ok := tx.(StdTx)
		if !ok {
			// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
			// during runTx.
			newCtx = SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		params := ak.GetParams(ctx)

		newCtx = SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

		// AnteHandlers must have their own defer/recover in order for the BaseApp
		// to know how much gas was used! This is because the GasMeter is created in
		// the AnteHandler, but if it panics the context won't be set properly in
		// runTx's recover call.
		defer func() {
			if r := recover(); r != nil {
				switch rType := r.(type) {
				case sdk.ErrorOutOfGas:
					log := fmt.Sprintf(
						"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
						rType.Descriptor, stdTx.Fee.Gas, newCtx.GasMeter().GasConsumed(),
					)
					res = sdk.ErrOutOfGas(log).Result()

					res.GasWanted = stdTx.Fee.Gas
					res.GasUsed = newCtx.GasMeter().GasConsumed()
					abort = true
				default:
					panic(r)
				}
			}
		}()

		newCtx.GasMeter().ConsumeGas(params.TxSizeCostPerByte*sdk.Gas(len(newCtx.TxBytes())), "txSize")

		signerAccs[0], res = GetSignerAcc(newCtx, ak, signerAddrs[0])
		if !res.IsOK() {
			return newCtx, res, true
		}
		// deduct the fees
		if !stdTx.Fee.Amount.IsZero() {
			res = DeductFees(supplyKeeper, newCtx, signerAccs[0], stdTx.Fee.Amount)
			if !res.IsOK() {
				return newCtx, res, true
			}
		}

		return newCtx, sdk.Result{GasWanted: stdTx.Fee.Gas}, false // continue...
	}
}

func DeductFees(supplyKeeper SupplyKeeper, ctx sdk.Context, acc authexported.Account, fees sdk.Coins) sdk.Result {
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
