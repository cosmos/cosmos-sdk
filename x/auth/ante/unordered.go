package ante

import (
	"crypto/sha256"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
//
// The UnorderedTxDecorator should be placed as early as possible in the AnteHandler
// chain to ensure that during DeliverTx, the transaction is added to the UnorderedTxManager.
type UnorderedTxDecorator struct {
	// maxUnOrderedTTL defines the maximum TTL a transaction can define.
	maxUnOrderedTTL uint64
	txManager       *unorderedtx.Manager
}

func NewUnorderedTxDecorator(maxTTL uint64, m *unorderedtx.Manager) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxUnOrderedTTL: maxTTL,
		txManager:       m,
	}
}

func (d *UnorderedTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, _ bool, next sdk.AnteHandler) (sdk.Context, error) {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return next(ctx, tx, ctx.ExecMode() == sdk.ExecModeSimulate)
	}

	// TTL is defined as a specific block height at which this tx is no longer valid
	ttl := unorderedTx.GetTimeoutHeight()

	if ttl == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction must have timeout_height set")
	}
	if ttl < uint64(ctx.BlockHeight()) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction has a timeout_height that has already passed")
	}
	if ttl > uint64(ctx.BlockHeight())+d.maxUnOrderedTTL {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unordered tx ttl exceeds %d", d.maxUnOrderedTTL)
	}

	txHash := sha256.Sum256(ctx.TxBytes())

	// check for duplicates
	if d.txManager.Contains(txHash) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "tx %X is duplicated")
	}

	if ctx.ExecMode() == sdk.ExecModeFinalize {
		// a new tx included in the block, add the hash to the unordered tx manager
		d.txManager.Add(txHash, ttl)
	}

	return next(ctx, tx, ctx.ExecMode() == sdk.ExecModeSimulate)
}
