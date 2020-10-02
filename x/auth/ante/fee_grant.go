package ante

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RejectFeeGranterDecorator is a placeholder AnteDecorator for fee grants for
// when that functionality is enabled. Currently it simpy rejects transactions
// which have the Fee.granter field set.
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
