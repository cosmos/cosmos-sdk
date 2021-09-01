package middleware

import (
	"context"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type HasExtensionOptionsTx interface {
	GetExtensionOptions() []*codectypes.Any
	GetNonCriticalExtensionOptions() []*codectypes.Any
}

type rejectExtensionOptionsMiddleware struct {
	next tx.Handler
}

// NewRejectExtensionOptionsMiddleware creates a new rejectExtensionOptionsMiddleware.
// rejectExtensionOptionsMiddleware is a middleware that rejects all extension
// options which can optionally be included in protobuf transactions. Users that
// need extension options should create a custom AnteHandler chain that handles
// needed extension options properly and rejects unknown ones.
func RejectExtensionOptionsMiddleware(txh tx.Handler) tx.Handler {
	return rejectExtensionOptionsMiddleware{
		next: txh,
	}
}

var _ tx.Handler = rejectExtensionOptionsMiddleware{}

// CheckTx implements tx.Handler.CheckTx.
func (txh rejectExtensionOptionsMiddleware) CheckTx(ctx context.Context, tx sdk.Tx, req abci.RequestCheckTx) (abci.ResponseCheckTx, error) {
	if hasExtOptsTx, ok := tx.(HasExtensionOptionsTx); ok {
		if len(hasExtOptsTx.GetExtensionOptions()) != 0 {
			return abci.ResponseCheckTx{}, sdkerrors.ErrUnknownExtensionOptions
		}
	}

	return txh.next.CheckTx(ctx, tx, req)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh rejectExtensionOptionsMiddleware) DeliverTx(ctx context.Context, tx sdk.Tx, req abci.RequestDeliverTx) (abci.ResponseDeliverTx, error) {
	fmt.Println("1")
	if hasExtOptsTx, ok := tx.(HasExtensionOptionsTx); ok {
		if len(hasExtOptsTx.GetExtensionOptions()) != 0 {
			return abci.ResponseDeliverTx{}, sdkerrors.ErrUnknownExtensionOptions
		}
	}

	return txh.next.DeliverTx(ctx, tx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh rejectExtensionOptionsMiddleware) SimulateTx(ctx context.Context, sdkTx sdk.Tx, req tx.RequestSimulateTx) (tx.ResponseSimulateTx, error) {
	if hasExtOptsTx, ok := sdkTx.(HasExtensionOptionsTx); ok {
		if len(hasExtOptsTx.GetExtensionOptions()) != 0 {
			return tx.ResponseSimulateTx{}, sdkerrors.ErrUnknownExtensionOptions
		}
	}

	return txh.next.SimulateTx(ctx, sdkTx, req)
}
