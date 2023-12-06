package ante

import (
	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultMaxUnOrderedTTL defines the default maximum TTL an un-ordered transaction
	// can set.
	DefaultMaxUnOrderedTTL = 1024
)

var _ sdk.AnteDecorator = (*UnorderedTxDecorator)(nil)

// UnorderedTxDecorator defines an AnteHandler decorator that is responsible for
// checking if a transaction is intended to be unordered and if so, evaluates
// the transaction accordingly. An unordered transaction will bypass having it's
// nonce incremented, which allows fire-and-forget along with possible parallel
// transaction processing, without having to deal with nonces.
//
// The transaction sender must ensure that unordered=true and a timeout_height
// is appropriately set. The AnteHandler will check that the transaction is not
// a duplicate and will evict it from memory when the timeout is reached.
type UnorderedTxDecorator struct {
	// maxUnOrderedTTL defines the maximum TTL a transaction can define.
	maxUnOrderedTTL uint64
}

func NewUnorderedTxDecorator(maxTTL uint64) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxUnOrderedTTL: maxTTL,
	}
}

func (d *UnorderedTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return next(ctx, tx, simulate)
	}

	if unorderedTx.GetTimeoutHeight() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction must have timeout_height set")
	}

	if unorderedTx.GetTimeoutHeight() > uint64(ctx.BlockHeight())+d.maxUnOrderedTTL {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unordered tx ttl exceeds %d", d.maxUnOrderedTTL)
	}
}
