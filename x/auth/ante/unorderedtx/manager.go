package unorderedtx

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
)

const (
	// DefaultMaxTimeoutDuration defines the default maximum duration an unordered transaction
	// can set.
	DefaultMaxTimeoutDuration = time.Minute * 40
)

var (
	unorderedSequencePrefix = collections.NewPrefix(0)
)

type UnorderedSequence []byte

func NewUnorderedSequence(addr string, timestamp uint64) UnorderedSequence {
	tsBz := sdk.Uint64ToBigEndian(timestamp)
	tsBz = append(tsBz, []byte("/"+addr)...)
	return tsBz
}

func (u UnorderedSequence) String() string {
	return string(u)
}

// Manager contains the tx hash dictionary for duplicates checking, and expire
// them when block production progresses.
type Manager struct {
	unorderedSequences collections.KeySet[[]byte]
}

func NewManager(kvStore store.KVStoreService) *Manager {
	sb := collections.NewSchemaBuilder(kvStore)

	m := &Manager{
		unorderedSequences: collections.NewKeySet(
			sb,
			unorderedSequencePrefix,
			"unordered_sequences",
			collections.BytesKey,
		),
	}

	return m
}

func (m *Manager) Contains(ctx sdk.Context, sender string, timestamp uint64) (bool, error) {
	return m.unorderedSequences.Has(ctx, NewUnorderedSequence(sender, timestamp))
}

func (m *Manager) Add(ctx sdk.Context, sender string, timestamp uint64) error {
	return m.unorderedSequences.Set(ctx, NewUnorderedSequence(sender, timestamp))
}

func (m *Manager) RemoveExpired(ctx sdk.Context) error {

	keyEnd := sdk.Uint64ToBigEndian(uint64(ctx.BlockTime().UnixNano()))
	keyEnd = append(keyEnd, []byte("/")...)
	it, err := m.unorderedSequences.IterateRaw(ctx, nil, keyEnd, collections.OrderAscending)
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
