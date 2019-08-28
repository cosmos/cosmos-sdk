package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func SetUpDecorator(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx Context, error) {
	// all transactions must be of type auth.StdTx
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
		// during runTx.
		newCtx = SetGasMeter(simulate, ctx, 0)
		return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
	}

	ctx = SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

	// Decorator will catch an OutOfGasPanic caused in the next antehandler
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
				
				// TODO: figure out how to return Context, error so that baseapp can recover gasWanted/gasUsed
			default:
				panic(r)
			}
		}
	}()

	newCtx, err = next(ctx, tx, simulate)
	return
}