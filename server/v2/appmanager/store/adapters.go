package store

import (
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/server/v2/core/store"
	storev2 "cosmossdk.io/store/v2"
)

// iterAdapter adapts a storev2.Iterator into a corestore.Iterator
type iterAdapter struct {
	iter corestore.Iterator
}

func (i iterAdapter) Domain() (start, end []byte) {
	return i.iter.Domain()
}

func (i iterAdapter) Valid() bool {
	return i.iter.Valid()
}

func (i iterAdapter) Next() {
	i.iter.Next()
}

func (i iterAdapter) Key() (key []byte) {
	return i.iter.Key()
}

func (i iterAdapter) Value() (value []byte) {
	return i.iter.Value()
}

func (i iterAdapter) Error() error {
	return i.iter.Error()
}

func (i iterAdapter) Close() error {
	i.iter.Close()
	return nil
}

func newIterAdapter(iter corestore.Iterator) corestore.Iterator {
	return iterAdapter{iter: iter}
}

func intoStoreV2ChangeSet(changes []store.ChangeSet) *storev2.Changeset {
	kvPairs := make([]storev2.KVPair, len(changes))
	for i, c := range changes {
		kvPairs[i] = storev2.KVPair{
			Key:   c.Key,
			Value: c.Value,
		}
	}
	return &storev2.Changeset{
		Pairs: map[string]storev2.KVPairs{
			"": kvPairs,
		},
	}
}
