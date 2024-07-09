package ante

import (
	"crypto/sha256"
	"time"

	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
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
	maxTimeoutDuration time.Duration
	txManager          *unorderedtx.Manager
	env                appmodule.Environment
}

func NewUnorderedTxDecorator(maxDuration time.Duration, m *unorderedtx.Manager, env appmodule.Environment) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxTimeoutDuration: maxDuration,
		txManager:          m,
		env:                env,
	}
}

func (d *UnorderedTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, _ bool, next sdk.AnteHandler) (sdk.Context, error) {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return next(ctx, tx, false)
	}

	headerInfo := d.env.HeaderService.HeaderInfo(ctx)
	timeoutTimestamp := unorderedTx.GetTimeoutTimeStamp()
	if timeoutTimestamp.IsZero() || timeoutTimestamp.Unix() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction must have timeout_timestamp set")
	}
	if timeoutTimestamp.Before(headerInfo.Time) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "unordered transaction has a timeout_timestamp that has already passed")
	}
	if timeoutTimestamp.After(headerInfo.Time.Add(d.maxTimeoutDuration)) {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unordered tx ttl exceeds %s", d.maxTimeoutDuration.String())
	}

	txHash := sha256.Sum256(ctx.TxBytes())

	// check for duplicates
	if d.txManager.Contains(txHash) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "tx %X is duplicated")
	}

	if d.env.TransactionService.ExecMode(ctx) == transaction.ExecModeFinalize {
		// a new tx included in the block, add the hash to the unordered tx manager
		d.txManager.Add(txHash, timeoutTimestamp)
	}

	return next(ctx, tx, false)
}
