package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SetUpDecorator sets the GasMeter in the Context and wraps the next AnteHandler with a defer clause
// to recover from any downstream OutOfGas panics in the AnteHandler chain to return an error with information
// on gas provided and gas used.
// Should be the first decorator in the chain
type SetUpDecorator struct{}

func NewSetupDecorator() SetUpDecorator {
	return SetUpDecorator{}
}

func (sud SetUpDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// all transactions must be of type auth.StdTx
	stdTx, ok := tx.(types.StdTx)
	if !ok {
		// Set a gas meter with limit 0 as to prevent an infinite gas meter attack
		// during runTx.
		newCtx = SetGasMeter(simulate, ctx, 0)
		return newCtx, errs.Wrap(errs.ErrTxDecode, "Tx must be auth.StdTx")
	}

	newCtx = SetGasMeter(simulate, ctx, stdTx.Fee.Gas)

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
					rType.Descriptor, stdTx.Fee.Gas, newCtx.GasMeter().GasConsumed())

				err = errs.Wrap(errs.ErrOutOfGas, log)
				// TODO: figure out how to return Context, error so that baseapp can recover gasWanted/gasUsed
			default:
				panic(r)
			}
		}
	}()

	return next(newCtx, tx, simulate)
}

// SetGasMeter returns a new context with a gas meter set from a given context.
func SetGasMeter(simulate bool, ctx sdk.Context, gasLimit uint64) sdk.Context {
	// In various cases such as simulation and during the genesis block, we do not
	// meter any gas utilization.
	if simulate || ctx.BlockHeight() == 0 {
		return ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	}

	return ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))
}
