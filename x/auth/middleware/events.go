package middleware

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type eventsTxHandler struct {
	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}
	inner       tx.TxHandler
}

func NewEventsTxMiddleware(indexEvents map[string]struct{}) tx.TxMiddleware {
	return func(txHandler tx.TxHandler) tx.TxHandler {
		return eventsTxHandler{
			indexEvents: indexEvents,
			inner:       txHandler,
		}
	}
}

var _ tx.TxHandler = eventsTxHandler{}

func (txh eventsTxHandler) CheckTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	res, err := txh.inner.CheckTx(ctx, tx, req)
	if err != nil {
		return res, err
	}

	res.Events = sdk.MarkEventsToIndex(res.Events, txh.indexEvents)
	return res, nil
}

func (txh eventsTxHandler) DeliverTx(ctx sdk.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	res, err := txh.inner.DeliverTx(ctx, tx, req)
	if err != nil {
		return res, err
	}

	res.Events = sdk.MarkEventsToIndex(res.Events, txh.indexEvents)
	return res, nil
}
