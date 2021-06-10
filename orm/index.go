package orm

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// indexer creates and modifies the second MultiKeyIndex based on the operations and changes on the primary object.
type indexer interface {
	OnCreate(store sdk.KVStore, rowID RowID, value interface{}) error
	OnDelete(store sdk.KVStore, rowID RowID, value interface{}) error
	OnUpdate(store sdk.KVStore, rowID RowID, newValue, oldValue interface{}) error
}

// MultiKeyIndex is an index where multiple entries can point to the same underlying object as opposite to a unique index
// where only one entry is allowed.
type MultiKeyIndex struct {
	storeKey      sdk.StoreKey
	prefix        byte
	rowGetter     RowGetter
	indexer       indexer
	indexKeyCodec IndexKeyCodec
}

// NewIndex builds a MultiKeyIndex
func NewIndex(builder Indexable, prefix byte, indexer IndexerFunc) MultiKeyIndex {
	return newIndex(builder, prefix, NewIndexer(indexer, builder.IndexKeyCodec()))
}

func newIndex(builder Indexable, prefix byte, indexer *Indexer) MultiKeyIndex {
	codec := builder.IndexKeyCodec()
	if codec == nil {
		panic("IndexKeyCodec must not be nil")
	}
	storeKey := builder.StoreKey()
	if storeKey == nil {
		panic("StoreKey must not be nil")
	}
	rowGetter := builder.RowGetter()
	if rowGetter == nil {
		panic("RowGetter must not be nil")
	}

	idx := MultiKeyIndex{
		storeKey:      storeKey,
		prefix:        prefix,
		rowGetter:     rowGetter,
		indexer:       indexer,
		indexKeyCodec: codec,
	}
	builder.AddAfterSaveInterceptor(idx.onSave)
	builder.AddAfterDeleteInterceptor(idx.onDelete)
	return idx
}

// Has checks if a key exists. Panics on nil key.
func (i MultiKeyIndex) Has(ctx HasKVStore, key []byte) bool {
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	it := store.Iterator(PrefixRange(key))
	defer it.Close()
	return it.Valid()
}

// Get returns a result iterator for the searchKey. Parameters must not be nil.
func (i MultiKeyIndex) Get(ctx HasKVStore, searchKey []byte) (Iterator, error) {
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	it := store.Iterator(PrefixRange(searchKey))
	return indexIterator{ctx: ctx, it: it, rowGetter: i.rowGetter, keyCodec: i.indexKeyCodec}, nil
}

// GetPaginated creates an iterator for the searchKey
// starting from pageRequest.Key if provided.
// The pageRequest.Key is the rowID while searchKey is a MultiKeyIndex key.
func (i MultiKeyIndex) GetPaginated(ctx HasKVStore, searchKey []byte, pageRequest *query.PageRequest) (Iterator, error) {
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	start, end := PrefixRange(searchKey)

	if pageRequest != nil && len(pageRequest.Key) != 0 {
		start = i.indexKeyCodec.BuildIndexKey(searchKey, RowID(pageRequest.Key))
	}
	it := store.Iterator(start, end)
	return indexIterator{ctx: ctx, it: it, rowGetter: i.rowGetter, keyCodec: i.indexKeyCodec}, nil
}

// PrefixScan returns an Iterator over a domain of keys in ascending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid and error is returned.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a PrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits.
// Example:
//			it, err := idx.PrefixScan(ctx, start, end)
//			if err !=nil {
//				return err
//			}
//			const defaultLimit = 20
//			it = LimitIterator(it, defaultLimit)
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (i MultiKeyIndex) PrefixScan(ctx HasKVStore, start []byte, end []byte) (Iterator, error) {
	if start != nil && end != nil && bytes.Compare(start, end) >= 0 {
		return NewInvalidIterator(), errors.Wrap(ErrArgument, "start must be less than end")
	}
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	it := store.Iterator(start, end)
	return indexIterator{ctx: ctx, it: it, rowGetter: i.rowGetter, keyCodec: i.indexKeyCodec}, nil
}

