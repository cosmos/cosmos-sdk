package collections

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/core/store"
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

type rangeBoundKind uint8

const (
	rangeBoundKindNone rangeBoundKind = iota
	rangeBoundKindNextKey
	rangeBoundKindNextPrefixKey
)

type RangeBound[K any] struct {
	kind rangeBoundKind
	key  K
}

func RangeBoundNextKey[K any](key K) *RangeBound[K] {
	return &RangeBound[K]{key: key, kind: rangeBoundKindNextKey}
}

func RangeBoundNextPrefixKey[K any](key K) *RangeBound[K] {
	return &RangeBound[K]{key: key, kind: rangeBoundKindNextPrefixKey}
}

func RangeBoundNone[K any](key K) *RangeBound[K] {
	return &RangeBound[K]{key: key, kind: rangeBoundKindNone}
}

// Ranger defines a generic interface that provides a range of keys.
type Ranger[K any] interface {
	// RangeValues is defined by Ranger implementers.
	// TODO doc
	RangeValues() (start *RangeBound[K], end *RangeBound[K], order Order, err error)
}

// Range is a Ranger implementer.
type Range[K any] struct {
	start *RangeBound[K]
	end   *RangeBound[K]
	order Order
}

// Prefix sets a fixed prefix for the key range.
func (r *Range[K]) Prefix(key K) *Range[K] {
	r.start = RangeBoundNone(key)
	r.end = RangeBoundNextPrefixKey(key)
	return r
}

// StartInclusive makes the range contain only keys which are bigger or equal to the provided start K.
func (r *Range[K]) StartInclusive(start K) *Range[K] {
	r.start = RangeBoundNone(start)
	return r
}

// StartExclusive makes the range contain only keys which are bigger to the provided start K.
func (r *Range[K]) StartExclusive(start K) *Range[K] {
	r.start = RangeBoundNextKey(start)
	return r
}

// EndInclusive makes the range contain only keys which are smaller or equal to the provided end K.
func (r *Range[K]) EndInclusive(end K) *Range[K] {
	r.end = RangeBoundNextKey(end)
	return r
}

// EndExclusive makes the range contain only keys which are smaller to the provided end K.
func (r *Range[K]) EndExclusive(end K) *Range[K] {
	r.end = RangeBoundNone(end)
	return r
}

func (r *Range[K]) Descending() *Range[K] {
	r.order = OrderDescending
	return r
}

// test sentinel error
var errRange = errors.New("collections: range error")
var errOrder = errors.New("collections: invalid order")

func (r *Range[K]) RangeValues() (start *RangeBound[K], end *RangeBound[K], order Order, err error) {
	return r.start, r.end, r.order, nil
}

// iteratorFromRanger generates an Iterator instance, with the proper prefixing and ranging.
// a nil Ranger can be seen as an ascending iteration over all the possible keys.
func iteratorFromRanger[K, V any](ctx context.Context, m Map[K, V], r Ranger[K]) (iter Iterator[K, V], err error) {
	var (
		start *RangeBound[K]
		end   *RangeBound[K]
		order = OrderAscending
	)

	if r == nil {
		start, end, order, err = r.RangeValues()
		if err != nil {
			return iter, err
		}
	}

	startBytes := m.prefix
	if start != nil {
		startBytes, err = encodeRangeBound(m.prefix, m.kc, start)
		if err != nil {
			return iter, err
		}
	}
	var endBytes []byte
	if end != nil {
		endBytes, err = encodeRangeBound(m.prefix, m.kc, end)
		if err != nil {
			return iter, err
		}
	} else {
		endBytes = nextBytesPrefixKey(m.prefix)
	}

	kv := m.sa(ctx)
	switch order {
	case OrderAscending:
		return newIterator(kv.Iterator(startBytes, endBytes), m), nil
	case OrderDescending:
		return newIterator(kv.ReverseIterator(startBytes, endBytes), m), nil
	default:
		return iter, errOrder
	}
}

func newIterator[K, V any](iterator store.Iterator, m Map[K, V]) Iterator[K, V] {
	return Iterator[K, V]{
		kc:           m.kc,
		vc:           m.vc,
		iter:         iterator,
		prefixLength: len(m.prefix),
	}
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

// encodeRangeBound encodes a range bound, modifying the key bytes to adhere to bound semantics.
func encodeRangeBound[T any](prefix []byte, keyCodec KeyCodec[T], bound *RangeBound[T]) ([]byte, error) {
	key, err := encodeKeyWithPrefix(prefix, keyCodec, bound.key)
	if err != nil {
		return nil, err
	}
	switch bound.kind {
	case rangeBoundKindNone:
		return key, nil
	case rangeBoundKindNextKey:
		return nextBytesKey(key), nil
	case rangeBoundKindNextPrefixKey:
		return nextBytesPrefixKey(key), nil
	default:
		panic("undefined bound kind")
	}
}

// nextBytesKey returns the next byte key after this one.
func nextBytesKey(b []byte) []byte {
	return append(b, 0)
}

// nextBytesPrefixKey returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func nextBytesPrefixKey(prefix []byte) []byte {
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
