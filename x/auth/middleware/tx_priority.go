package middleware

import (
	"context"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

var _ tx.Handler = txPriorityHandler{}

type txPriorityHandler struct {
	next tx.Handler
}

// TxPriorityHandler implements tx handling middleware that determines a
// transaction's priority via a naive mechanism -- the total sum of fees provided.
// It sets the Priority in ResponseCheckTx only.
func TxPriorityHandler(txh tx.Handler) tx.Handler {
	return txPriorityHandler{next: txh}
}

func (h txPriorityHandler) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return abci.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()

	res, err := h.next.CheckTx(ctx, tx, req)
	res.Priority = GetTxPriority(feeCoins)

	return res, err
}

func (h txPriorityHandler) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	return h.next.DeliverTx(ctx, tx, req)
}

func (h txPriorityHandler) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	return h.next.SimulateTx(ctx, sdkTx, req)
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
