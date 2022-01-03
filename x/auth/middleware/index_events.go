package middleware

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type indexEventsTxHandler struct {
	// indexEvents defines the set of events in the form {eventType}.{attributeKey},
	// which informs Tendermint what to index. If empty, all events will be indexed.
	indexEvents map[string]struct{}
	next        tx.Handler
}

// NewIndexEventsTxMiddleware defines a middleware to optionally only index a
// subset of the emitted events inside the Tendermint events indexer.
func NewIndexEventsTxMiddleware(indexEvents map[string]struct{}) tx.Middleware {
	return func(txHandler tx.Handler) tx.Handler {
		return indexEventsTxHandler{
			indexEvents: indexEvents,
			next:        txHandler,
		}
	}
}

var _ tx.Handler = indexEventsTxHandler{}

// CheckTx implements tx.Handler.CheckTx method.
func (txh indexEventsTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	res, resCheckTx, err := txh.next.CheckTx(ctx, req, checkReq)
	if err != nil {
		return res, tx.ResponseCheckTx{}, err
	}

	res.Events = sdk.MarkEventsToIndex(res.Events, txh.indexEvents)
	return res, resCheckTx, nil
}

// DeliverTx implements tx.Handler.DeliverTx method.
func (txh indexEventsTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	res, err := txh.next.DeliverTx(ctx, req)
	if err != nil {
		return res, err
	}

	res.Events = sdk.MarkEventsToIndex(res.Events, txh.indexEvents)
	return res, nil
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh indexEventsTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	res, err := txh.next.SimulateTx(ctx, req)
	if err != nil {
		return res, err
	}

	res.Events = sdk.MarkEventsToIndex(res.Events, txh.indexEvents)
	return res, nil
}
