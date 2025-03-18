package ante

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	errorsmod "cosmossdk.io/errors"
)

const (
	// MaxTimeoutDuration defines the maximum TTL a transaction can define.
	MaxTimeoutDuration = 10 * time.Minute
)

var _ sdk.AnteDecorator = (*UnorderedTxDecorator)(nil)

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
// chain to ensure that during DeliverTx, the transaction is added to the UnorderedTxManager.
type UnorderedTxDecorator struct {
	txManager authkeeper.UnorderedTxManager
}

func NewUnorderedTxDecorator(
	utxm authkeeper.UnorderedTxManager,
) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		txManager: utxm,
	}
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
	if timeoutTimestamp.After(blockTime.Add(MaxTimeoutDuration)) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"unordered tx ttl exceeds %s",
			MaxTimeoutDuration.String(),
		)
	}

	execMode := ctx.ExecMode()
	if execMode == sdk.ExecModeSimulate {
		return nil
	}

	signerAddrs, err := getSigners(tx)
	if err != nil {
		return err
	}

	for _, signerAddr := range signerAddrs {
		contains, err := d.txManager.Contains(ctx, signerAddr, unorderedTx.GetTimeoutTimeStamp())
		if err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrIO,
				"failed to check contains for signer %x", signerAddr,
			)
		}
		if contains {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"tx is duplicated for signer %x", signerAddr,
			)
		}

		if err := d.txManager.Add(ctx, signerAddr, unorderedTx.GetTimeoutTimeStamp()); err != nil {
			return errorsmod.Wrapf(
				sdkerrors.ErrIO,
				"failed to add unordered nonce to state for signer %x", signerAddr,
			)
		}
	}

	return nil
}

func getSigners(tx sdk.Tx) ([][]byte, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return nil, err
	}

	addresses := make([][]byte, 0, len(sigs))
	for _, sig := range sigs {
		addresses = append(addresses, sig.PubKey.Bytes())
	}

	return addresses, nil
}
