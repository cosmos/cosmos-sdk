package collections

import (
	"context"
	"fmt"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Order defines the key order.
type Order uint8

const (
	// OrderAscending instructs the Iterator to provide keys from the smallest to the greatest.
	OrderAscending Order = 0
	// OrderDescending instructs the Iterator to provide keys from the greatest to the smallest.
	OrderDescending Order = 1
)

// BoundInclusive creates a Bound of the provided key K
// which is inclusive. Meaning, if it is used as Ranger.RangeValues start,
// the provided key will be included if it exists in the Iterator range.
func BoundInclusive[K any](key K) *Bound[K] {
	return &Bound[K]{
		value:     key,
		inclusive: true,
	}
}

// BoundExclusive creates a Bound of the provided key K
// which is exclusive. Meaning, if it is used as Ranger.RangeValues start,
// the provided key will be excluded if it exists in the Iterator range.
func BoundExclusive[K any](key K) *Bound[K] {
	return &Bound[K]{
		value:     key,
		inclusive: false,
	}
}

// Bound defines key bounds for Start and Ends of iterator ranges.
type Bound[K any] struct {
	value     K
	inclusive bool
}

// Ranger defines a generic interface that provides a range of keys.
type Ranger[K any] interface {
	// RangeValues is defined by Ranger implementers.
	// It provides instructions to generate an Iterator instance.
	// If prefix is not nil, then the Iterator will return only the keys which start
	// with the given prefix.
	// If start is not nil, then the Iterator will return only keys which are greater than the provided start
	// or greater equal depending on the bound is inclusive or exclusive.
	// If end is not nil, then the Iterator will return only keys which are smaller than the provided end
	// or smaller equal depending on the bound is inclusive or exclusive.
	RangeValues() (prefix *K, start *Bound[K], end *Bound[K], order Order)
}

// Range is a Ranger implementer.
type Range[K any] struct {
	prefix *K
	start  *Bound[K]
	end    *Bound[K]
	order  Order
}

// Prefix sets a fixed prefix for the key range.
func (r *Range[K]) Prefix(key K) *Range[K] {
	r.prefix = &key
	return r
}

// StartInclusive makes the range contain only keys which are bigger or equal to the provided start K.
func (r *Range[K]) StartInclusive(start K) *Range[K] {
	r.start = BoundInclusive(start)
	return r
}

// StartExclusive makes the range contain only keys which are bigger to the provided start K.
func (r *Range[K]) StartExclusive(start K) *Range[K] {
	r.start = BoundExclusive(start)
	return r
}

// EndInclusive makes the range contain only keys which are smaller or equal to the provided end K.
func (r *Range[K]) EndInclusive(end K) *Range[K] {
	r.end = BoundInclusive(end)
	return r
}

// EndExclusive makes the range contain only keys which are smaller to the provided end K.
func (r *Range[K]) EndExclusive(end K) *Range[K] {
	r.end = BoundExclusive(end)
	return r
}

func (r *Range[K]) Descending() *Range[K] {
	r.order = OrderDescending
	return r
}

func (r *Range[K]) RangeValues() (prefix *K, start *Bound[K], end *Bound[K], order Order) {
	return r.prefix, r.start, r.end, r.order
}

// iteratorFromRanger generates an Iterator instance, with the proper prefixing and ranging.
// a nil Ranger can be seen as an ascending iteration over all the possible keys.
func iteratorFromRanger[K, V any](ctx context.Context, m Map[K, V], r Ranger[K]) (iter Iterator[K, V], err error) {
	// defaults
	var (
		prefix *K
		start  *Bound[K]
		end    *Bound[K]
		order  = OrderAscending
	)
	// override defaults only if a Ranger is provided.
	if r != nil {
		prefix, start, end, order = r.RangeValues()
	}
	var prefixBytes []byte
	if prefix != nil {
		prefixBytes, err = encodeKeyWithPrefix(m.prefix, m.kc, *prefix)
		if err != nil {
			return iter, err
		}
	} else {
		prefixBytes = m.prefix
	}
	var startBytes []byte // default is nil
	if start != nil {
		startBytes, err = encodeKeyWithPrefix(nil, m.kc, start.value)
		if err != nil {
			return
		}
		// iterators are inclusive at start by default
		// so if we want to make the iteration exclusive
		// we extend by one byte.
		if !start.inclusive {
			startBytes = extendOneByte(startBytes)
		}
	}
	var endBytes []byte // default is nil
	if end != nil {
		endBytes, err = encodeKeyWithPrefix(nil, m.kc, end.value)
		if err != nil {
			return iter, err
		}
		// iterators are exclusive at end by default
		// so if we want to make the iteration
		// inclusive we need to extend by one byte.
		if end.inclusive {
			endBytes = extendOneByte(endBytes)
		}
	}

	store, err := m.getStore(ctx)
	if err != nil {
		return iter, err
	}
	prefixedStartBytes := append(prefixBytes, startBytes...)
	// since we're dealing with a prefixed end, if end bytes
	// are not specified then we need to increase the prefix
	// by 1, otherwise we would end up ranging over [bytes, bytes),
	// which is invalid.
	var prefixedEndBytes []byte
	if endBytes == nil {
		prefixedEndBytes = storetypes.PrefixEndBytes(prefixBytes)
	} else {
		prefixedEndBytes = append(prefixBytes, endBytes...)
	}

	var storeIter storetypes.Iterator
	switch order {
	case OrderAscending:
		storeIter = store.Iterator(prefixedStartBytes, prefixedEndBytes)
	case OrderDescending:
		storeIter = store.ReverseIterator(prefixedStartBytes, prefixedEndBytes)
	default:
		return iter, fmt.Errorf("collections: unrecognized order identifier: %d", order)
	}

	iter.kc = m.kc
	iter.vc = m.vc
	iter.iter = storeIter
	iter.prefixLength = len(m.prefix)

	return iter, nil
}

// Iterator defines a generic wrapper around an sdk.Iterator.
// This iterator provides automatic key and value encoding,
// it assumes all the keys and values contained within the sdk.Iterator
// range are the same.
type Iterator[K, V any] struct {
	kc KeyCodec[K]
	vc ValueCodec[V]

	iter storetypes.Iterator

	prefixLength int
}

// Value returns the current iterator value bytes decoded.
func (i Iterator[K, V]) Value() (V, error) {
	return i.vc.Decode(i.iter.Value())
}

// Key returns the current sdk.Iterator decoded key.
func (i Iterator[K, V]) Key() (K, error) {
	bytesKey := i.iter.Key()[i.prefixLength:] // strip prefix namespace

	read, key, err := i.kc.Decode(bytesKey)
	if err != nil {
		var k K
		return k, err
	}
	if read != len(bytesKey) {
		var k K
		return k, fmt.Errorf("%w: key decoder didn't fully consume the key: %T %x %d", ErrEncoding, i.kc, bytesKey, read)
	}
	return key, nil
}

// Values fully consumes the iterator and returns all the decoded values contained within the range.
func (i Iterator[K, V]) Values() ([]V, error) {
	defer i.Close()

	var values []V
	for ; i.iter.Valid(); i.iter.Next() {
		value, err := i.Value()
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// Keys fully consumes the iterator and returns all the decoded keys contained within the range.
func (i Iterator[K, V]) Keys() ([]K, error) {
	defer i.Close()

	var keys []K
	for ; i.iter.Valid(); i.iter.Next() {
		key, err := i.Key()
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// KeyValue returns the current key and value decoded.
func (i Iterator[K, V]) KeyValue() (kv KeyValue[K, V], err error) {
	key, err := i.Key()
	if err != nil {
		return kv, err
	}
	value, err := i.Value()
	if err != nil {
		return kv, err
	}
	kv.Key = key
	kv.Value = value
	return kv, nil
}

// KeyValues fully consumes the iterator and returns the list of key and values within the iterator range.
func (i Iterator[K, V]) KeyValues() ([]KeyValue[K, V], error) {
	defer i.Close()

	var kvs []KeyValue[K, V]
	for ; i.iter.Valid(); i.iter.Next() {
		kv, err := i.KeyValue()
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, kv)
	}

	return kvs, nil
}

func (i Iterator[K, V]) Close() error { return i.iter.Close() }
func (i Iterator[K, V]) Next()        { i.iter.Next() }
func (i Iterator[K, V]) Valid() bool  { return i.iter.Valid() }

type KeyValue[K, V any] struct {
	Key   K
	Value V
}

func extendOneByte(b []byte) []byte {
	return append(b, 0)
}
