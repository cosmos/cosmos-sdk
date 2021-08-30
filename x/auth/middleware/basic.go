package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ValidateBasicDecorator will call tx.ValidateBasic, msg.ValidateBasic(for each msg inside tx)
// and return any non-nil error.
// If ValidateBasic passes, middleware calls next middleware in chain. Note,
// validateBasicMiddleware will not get executed on ReCheckTx since it
// is not dependent on application state.
type validateBasicMiddleware struct {
	next tx.Handler
}

func ValidateBasicMiddleware(txh tx.Handler) tx.Handler {
	return validateBasicMiddleware{
		next: txh,
	}
}

var _ tx.Handler = validateBasicMiddleware{}

// CheckTx implements tx.Handler.CheckTx.
func (basic validateBasicMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// no need to validate basic on recheck tx, call next antehandler
	if sdkCtx.IsReCheckTx() {
		return basic.next.CheckTx(ctx, tx, req)
	}

	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseCheckTx{}, err
	}

	return basic.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (basic validateBasicMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	if err := tx.ValidateBasic(); err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	return basic.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx.
func (basic validateBasicMiddleware) SimulateTx(ctx context.Context, tx sdk.Tx, req txtypes.RequestSimulateTx) (txtypes.ResponseSimulateTx, error) {
	if err := tx.ValidateBasic(); err != nil {
		return txtypes.ResponseSimulateTx{}, err
	}

	return basic.next.SimulateTx(ctx, tx, req)
}
