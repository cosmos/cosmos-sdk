package ante

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"sync"

	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/x/auth/ante/unorderedtx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/protobuf/proto"
)

// bufPool is a pool of bytes.Buffer objects to reduce memory allocations.
var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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
	env             appmodule.Environment
}

func NewUnorderedTxDecorator(maxTTL uint64, m *unorderedtx.Manager, env appmodule.Environment) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxUnOrderedTTL: maxTTL,
		txManager:       m,
		env:             env,
	}
}

func (d *UnorderedTxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, _ bool, next sdk.AnteHandler) (sdk.Context, error) {
	unorderedTx, ok := tx.(sdk.TxWithUnordered)
	if !ok || !unorderedTx.GetUnordered() {
		// If the transaction does not implement unordered capabilities or has the
		// unordered value as false, we bypass.
		return next(ctx, tx, false)
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

	// in order to create a deterministic hash based on the tx, we need to hash the contents of the tx with signature
	// Get a Buffer from the pool
	buf := bufPool.Get().(*bytes.Buffer)
	// Make sure to reset the buffer
	buf.Reset()

	// Use the buffer
	for _, msg := range tx.GetMsgs() {
		// loop through the messages and write them to the buffer
		// encoding the msg to bytes makes it deterministic within the state machine.
		// Malleability is not a concern here because the state machine will encode the transaction deterministically.
		bz, err := proto.Marshal(msg)
		if err != nil {
			return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to marshal message")
		}

		buf.Write(bz)
	}

	if err := binary.Write(buf, binary.LittleEndian, unorderedTx.GetTimeoutHeight()); err != nil {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "failed to write timeout_height to buffer")
	}

	txHash := sha256.Sum256(buf.Bytes())

	// Return the Buffer to the pool
	bufPool.Put(buf)

	// check for duplicates
	if d.txManager.Contains(txHash) {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "tx %X is duplicated")
	}

	if d.env.TransactionService.ExecMode(ctx) == transaction.ExecModeFinalize {
		// a new tx included in the block, add the hash to the unordered tx manager
		d.txManager.Add(txHash, ttl)
	}

	return next(ctx, tx, false)
}
