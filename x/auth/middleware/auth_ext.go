package middleware

import (
	"context"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// ExtensionOptionChecker is a function that returns true if the extension option is accepted.
type ExtensionOptionChecker func(*codectypes.Any) bool

func rejectExtensionOption(*codectypes.Any) bool {
	return false
}

type HasAuthExtensionOptionsTx interface {
	AuthExtensionOptions() []*codectypes.Any
}

type checkAuthExtensionOptionsTxHandler struct {
	next    tx.Handler
	checker ExtensionOptionChecker
}

// NewAuthExtensionOptionsMiddleware creates a new checkAuthExtensionOptionsMiddleware.
// NewAuthExtensionOptionsMiddleware is a middleware that checks all auth_info extension
// options pass the checker.
// If checker is nil, it defaults to a function that rejects all extension options.
func NewAuthExtensionOptionsMiddleware(checker ExtensionOptionChecker) tx.Middleware {
	if checker == nil {
		checker = rejectExtensionOption
	}
	return func(next tx.Handler) tx.Handler {
		return checkAuthExtensionOptionsTxHandler{
			next,
			checker,
		}
	}
}

var _ tx.Handler = checkAuthExtensionOptionsTxHandler{}

func checkAuthExtOpts(tx sdk.Tx, checker ExtensionOptionChecker) error {
	if hasExtOptsTx, ok := tx.(HasAuthExtensionOptionsTx); ok {
		for _, opt := range hasExtOptsTx.AuthExtensionOptions() {
			if !checker(opt) {
				return sdkerrors.ErrInvalidRequest.Wrapf("Unknown auth extension option: %T", opt)
			}
		}
	}

	return nil
}

// CheckTx implements tx.Handler.CheckTx.
func (txh checkAuthExtensionOptionsTxHandler) CheckTx(ctx context.Context, req tx.Request, checkReq tx.RequestCheckTx) (tx.Response, tx.ResponseCheckTx, error) {
	if err := checkAuthExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, tx.ResponseCheckTx{}, err
	}

	return txh.next.CheckTx(ctx, req, checkReq)
}

// DeliverTx implements tx.Handler.DeliverTx.
func (txh checkAuthExtensionOptionsTxHandler) DeliverTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := checkAuthExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, err
	}

	return txh.next.DeliverTx(ctx, req)
}

// SimulateTx implements tx.Handler.SimulateTx method.
func (txh checkAuthExtensionOptionsTxHandler) SimulateTx(ctx context.Context, req tx.Request) (tx.Response, error) {
	if err := checkAuthExtOpts(req.Tx, txh.checker); err != nil {
		return tx.Response{}, err
	}

	return txh.next.SimulateTx(ctx, req)
}
