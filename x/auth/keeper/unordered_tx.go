package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
)

var (
	// just arbitrarily picking some upper bound number.
	unorderedSequencePrefix = collections.NewPrefix(90)
)

// UnorderedTxManager manages the ephemeral timeout sequences from unordered transactions.
type UnorderedTxManager interface {
	// Contains reports whether the sender has used the timestamp already.
	Contains(ctx sdk.Context, sender []byte, timestamp time.Time) (bool, error)
	// Add marks the timestamp as used for the sender.
	// Further transactions sent from this sender with this timestamp will fail.
	Add(ctx sdk.Context, sender []byte, timestamp time.Time) error
	// RemoveExpired removes all sequences whose timestamp value is before the current block time.
	RemoveExpired(ctx sdk.Context) error
}

type unorderedTxManagerImpl struct {
	unorderedSequences collections.KeySet[collections.Pair[uint64, []byte]]
}

func NewUnorderedTxManager(kvStore store.KVStoreService) UnorderedTxManager {
	sb := collections.NewSchemaBuilder(kvStore)
	m := &unorderedTxManagerImpl{
		unorderedSequences: collections.NewKeySet(
			sb,
			unorderedSequencePrefix,
			"unordered_sequences",
			collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey),
		),
	}
	return m
}

func (m *unorderedTxManagerImpl) Contains(ctx sdk.Context, sender []byte, timestamp time.Time) (bool, error) {
	return m.unorderedSequences.Has(ctx, collections.Join(uint64(timestamp.UnixNano()), sender))
}

func (m *unorderedTxManagerImpl) Add(ctx sdk.Context, sender []byte, timestamp time.Time) error {
	return m.unorderedSequences.Set(ctx, collections.Join(uint64(timestamp.UnixNano()), sender))
}

func (m *unorderedTxManagerImpl) RemoveExpired(ctx sdk.Context) error {
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
