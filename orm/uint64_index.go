package orm

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

// UInt64IndexerFunc creates one or multiple multiKeyIndex keys of type uint64 for the source object.
type UInt64IndexerFunc func(value interface{}) ([]uint64, error)

// UInt64MultiKeyAdapter converts UInt64IndexerFunc to IndexerFunc
func UInt64MultiKeyAdapter(indexer UInt64IndexerFunc) IndexerFunc {
	return func(value interface{}) ([]RowID, error) {
		d, err := indexer(value)
		if err != nil {
			return nil, err
		}
		r := make([]RowID, len(d))
		for i, v := range d {
			r[i] = EncodeSequence(v)
		}
		return r, nil
	}
}

// UInt64Index is a typed index.
type UInt64Index struct {
	multiKeyIndex MultiKeyIndex
}

// NewUInt64Index creates a typed secondary index
func NewUInt64Index(builder Indexable, prefix byte, indexer UInt64IndexerFunc) UInt64Index {
	return UInt64Index{
		multiKeyIndex: NewIndex(builder, prefix, UInt64MultiKeyAdapter(indexer)),
	}
}

// Has checks if a key exists. Panics on nil key.
func (i UInt64Index) Has(ctx HasKVStore, key uint64) bool {
	return i.multiKeyIndex.Has(ctx, EncodeSequence(key))
}

// Get returns a result iterator for the searchKey. Parameters must not be nil.
func (i UInt64Index) Get(ctx HasKVStore, searchKey uint64) (Iterator, error) {
	return i.multiKeyIndex.Get(ctx, EncodeSequence(searchKey))
}

// GetPaginated creates an iterator for the searchKey
// starting from pageRequest.Key if provided.
// The pageRequest.Key is the rowID while searchKey is a MultiKeyIndex key.
func (i UInt64Index) GetPaginated(ctx HasKVStore, searchKey uint64, pageRequest *query.PageRequest) (Iterator, error) {
	return i.multiKeyIndex.GetPaginated(ctx, EncodeSequence(searchKey), pageRequest)
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
func (i UInt64Index) PrefixScan(ctx HasKVStore, start, end uint64) (Iterator, error) {
	return i.multiKeyIndex.PrefixScan(ctx, EncodeSequence(start), EncodeSequence(end))
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
func (i UInt64Index) ReversePrefixScan(ctx HasKVStore, start, end uint64) (Iterator, error) {
	return i.multiKeyIndex.ReversePrefixScan(ctx, EncodeSequence(start), EncodeSequence(end))
}
