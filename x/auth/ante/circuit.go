package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type circuitBreakerDecorator struct {
	keeper CircuitBreakerKeeper
}

func (d *circuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if !d.keeper.IsCircuitOpen(ctx) {
		return next(ctx, tx, simulate)
	}

	return ctx, sdkerrors.Wrap(sdkerrors.ErrActivation, "transaction is blocked due to Circuit Breaker activation")
}

func NewCircuitBreakerDecorator(keeper CircuitBreakerKeeper) sdk.AnteDecorator {
	return &circuitBreakerDecorator{keeper: keeper}
}
