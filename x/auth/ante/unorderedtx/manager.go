package unorderedtx

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/core/store"
)

const (
	// DefaultMaxTimeoutDuration defines the default maximum duration an unordered transaction
	// can set.
	DefaultMaxTimeoutDuration = time.Minute * 40
)

type UnorderedSequence string

func NewUnorderedSequence(addr string, timestamp uint64) UnorderedSequence {
	return UnorderedSequence(fmt.Sprintf("%d/%s", timestamp, addr))
}

// Manager contains the tx hash dictionary for duplicates checking, and expire
// them when block production progresses.
type Manager struct {
	kvStore store.KVStoreService
}

func NewManager(kvStore store.KVStoreService) *Manager {
	m := &Manager{
		kvStore: kvStore,
	}

	return m
}

func (m *Manager) Contains(ctx sdk.Context, sender string, timestamp uint64) (bool, error) {
	return m.kvStore.OpenKVStore(ctx).Has([]byte(NewUnorderedSequence(sender, timestamp)))
}

func (m *Manager) Add(ctx sdk.Context, sender string, timestamp uint64) error {
	return m.kvStore.OpenKVStore(ctx).Set([]byte(NewUnorderedSequence(sender, timestamp)), []byte(byte(0x0)))
}

func (m *Manager) RemoveExpired(ctx sdk.Context) error {
	kvStore := m.kvStore.OpenKVStore(ctx)
	it, err := kvStore.Iterator(nil, []byte(NewUnorderedSequence("", uint64(ctx.BlockTime().Unix()))))
	if err != nil {
		return err
	}
	defer it.Close()

	keys := make([][]byte, 0)
	for ; it.Valid(); it.Next() {
		keys = append(keys, it.Key())
	}

	for _, key := range keys {
		err := kvStore.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
