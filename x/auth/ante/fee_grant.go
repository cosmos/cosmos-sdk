package ante

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DeductGrantedFeeDecorator is a placeholder AnteDecorator for fee grants for
// when that functionality is enabled. Currently it simpy rejects transactions
// which have the Fee.granter field set.
type DeductGrantedFeeDecorator struct{}

// NewDeductGrantedFeeDecorator returns a new DeductGrantedFeeDecorator.
func NewDeductGrantedFeeDecorator() DeductGrantedFeeDecorator {
	return DeductGrantedFeeDecorator{}
}

var _ types.AnteDecorator = DeductGrantedFeeDecorator{}

func (d DeductGrantedFeeDecorator) AnteHandle(ctx types.Context, tx types.Tx, simulate bool, next types.AnteHandler) (newCtx types.Context, err error) {
	feeTx, ok := tx.(types.FeeTx)
	if ok && len(feeTx.FeeGranter()) != 0 {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "fee grants are not supported")
	}

	return next(ctx, tx, simulate)
}
