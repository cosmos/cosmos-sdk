package ante

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RejectFeeGranterDecorator is an AnteDecorator which rejects transactions which
// have the Fee.granter field set. It is to be used by chains which do not support
// fee grants.
type RejectFeeGranterDecorator struct{}

// NewRejectFeeGranterDecorator returns a new RejectFeeGranterDecorator.
func NewRejectFeeGranterDecorator() RejectFeeGranterDecorator {
	return RejectFeeGranterDecorator{}
}

var _ types.AnteDecorator = RejectFeeGranterDecorator{}

func (d RejectFeeGranterDecorator) AnteHandle(ctx types.Context, tx types.Tx, simulate bool, next types.AnteHandler) (newCtx types.Context, err error) {
	feeTx, ok := tx.(types.FeeTx)
	if ok && len(feeTx.FeeGranter()) != 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not supported")
	}

	return next(ctx, tx, simulate)
}
