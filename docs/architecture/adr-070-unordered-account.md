# ADR 070: Unordered Transactions

## Changelog

- Dec 4, 2023: Initial Draft (@yihuang, @tac0turtle, @alexanderbez)
- Jan 30, 2024: Include section on deterministic transaction encoding
- Mar 18, 2025: Revise implementation to use Cosmos SDK KV Store and require unique timeouts per-address (@technicallyty)
- Apr 25, 2025: Add note about rejecting unordered txs with sequence values.

## Status

ACCEPTED Not Implemented

## Abstract

We propose a way to do replay-attack protection without enforcing the order of
transactions and without requiring the use of monotonically increasing sequences. Instead, we propose
the use of a time-based, ephemeral sequence.

## Context

Account sequence values serve to prevent replay attacks and ensure transactions from the same sender are included into blocks and executed
in sequential order. Unfortunately, this makes it difficult to reliably send many concurrent transactions from the
same sender. Victims of such limitations include IBC relayers and crypto exchanges.

## Decision

We propose adding a boolean field `unordered` and a google.protobuf.Timestamp field `timeout_timestamp` to the transaction body.

Unordered transactions will bypass the traditional account sequence rules and follow the rules described
below, without impacting traditional ordered transactions which will follow the same sequence rules as before.

We will introduce new storage of time-based, ephemeral unordered sequences using the SDK's existing KV Store library. 
Specifically, we will leverage the existing x/auth KV store to store the unordered sequences.

When an unordered transaction is included in a block, a concatenation of the `timeout_timestamp` and sender’s address bytes
will be recorded to state (i.e. `542939323/<address_bytes>`). In cases of multi-party signing, one entry per signer
will be recorded to state.

New transactions will be checked against the state to prevent duplicate submissions. To prevent the state from growing indefinitely, we propose the following:

- Define an upper bound for the value of `timeout_timestamp` (i.e. 10 minutes).
- Add PreBlocker method x/auth that removes state entries with a `timeout_timestamp` earlier than the current block time.

### Transaction Format

```protobuf
message TxBody {
  ...
          
  bool unordered = 4;
  google.protobuf.Timestamp timeout_timestamp = 5
}
```

### Replay Protection

We facilitate replay protection by storing the unordered sequence in the Cosmos SDK KV store. Upon transaction ingress, we check if the transaction's unordered
sequence exists in state, or if the TTL value is stale, i.e. before the current block time. If so, we reject it. Otherwise,
we add the unordered sequence to the state. This section of the state will belong to the `x/auth` module.

The state is evaluated during x/auth's `PreBlocker`. All transactions with an unordered sequence earlier than the current block time
will be deleted.

```go
func (am AppModule) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	err := am.accountKeeper.RemoveExpired(sdk.UnwrapSDKContext(ctx))
	if err != nil {
		return nil, err
	}
	return &sdk.ResponsePreBlock{ConsensusParamsChanged: false}, nil
}
```

```golang
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
)

var (
	// just arbitrarily picking some upper bound number.
	unorderedSequencePrefix = collections.NewPrefix(90)
)

type AccountKeeper struct {
	// ...
	unorderedSequences collections.KeySet[collections.Pair[uint64, []byte]]
}

func (m *AccountKeeper) Contains(ctx sdk.Context, sender []byte, timestamp uint64) (bool, error) {
	return m.unorderedSequences.Has(ctx, collections.Join(timestamp, sender))
}

func (m *AccountKeeper) Add(ctx sdk.Context, sender []byte, timestamp uint64) error {
	return m.unorderedSequences.Set(ctx, collections.Join(timestamp, sender))
}

func (m *AccountKeeper) RemoveExpired(ctx sdk.Context) error {
	blkTime := ctx.BlockTime().UnixNano()
	it, err := m.unorderedSequences.Iterate(ctx, collections.NewPrefixUntilPairRange[uint64, []byte](uint64(blkTime)))
	if err != nil {
		return err
	}
	defer it.Close()

	keys, err := it.Keys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := m.unorderedSequences.Remove(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

```

### AnteHandler Decorator