// ReversePrefixScan returns an Iterator over a domain of keys in descending order. End is exclusive.
// Start is an MultiKeyIndex key or prefix. It must be less than end, or the Iterator is invalid  and error is returned.
// Iterator must be closed by caller.
// To iterate over entire domain, use PrefixScan(nil, nil)
//
// WARNING: The use of a ReversePrefixScan can be very expensive in terms of Gas. Please make sure you do not expose
// this as an endpoint to the public without further limits. See `LimitIterator`
//
// CONTRACT: No writes may happen within a domain while an iterator exists over it.
func (i MultiKeyIndex) ReversePrefixScan(ctx HasKVStore, start []byte, end []byte) (Iterator, error) {
	if start != nil && end != nil && bytes.Compare(start, end) >= 0 {
		return NewInvalidIterator(), errors.Wrap(ErrArgument, "start must be less than end")
	}
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	it := store.ReverseIterator(start, end)
	return indexIterator{ctx: ctx, it: it, rowGetter: i.rowGetter, keyCodec: i.indexKeyCodec}, nil
}

func (i MultiKeyIndex) onSave(ctx HasKVStore, rowID RowID, newValue, oldValue codec.ProtoMarshaler) error {
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	if oldValue == nil {
		return i.indexer.OnCreate(store, rowID, newValue)
	}
	return i.indexer.OnUpdate(store, rowID, newValue, oldValue)
}

func (i MultiKeyIndex) onDelete(ctx HasKVStore, rowID RowID, oldValue codec.ProtoMarshaler) error {
	store := prefix.NewStore(ctx.KVStore(i.storeKey), []byte{i.prefix})
	return i.indexer.OnDelete(store, rowID, oldValue)
}

type UniqueIndex struct {
	MultiKeyIndex
}

// NewUniqueIndex create a new Index object where duplicate keys are prohibited.
func NewUniqueIndex(builder Indexable, prefix byte, uniqueIndexerFunc UniqueIndexerFunc) UniqueIndex {
	return UniqueIndex{
		MultiKeyIndex: newIndex(builder, prefix, NewUniqueIndexer(uniqueIndexerFunc, builder.IndexKeyCodec())),
	}
}

// indexIterator uses rowGetter to lazy load new model values on request.
type indexIterator struct {
	ctx       HasKVStore
	rowGetter RowGetter
	it        types.Iterator
	keyCodec  IndexKeyCodec
}

// LoadNext loads the next value in the sequence into the pointer passed as dest and returns the key. If there
// are no more items the ErrIteratorDone error is returned
// The key is the rowID and not any MultiKeyIndex key.
func (i indexIterator) LoadNext(dest codec.ProtoMarshaler) (RowID, error) {
	if !i.it.Valid() {
		return nil, ErrIteratorDone
	}
	indexPrefixKey := i.it.Key()
	rowID := i.keyCodec.StripRowID(indexPrefixKey)
	i.it.Next()
	return rowID, i.rowGetter(i.ctx, rowID, dest)
}

// Close releases the iterator and should be called at the end of iteration
func (i indexIterator) Close() error {
	i.it.Close()
	return nil
}

// PrefixRange turns a prefix into a (start, end) range. The start is the given prefix value and
// the end is calculated by adding 1 bit to the start value. Nil is not allowed as prefix.
// 		Example: []byte{1, 3, 4} becomes []byte{1, 3, 5}
// 				 []byte{15, 42, 255, 255} becomes []byte{15, 43, 0, 0}
//
// In case of an overflow the end is set to nil.
//		Example: []byte{255, 255, 255, 255} becomes nil
//
func PrefixRange(prefix []byte) ([]byte, []byte) {
	if prefix == nil {
		panic("nil key not allowed")
	}
	// special case: no prefix is whole range
	if len(prefix) == 0 {
		return nil, nil
	}

	// copy the prefix and update last byte
	end := make([]byte, len(prefix))
	copy(end, prefix)
	l := len(end) - 1
	end[l]++

	// wait, what if that overflowed?....
	for end[l] == 0 && l > 0 {
		l--
		end[l]++
	}

	// okay, funny guy, you gave us FFF, no end to this range...
	if l == 0 && end[0] == 0 {
		end = nil
	}
	return prefix, end
}
