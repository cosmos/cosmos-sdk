package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// txPriorityDecorator implements tx antehandler that determines a
// transaction's priority via a naive mechanism -- the total sum of fees provided.
// It sets the Priority in CheckTx only.
//
// We set the Priority of the transaction
// to be ordered in the Tendermint mempool based naively on the total sum of all
// fees included. Applications that need more sophisticated mempool ordering
// should look to implement their own fee handling middleware instead of using
// TxPriorityHandler.
type txPriorityDecorator struct{}

func NewTxPriorityDecorator() sdk.AnteDecorator {
	return txPriorityDecorator{}
}

// AnteHandle implements sdk.AnteDecorator. We set the Priority of the transaction
// to be ordered in the Tendermint mempool based naively on the total sum of all
// fees included. Applications that need more sophisticated mempool ordering
// should look to implement their own fee handling middleware instead of using
// TxPriorityHandler.
func (vbd txPriorityDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if !ctx.IsCheckTx() {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()

	// TODO What's the best way to set the priority on the response?
	GetTxPriority(feeCoins)

	return next(ctx, tx, simulate)
}

// GetTxPriority returns a naive tx priority based on the total sum of all fees
// provided in a transaction.
func GetTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		priority += c.Amount.Int64()
	}

	return priority
}