To facilitate bypassing nonce verification, we must modify the existing
`IncrementSequenceDecorator` AnteHandler decorator to skip the nonce verification
when the transaction is marked as unordered.

```golang
func (isd IncrementSequenceDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
  if tx.UnOrdered() {
    return next(ctx, tx, simulate)
  }

  // ...
}
```

We also introduce a new decorator to perform the unordered transaction verification.

```golang
package ante

import (
	"slices"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	errorsmod "cosmossdk.io/errors"
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
// chain to ensure that during DeliverTx, the transaction is added to the unordered sequence state.
type UnorderedTxDecorator struct {
	// maxUnOrderedTTL defines the maximum TTL a transaction can define.
	maxTimeoutDuration time.Duration
	txManager          authkeeper.UnorderedTxManager
}

func NewUnorderedTxDecorator(
	utxm authkeeper.UnorderedTxManager,
) *UnorderedTxDecorator {
	return &UnorderedTxDecorator{
		maxTimeoutDuration: 10 * time.Minute,
		txManager:          utxm,
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

	execMode := ctx.ExecMode()
	if execMode == sdk.ExecModeSimulate {
		return nil
	}

	signerAddrs, err := getSigners(tx)
	if err != nil {
		return err
	}
	
	for _, signer := range signerAddrs {
		contains, err := d.txManager.Contains(ctx, signer, uint64(unorderedTx.GetTimeoutTimeStamp().Unix()))
		if err != nil {
			return errorsmod.Wrap(
				sdkerrors.ErrIO,
				"failed to check contains",
			)
		}
		if contains {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"tx is duplicated for signer %x", signer,
			)
		}

		if err := d.txManager.Add(ctx, signer, uint64(unorderedTx.GetTimeoutTimeStamp().Unix())); err != nil {
			return errorsmod.Wrap(
				sdkerrors.ErrIO,
				"failed to add unordered sequence to state",
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
	return sigTx.GetSigners()
}

```

### Unordered Sequences

Unordered sequences provide a simple, straightforward mechanism to protect against both transaction malleability and
transaction duplication. It is important to note that the unordered sequence must still be unique. However,
the value is not required to be strictly increasing as with regular sequences, and the order in which the node receives
the transactions no longer matters. Clients can handle building unordered transactions similarly to the code below:

```go
for _, tx := range txs {
	tx.SetUnordered(true)
	tx.SetTimeoutTimestamp(time.Now() + 1 * time.Nanosecond)
}
```

We will reject transactions that have both sequence and unordered timeouts set. We do this to avoid assuming the intent of the user.

### State Management

The storage of unordered sequences will be facilitated using the Cosmos SDK's KV Store service.

## Note On Previous Design Iteration

The previous iteration of unordered transactions worked by using an ad-hoc state-management system that posed severe 
risks and a vector for duplicated tx processing. It relied on graceful app closure which would flush the current state
of the unordered sequence mapping. If the 2/3's of the network crashed, and the graceful closure did not trigger, 
the system would lose track of all sequences in the mapping, allowing those transactions to be replayed. The 
implementation proposed in the updated version of this ADR solves this by writing directly to the Cosmos KV Store.
While this is less performant, for the initial implementation, we opted to choose a safer path and postpone performance optimizations until we have more data on real-world impacts and a more battle-tested approach to optimization.

Additionally, the previous iteration relied on using hashes to create what we call an "unordered sequence." There are known
issues with transaction malleability in Cosmos SDK signing modes. This ADR gets away from this problem by enforcing
single-use unordered nonces, instead of deriving nonces from bytes in the transaction.

## Consequences

### Positive

* Support unordered transaction inclusion, enabling the ability to "fire and forget" many transactions at once.

### Negative

* Requires additional storage overhead.
* Requirement of unique timestamps per transaction causes a small amount of additional overhead for clients. Clients must ensure each transaction's timeout timestamp is different. However, nanosecond differentials suffice.
* Usage of Cosmos SDK KV store is slower in comparison to using a non-merklized store or ad-hoc methods, and block times may slow down as a result.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009

