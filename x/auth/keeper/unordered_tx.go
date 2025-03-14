package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
)

var (
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
