package ante

import (
	"time"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

const (
	// DefaultMaxTimoutDuration defines a default maximum TTL a transaction can define.
	DefaultMaxTimoutDuration = 10 * time.Minute
	// DefaultUnorderedTxGasCost defines a default gas cost for unordered transactions.
	// We must charge extra gas for unordered transactions
	// as they incur extra processing time for cleaning up the expired txs in x/auth PreBlocker.
	// Note: this value was chosen by 2x-ing the cost of fetching and removing an unordered nonce entry.
	DefaultUnorderedTxGasCost = uint64(2240)
)

var _ sdk.AnteDecorator = (*UnorderedTxDecorator)(nil)

type UnorderedTxDecoratorOptions func(*UnorderedTxDecorator)

// WithTimeoutDuration allows for changing the timeout duration for unordered txs.
func WithTimeoutDuration(duration time.Duration) UnorderedTxDecoratorOptions {
	return func(utx *UnorderedTxDecorator) {
		utx.maxTxTimeoutDuration = duration
	}
}

func WithUnorderedTxGasCost(cost uint64) UnorderedTxDecoratorOptions {
	return func(utx *UnorderedTxDecorator) {
		utx.txGasCost = cost
	}
}

// UnorderedTxDecorator defines an AnteHandler decorator that is responsible for
// checking if a transaction is intended to be unordered and, if so, evaluates
// the transaction accordingly. An unordered transaction will bypass having its
// nonce incremented, which allows fire-and-forget transaction broadcasting,
// removing the necessity of ordering on the sender-side.
//
// The transaction sender must ensure that unordered=true and a timeout_height
// is appropriately set. The AnteHandler will check that the transaction is not
// a duplicate and will evict it from state when the timeout is reached.
//
// The UnorderedTxDecorator should be placed as early as possible in the AnteHandler
// chain to ensure that during DeliverTx, the transaction is added to the UnorderedNonceManager.
type UnorderedTxDecorator struct {
	maxTxTimeoutDuration time.Duration
	txGasCost            uint64
	txManager            UnorderedNonceManager
}

func NewUnorderedTxDecorator(
	utxm UnorderedNonceManager,
	opts ...UnorderedTxDecoratorOptions,
) *UnorderedTxDecorator {
	utx := &UnorderedTxDecorator{
		maxTxTimeoutDuration: DefaultMaxTimoutDuration,
		txGasCost:            DefaultUnorderedTxGasCost,
		txManager:            utxm,
	}
	for _, opt := range opts {
		opt(utx)
	}

	return utx
}

func (d *UnorderedTxDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	_ bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	if err := d.ValidateTx(ctx, tx); err != nil {
		return ctx, err
	}
	return next(ctx, tx, false)
}

func (d *UnorderedTxDecorator) ValidateTx(ctx sdk.Context, tx sdk.Tx) error {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return nil
	}

	blockTime := ctx.BlockTime()
	timeoutTimestamp := unorderedTx.GetTimeoutTimeStamp()
	if timeoutTimestamp.IsZero() || timeoutTimestamp.Unix() == 0 {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"unordered transaction must have timeout_timestamp set",
		)
	}
	if timeoutTimestamp.Before(blockTime) {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"unordered transaction has a timeout_timestamp that has already passed",
		)
	}
	if timeoutTimestamp.After(blockTime.Add(d.maxTxTimeoutDuration)) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"unordered tx ttl exceeds %s",
			d.maxTxTimeoutDuration.String(),
		)
	}

	ctx.GasMeter().ConsumeGas(d.txGasCost, "unordered tx")

	execMode := ctx.ExecMode()
	if execMode == sdk.ExecModeSimulate {
		return nil
	}

	signerAddrs, err := extractSignersBytes(tx)
	if err != nil {
		return err
	}

	for _, signerAddr := range signerAddrs {
		if err := d.txManager.TryAddUnorderedNonce(ctx, signerAddr, unorderedTx.GetTimeoutTimeStamp()); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"failed to add unordered nonce: %s", err,
			)
		}
	}

	return nil
}

func extractSignersBytes(tx sdk.Tx) ([][]byte, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	return sigTx.GetSigners()
}
