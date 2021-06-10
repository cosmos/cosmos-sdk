package orm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
)

// IndexerFunc creates one or multiple index keys for the source object.
type IndexerFunc func(value interface{}) ([]RowID, error)

// IndexerFunc creates exactly one index key for the source object.
type UniqueIndexerFunc func(value interface{}) (RowID, error)

// Indexer manages the persistence for an Index based on searchable keys and operations.
type Indexer struct {
	indexerFunc   IndexerFunc
	addFunc       func(store sdk.KVStore, codec IndexKeyCodec, secondaryIndexKey []byte, rowID RowID) error
	indexKeyCodec IndexKeyCodec
}

// NewIndexer returns an indexer that supports multiple reference keys for an entity.
func NewIndexer(indexerFunc IndexerFunc, codec IndexKeyCodec) *Indexer {
	if indexerFunc == nil {
		panic("Indexer func must not be nil")
	}
	if codec == nil {
		panic("IndexKeyCodec must not be nil")
	}
	return &Indexer{
		indexerFunc:   pruneEmptyKeys(indexerFunc),
		addFunc:       multiKeyAddFunc,
		indexKeyCodec: codec,
	}
}

// NewUniqueIndexer returns an indexer that requires exactly one reference keys for an entity.
func NewUniqueIndexer(f UniqueIndexerFunc, codec IndexKeyCodec) *Indexer {
	if f == nil {
		panic("indexer func must not be nil")
	}
	adaptor := func(indexerFunc UniqueIndexerFunc) IndexerFunc {
		return func(v interface{}) ([]RowID, error) {
			k, err := indexerFunc(v)
			return []RowID{k}, err
		}
	}
	idx := NewIndexer(adaptor(f), codec)
	idx.addFunc = uniqueKeysAddFunc
	return idx
}

// OnCreate persists the secondary index entries for the new object.
func (i Indexer) OnCreate(store sdk.KVStore, rowID RowID, value interface{}) error {
	secondaryIndexKeys, err := i.indexerFunc(value)
	if err != nil {
		return err
	}

	for _, secondaryIndexKey := range secondaryIndexKeys {
		if err := i.addFunc(store, i.indexKeyCodec, secondaryIndexKey, rowID); err != nil {
			return err
		}
	}
	return nil
}

// OnDelete removes the secondary index entries for the deleted object.
func (i Indexer) OnDelete(store sdk.KVStore, rowID RowID, value interface{}) error {
	secondaryIndexKeys, err := i.indexerFunc(value)
	if err != nil {
		return err
	}

	for _, secondaryIndexKey := range secondaryIndexKeys {
		indexKey := i.indexKeyCodec.BuildIndexKey(secondaryIndexKey, rowID)
		store.Delete(indexKey)
	}
	return nil
}

// OnUpdate rebuilds the secondary index entries for the updated object.
func (i Indexer) OnUpdate(store sdk.KVStore, rowID RowID, newValue, oldValue interface{}) error {
	oldSecIdxKeys, err := i.indexerFunc(oldValue)
	if err != nil {
		return err
	}
	newSecIdxKeys, err := i.indexerFunc(newValue)
	if err != nil {
		return err
	}
	for _, oldIdxKey := range difference(oldSecIdxKeys, newSecIdxKeys) {
		store.Delete(i.indexKeyCodec.BuildIndexKey(oldIdxKey, rowID))
	}
	for _, newIdxKey := range difference(newSecIdxKeys, oldSecIdxKeys) {
		if err := i.addFunc(store, i.indexKeyCodec, newIdxKey, rowID); err != nil {
			return err
		}
	}
	return nil
}

// uniqueKeysAddFunc enforces keys to be unique
func uniqueKeysAddFunc(store sdk.KVStore, codec IndexKeyCodec, secondaryIndexKey []byte, rowID RowID) error {
	if len(secondaryIndexKey) == 0 {
		return errors.Wrap(ErrArgument, "empty index key")
	}
	it := store.Iterator(PrefixRange(secondaryIndexKey))
	defer it.Close()
	if it.Valid() {
		return ErrUniqueConstraint
	}
	indexKey := codec.BuildIndexKey(secondaryIndexKey, rowID)
	store.Set(indexKey, []byte{})
	return nil
}

// multiKeyAddFunc allows multiple entries for a key
func multiKeyAddFunc(store sdk.KVStore, codec IndexKeyCodec, secondaryIndexKey []byte, rowID RowID) error {
	if len(secondaryIndexKey) == 0 {
		return errors.Wrap(ErrArgument, "empty index key")
	}

	indexKey := codec.BuildIndexKey(secondaryIndexKey, rowID)
	store.Set(indexKey, []byte{})
	return nil
}

// difference returns the list of elements that are in a but not in b.
func difference(a []RowID, b []RowID) []RowID {
	set := make(map[string]struct{}, len(b))
	for _, v := range b {
		set[string(v)] = struct{}{}
	}
	var result []RowID
	for _, v := range a {
		if _, ok := set[string(v)]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// pruneEmptyKeys drops any empty key from IndexerFunc f returned
func pruneEmptyKeys(f IndexerFunc) IndexerFunc {
	return func(v interface{}) ([]RowID, error) {
		keys, err := f(v)
		if err != nil || keys == nil {
			return keys, err
		}
		r := make([]RowID, 0, len(keys))
		for i := range keys {
			if len(keys[i]) != 0 {
				r = append(r, keys[i])
			}
		}
		return r, nil
	}
}
