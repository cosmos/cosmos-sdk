package ante

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/gogoproto/proto"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante/unorderedtx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// bufPool is a pool of bytes.Buffer objects to reduce memory allocations.
var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// DefaultSha256GasCost is the suggested default gas cost for Sha256 operations in unordered transaction handling.
const DefaultSha256GasCost = 25

var _ sdk.AnteDecorator = (*UnorderedTxDecorator)(nil)

// UnorderedTxDecorator defines an AnteHandler decorator that is responsible for
// checking if a transaction is intended to be unordered and if so, evaluates
// the transaction accordingly. An unordered transaction will bypass having its
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
	sha256GasCost      uint64
}

func NewUnorderedTxDecorator(
	maxDuration time.Duration,
	m *unorderedtx.Manager,
	sha256GasCost uint64,
) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxTimeoutDuration: maxDuration,
		txManager:          m,
		sha256GasCost:      sha256GasCost,
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
	if timeoutTimestamp.After(blockTime.Add(d.maxTimeoutDuration)) {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"unordered tx ttl exceeds %s",
			d.maxTimeoutDuration.String(),
		)
	}

	// consume gas in all exec modes to avoid gas estimation discrepancies
	ctx.GasMeter().ConsumeGas(d.sha256GasCost, "consume gas for calculating tx hash")

	// Avoid checking for duplicates and creating the identifier in simulation mode
	// This is done to avoid sha256 computation in simulation mode
	if ctx.ExecMode() == sdk.ExecModeSimulate {
		return nil
	}

	// calculate the tx hash
	txHash, err := TxHashFromTimeout(uint64(timeoutTimestamp.Unix()), tx)
	if err != nil {
		return err
	}

	// check for duplicates
	if d.txManager.Contains(txHash) {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"tx %X is duplicated",
		)
	}
	if ctx.ExecMode() == sdk.ExecModeFinalize {
		// a new tx included in the block, add the hash to the unordered tx manager
		d.txManager.Add(txHash, timeoutTimestamp)
	}

	return nil
}

// TxHashFromTimeout returns a TxHash for an unordered transaction.
func TxHashFromTimeout(timeout uint64, tx sdk.Tx) (unorderedtx.TxHash, error) {
	sigTx, ok := tx.(authsigning.Tx)
	if !ok {
		return unorderedtx.TxHash{}, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid transaction type")
	}

	if sigTx.GetFee().IsZero() {
		return unorderedtx.TxHash{}, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"unordered transaction must have a fee",
		)
	}

	buf := bufPool.Get().(*bytes.Buffer)
	// Make sure to reset the buffer
	buf.Reset()
	defer bufPool.Put(buf)

	// Add signatures to the transaction identifier
	signatures, err := sigTx.GetSignaturesV2()
	if err != nil {
		return unorderedtx.TxHash{}, err
	}

	for _, sig := range signatures {
		if err := addSignatures(sig.Data, buf); err != nil {
			return unorderedtx.TxHash{}, err
		}
	}

	// Use the buffer
	for _, msg := range tx.GetMsgs() {
		// loop through the messages and write them to the buffer
		// encoding the msg to bytes makes it deterministic within the state machine.
		// Malleability is not a concern here because the state machine will encode the transaction deterministically.
		bz, err := proto.Marshal(msg)
		if err != nil {
			return unorderedtx.TxHash{}, errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"failed to marshal message",
			)
		}

		if _, err := buf.Write(bz); err != nil {
			return unorderedtx.TxHash{}, errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"failed to write message to buffer",
			)
		}
	}

	// write the timeout height to the buffer
	if err := binary.Write(buf, binary.LittleEndian, timeout); err != nil {
		return unorderedtx.TxHash{}, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"failed to write timeout_height to buffer",
		)
	}

	// write gas to the buffer
	if err := binary.Write(buf, binary.LittleEndian, sigTx.GetGas()); err != nil {
		return unorderedtx.TxHash{}, errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"failed to write unordered to buffer",
		)
	}

	txHash := sha256.Sum256(buf.Bytes())

	// Return the Buffer to the pool
	return txHash, nil
}

func addSignatures(sig signing.SignatureData, buf *bytes.Buffer) error {
	switch data := sig.(type) {
	case *signing.SingleSignatureData:
		if _, err := buf.Write(data.Signature); err != nil {
			return errorsmod.Wrap(
				sdkerrors.ErrInvalidRequest,
				"failed to write single signature to buffer",
			)
		}
		return nil

	case *signing.MultiSignatureData:
		for _, sigdata := range data.Signatures {
			if err := addSignatures(sigdata, buf); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected SignatureData %T", data)
	}

	return nil
}
