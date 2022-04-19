package middleware

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type HasExtensionOptionsTx interface {
	GetExtensionOptions() []*codectypes.Any
	GetNonCriticalExtensionOptions() []*codectypes.Any
}

// ExtensionOptionChecker is a function that returns true if the extension option is accepted.
type ExtensionOptionChecker func(*codectypes.Any) bool

// rejectExtensionOption is the default extension check that reject all tx
// extensions.
func rejectExtensionOption(*codectypes.Any) bool {
	return false
}

type rejectExtensionOptionsTxHandler struct {
	next    tx.Handler
	checker ExtensionOptionChecker
}

// NewExtensionOptionsMiddleware creates a new middleware that rejects all extension
// options which can optionally be included in protobuf transactions that don't pass the checker.
// Users that need extension options should pass a custom checker that returns true for the
// needed extension options.
func NewExtensionOptionsMiddleware(checker ExtensionOptionChecker) tx.Middleware {
	if checker == nil {
		checker = rejectExtensionOption
	}
	return func(txh tx.Handler) tx.Handler {
		return rejectExtensionOptionsTxHandler{
			next:    txh,
			checker: checker,
		}
	}
}

var _ tx.Handler = rejectExtensionOptionsTxHandler{}

func checkExtOpts(tx sdk.Tx, checker ExtensionOptionChecker) error {
	if hasExtOptsTx, ok := tx.(HasExtensionOptionsTx); ok {
		for _, opt := range hasExtOptsTx.GetExtensionOptions() {
			if !checker(opt) {
				return sdkerrors.ErrUnknownExtensionOptions
			}
		}
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (txh rejectExtensionOptionsTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	if err := checkExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh rejectExtensionOptionsTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := checkExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, err
	}

	return txh.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh rejectExtensionOptionsTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := checkExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, err
	}

	return txh.next.SimulateTx(ctx, req)
}
