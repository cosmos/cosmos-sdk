# ADR 070: Unordered Transactions

## Changelog

- Dec 4, 2023: Initial Draft (@yihuang, @tac0turtle, @alexanderbez)
- Jan 30, 2024: Include section on deterministic transaction encoding
- Mar 12, 2025: Revise implementation to use Cosmos SDK KV Store and require unique timeouts per-address (@technicallyty)

## Status

Draft

## Abstract

We propose a way to do replay-attack protection without enforcing the order of
transactions and without requiring the use of monotonically increasing sequences. Instead, we propose
the use of a time-based, ephemeral sequence.

## Context

Account sequence values serve to prevent replay-attacks and ensure transactions from the same sender are included into blocks and executed
in sequential order. Unfortunately, this makes it difficult to reliably send many concurrent transactions from the
same sender. Victims of such limitations include IBC relayers and crypto exchanges.

## Decision

We propose adding a boolean field `unordered` and a uint64 field `timeout_timestamp` to the transaction body.

Unordered transactions will bypass the traditional account sequence rules and follow the rules described
below, without impacting traditional ordered transactions; they'll follow the sequence rules the same as before.

We will introduce new storage of time-based, ephemeral unordered sequences using the SDK's existing KV Store library.

When an unordered transaction is included in a block, a concatenation of the `timeout_timestamp` and sender’s bech32 address
will be recorded to state (i.e. `542939323/cosmos1v1234567890AbcDeF`). In cases of multi-party signing, we will use a
comma-separated list of the addresses that signed the transaction (i.e. `5532231/cosmosv11,cosmosv12,cosmosv13`)

New transactions will be checked against the state to prevent duplicate submissions. To prevent the state from growing indefinitely, we propose the following:

- Define an upper bound for the value of `timeout_timestamp` (i.e. 10 minutes).
- Extend the PreBlocker method to remove state entries with a `timeout_timestamp` earlier than the current block time.

### Transaction Format

```protobuf
message TxBody {
  ...
          
  bool unordered = 4;
  google.protobuf.Timestamp timeout_timestamp = 5
}
```

### Replay Protection

We facilitate replay protection by storing the unordered sequence, in the Cosmos SDK KV store. Upon transaction ingress, we check if the transaction's unordered
sequence exists in state, or if the TTL value is stale, i.e. before the current block time. If so, we reject it. Otherwise,
we add the unordered sequence to state. This section of the state will belong to the `x/auth` module.

The state is evaluated during x/auth's `PreBlocker`. All transactions with an unordered sequence earlier than the current block time
will be deleted.

```go
func (am AppModule) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	err := am.accountKeeper.GetUnorderedTxManager().RemoveExpired(sdk.UnwrapSDKContext(ctx))
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

type UnorderedTxManager struct {
	unorderedSequences collections.KeySet[collections.Pair[uint64, string]]
}

func NewUnorderedTxManager(kvStore store.KVStoreService) *UnorderedTxManager {
	sb := collections.NewSchemaBuilder(kvStore)
	m := &UnorderedTxManager{
		unorderedSequences: collections.NewKeySet(
			sb,
			unorderedSequencePrefix,
			"unordered_sequences",
			collections.PairKeyCodec(collections.Uint64Key, collections.StringKey),
		),
	}
	return m
}

func (m *UnorderedTxManager) Contains(ctx sdk.Context, sender string, timestamp uint64) (bool, error) {
	return m.unorderedSequences.Has(ctx, collections.Join(timestamp, sender))
}

func (m *UnorderedTxManager) Add(ctx sdk.Context, sender string, timestamp uint64) error {
	return m.unorderedSequences.Set(ctx, collections.Join(timestamp, sender))
}

func (m *UnorderedTxManager) RemoveExpired(ctx sdk.Context) error {
	blkTime := ctx.BlockTime().UnixNano()
	it, err := m.unorderedSequences.Iterate(ctx, collections.NewPrefixUntilPairRange[uint64, string](uint64(blkTime)))
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

In order to facilitate bypassing nonce verification, we have to modify the existing
`IncrementSequenceDecorator` AnteHandler decorator to skip the nonce verification
when the transaction is marked as un-ordered.

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
// chain to ensure that during DeliverTx, the transaction is added to the UnorderedTxManager.
type UnorderedTxDecorator struct {
	// maxUnOrderedTTL defines the maximum TTL a transaction can define.
	maxTimeoutDuration time.Duration
	txManager          *authkeeper.UnorderedTxManager
}

func NewUnorderedTxDecorator(
	utxm *authkeeper.UnorderedTxManager,
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
	slices.Sort(signerAddrs)
	signers := strings.Join(signerAddrs, ",")

	contains, err := d.txManager.Contains(ctx, signers, uint64(unorderedTx.GetTimeoutTimeStamp().Unix()))
	if err != nil {
		return errorsmod.Wrap(
			sdkerrors.ErrIO,
			"failed to check contains",
		)
	}
	if contains {
		return errorsmod.Wrap(
			sdkerrors.ErrInvalidRequest,
			"tx is duplicated",
		)
	}

	if err := d.txManager.Add(ctx, signers, uint64(unorderedTx.GetTimeoutTimeStamp().Unix())); err != nil {
		return errorsmod.Wrap(
			sdkerrors.ErrIO,
			"failed to add unordered nonce to state",
		)
	}

	return nil
}

func getSigners(tx sdk.Tx) ([]string, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return nil, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}
	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return nil, err
	}

	addresses := make([]string, 0, len(sigs))
	for _, sig := range sigs {
		addresses = append(addresses, sig.PubKey.Address().String())
	}

	return addresses, nil
}

```

### Unordered Sequences

Unordered sequences provide a simple, straightforward mechanism to protect against both transaction malleability and
transaction duplication. It is important to note, however, that the unordered sequence must still be unique, however
the value is not required to be strictly increasing as with regular sequences, and the order in which the node receives
the transactions no longer matters. Clients can handle setting timeouts similarly to the code below:

```go
for _, tx := range txs {
	tx.SetUnordered(true)
	tx.SetTimeoutTimestamp(time.Now() + 1 * time.Nanosecond)
}
```

### State Management

The storage of unordered sequences will be facilitated using the Cosmos SDK's KV Store service.

## Note On Previous Iteration

The previous iteration of unordered transactions worked by using an ad-hoc state-management system that posed severe 
risks and a vector for duplicated tx processing. It relied on graceful app closure which would flush the current state
of the unordered sequence mapping. If the 2/3's of the network crashed, and the graceful closure did not trigger, 
the system would lose track of all sequences in the mapping, allowing those transactions to be replayed. The 
implementation proposed in the updated version of this ADR solves this by writing directly to the Cosmos KV Store.

Additionally, the previous iteration relied on using hashes to create what we call an "unordered sequence." 

## Consequences

* Usage of Cosmos SDK KV store is slower in comparison to using a non merklized store or ad-hoc methods, and block times may slow down as a result.

### Positive

* Support unordered transaction inclusion, enabling the ability to "fire and forget" many transactions at once.

### Negative

* Requires additional storage overhead.
* Requirement of unique timestamps per transaction causes a small amount of additional overhead for clients. Clients must ensure each transaction's timeout timestamp is different. However, nanosecond differentials suffice.

## References

* https://github.com/cosmos/cosmos-sdk/issues/13009

