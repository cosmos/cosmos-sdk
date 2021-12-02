package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

var _ tx.Handler = txPriorityHandler{}

type txPriorityHandler struct {
	next tx.Handler
}

// TxPriorityMiddleware implements tx handling middleware that determines a
// transaction's priority via a naive mechanism -- the total sum of fees provided.
// It sets the Priority in ResponseCheckTx only.
func TxPriorityMiddleware(txh tx.Handler) tx.Handler {
	return txPriorityHandler{next: txh}
}

// CheckTx implements tx.Handler.CheckTx. We set the Priority of the transaction
// to be ordered in the Tendermint mempool based naively on the total sum of all
// fees included. Applications that need more sophisticated mempool ordering
// should look to implement their own fee handling middleware instead of using
// TxPriorityHandler.
func (h txPriorityHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	feeTx, ok := req.Tx.(sdk.FeeTx)
	if !ok {
		return tx.Response{}, tx.ResponseCheckTx{}, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()

	res, checkRes, err := h.next.CheckTx(ctx, req, checkReq)
	checkRes.Priority = GetTxPriority(feeCoins)

	return res, checkRes, err
}

func (h txPriorityHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return h.next.DeliverTx(ctx, req)
}

func (h txPriorityHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	return h.next.SimulateTx(ctx, req)
}

// GetTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func GetTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := c.Amount.Int64()
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
