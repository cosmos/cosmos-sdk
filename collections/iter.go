package collections

import (
	"context"
	"cosmossdk.io/core/store"
	"errors"
	"fmt"
)

// ErrInvalidIterator is returned when an Iterate call resulted in an invalid iterator.
var ErrInvalidIterator = errors.New("collections: invalid iterator")

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
	RangeValues() (prefix *K, start *Bound[K], end *Bound[K], order Order, err error)
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

// test sentinel error
var errRange = errors.New("collections: range error")
var errOrder = errors.New("collections: invalid order")

func (r *Range[K]) RangeValues() (prefix *K, start *Bound[K], end *Bound[K], order Order, err error) {
	if r.prefix != nil && (r.end != nil || r.start != nil) {
		return nil, nil, nil, order, fmt.Errorf("%w: prefix must not be set if either start or end are specified", errRange)
	}
	return r.prefix, r.start, r.end, r.order, nil
}

// iteratorFromRanger generates an Iterator instance, with the proper prefixing and ranging.
// a nil Ranger can be seen as an ascending iteration over all the possible keys.
func iteratorFromRanger[K, V any](ctx context.Context, m Map[K, V], r Ranger[K]) (iter Iterator[K, V], err error) {
	var (
		prefix *K
		start  *Bound[K]
		end    *Bound[K]
		order  = OrderAscending
	)
	// if Ranger is specified then we override the defaults
	if r != nil {
		prefix, start, end, order, err = r.RangeValues()
		if err != nil {
			return iter, err
		}
	}
	if prefix != nil && (start != nil || end != nil) {
		return iter, fmt.Errorf("%w: prefix must not be set if either start or end are specified", errRange)
	}

	// compute start and end bytes
	var startBytes, endBytes []byte
	if prefix != nil {
		startBytes, endBytes, err = prefixStartEndBytes(m, *prefix)
		if err != nil {
			return iter, err
		}
	} else {
		startBytes, endBytes, err = rangeStartEndBytes(m, start, end)
		if err != nil {
			return iter, err
		}
	}

	// get store
	kv := m.sk(ctx)

	// create iter
	var storeIter store.Iterator
	switch order {
	case OrderAscending:
		storeIter = kv.Iterator(startBytes, endBytes)
	case OrderDescending:
		storeIter = kv.ReverseIterator(startBytes, endBytes)
	default:
		return iter, fmt.Errorf("%w: %d", errOrder, order)
	}

	// check if valid
	if !storeIter.Valid() {
		return iter, ErrInvalidIterator
	}

	// all good
	iter.kc = m.kc
	iter.vc = m.vc
	iter.prefixLength = len(m.prefix)
	iter.iter = storeIter
	return iter, nil
}

// rangeStartEndBytes computes a range's start and end bytes to be passed to the store's iterator.
func rangeStartEndBytes[K, V any](m Map[K, V], start, end *Bound[K]) (startBytes, endBytes []byte, err error) {
	startBytes = m.prefix
	if start != nil {
		startBytes, err = encodeKeyWithPrefix(m.prefix, m.kc, start.value)
		if err != nil {
			return startBytes, endBytes, err
		}
		// the start of iterators is by default inclusive,
		// in order to make it exclusive we extend the start
		// by one single byte.
		if !start.inclusive {
			startBytes = extendOneByte(startBytes)
		}
	}
	if end != nil {
		endBytes, err = encodeKeyWithPrefix(m.prefix, m.kc, end.value)
		if err != nil {
			return startBytes, endBytes, err
		}
		// the end of iterators is by default exclusive
		// in order to make it inclusive we extend the end
		// by one single byte.
		if end.inclusive {
			endBytes = extendOneByte(endBytes)
		}
	} else {
		// if end is not specified then we simply are
		// inclusive up to the last key of the Prefix
		// of the collection.
		endBytes = prefixEndBytes(m.prefix)
	}

	return startBytes, endBytes, nil
}

// prefixStartEndBytes returns the start and end bytes to be provided to the store's iterator, considering we're prefixing
// over a specific key.
func prefixStartEndBytes[K, V any](m Map[K, V], prefix K) (startBytes, endBytes []byte, err error) {
	startBytes, err = encodeKeyWithPrefix(m.prefix, m.kc, prefix)
	if err != nil {
		return
	}
	return startBytes, prefixEndBytes(startBytes), nil
}

// Iterator defines a generic wrapper around an sdk.Iterator.
// This iterator provides automatic key and value encoding,
// it assumes all the keys and values contained within the sdk.Iterator
// range are the same.
type Iterator[K, V any] struct {
	kc KeyCodec[K]
	vc ValueCodec[V]

	iter store.Iterator

	prefixLength int // prefixLength refers to the bytes provided by Prefix.Bytes, not Ranger.RangeValues() prefix.
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

// KeyValue represent a Key and Value pair of an iteration.
type KeyValue[K, V any] struct {
	Key   K
	Value V
}

func extendOneByte(b []byte) []byte {
	return append(b, 0)
}

// prefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		}

		end = end[:len(end)-1]

		if len(end) == 0 {
			end = nil
			break
		}
	}

	return end
}
