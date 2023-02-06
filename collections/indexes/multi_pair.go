package indexes

import (
	"context"
	"cosmossdk.io/collections"
)

// MultiPair is an index that is used with collections.Pair keys. It indexes objects by their second part of the key.
// When the value is being indexed by collections.IndexedMap then MultiPair will create a relationship between
// the second part of the primary key and the first part.
type MultiPair[K1, K2, Value any] collections.GenericMultiIndex[K2, K1, collections.Pair[K1, K2], Value]

// TODO(tip): this is an interface to cast a collections.KeyCodec
// to a pair codec. currently we return it as a KeyCodec[Pair[K1, K2]]
// to improve dev experience with type inference, which means we cannot
// get the concrete implementation which exposes KeyCodec1 and KeyCodec2.
type pairKeyCodec[K1, K2 any] interface {
	KeyCodec1() collections.KeyCodec[K1]
	KeyCodec2() collections.KeyCodec[K2]
}

// NewMultiPair instantiates a new MultiPair index.
// NOTE: when using this function you will need to type hint: doing NewMultiPair[Value]()
// Example: if the value of the indexed map is string, you need to do NewMultiPair[string](...)
func NewMultiPair[Value any, K1, K2 any](
	sb *collections.SchemaBuilder,
	prefix collections.Prefix,
	name string,
	pairCodec collections.KeyCodec[collections.Pair[K1, K2]],
) *MultiPair[K1, K2, Value] {
	pkc := pairCodec.(pairKeyCodec[K1, K2])
	mi := collections.NewGenericMultiIndex(
		sb,
		prefix,
		name,
		pkc.KeyCodec2(),
		pkc.KeyCodec1(),
		func(pk collections.Pair[K1, K2], _ Value) ([]collections.IndexReference[K2, K1], error) {
			return []collections.IndexReference[K2, K1]{
				collections.NewIndexReference(pk.K2(), pk.K1()),
			}, nil
		},
	)

	return (*MultiPair[K1, K2, Value])(mi)
}

// Iterate exposes the raw iterator API.
func (i *MultiPair[K1, K2, Value]) Iterate(ctx context.Context, ranger collections.Ranger[collections.Pair[K2, K1]]) (iter MultiPairIterator[K2, K1], err error) {
	sIter, err := (*collections.GenericMultiIndex[K2, K1, collections.Pair[K1, K2], Value])(i).Iterate(ctx, ranger)
	if err != nil {
		return iter, err
	}
	return (MultiPairIterator[K2, K1])(sIter), nil
}

// MatchExact will return an iterator containing only the primary keys starting with the provided second part of the multipart pair key.
func (i *MultiPair[K1, K2, Value]) MatchExact(ctx context.Context, key K2) (MultiPairIterator[K2, K1], error) {
	return i.Iterate(ctx, collections.NewPrefixedPairRange[K2, K1](key))
}

// Reference implements collections.Index
func (i *MultiPair[K1, K2, Value]) Reference(ctx context.Context, pk collections.Pair[K1, K2], value Value, oldValue *Value) error {
	return (*collections.GenericMultiIndex[K2, K1, collections.Pair[K1, K2], Value])(i).Reference(ctx, pk, value, oldValue)
}

// Unreference implements collections.Index
func (i *MultiPair[K1, K2, Value]) Unreference(ctx context.Context, pk collections.Pair[K1, K2], value Value) error {
	return (*collections.GenericMultiIndex[K2, K1, collections.Pair[K1, K2], Value])(i).Unreference(ctx, pk, value)
}

// MultiPairIterator is a helper type around a collections.KeySetIterator when used to work
// with MultiPair indexes iterations.
type MultiPairIterator[K2, K1 any] collections.KeySetIterator[collections.Pair[K2, K1]]

// PrimaryKey returns the primary key from the index. The index is composed like a reverse
// pair key. So we just fetch the pair key from the index and return the reverse.
func (m MultiPairIterator[K2, K1]) PrimaryKey() (pair collections.Pair[K1, K2], err error) {
	reversePair, err := (collections.KeySetIterator[collections.Pair[K2, K1]])(m).Key()
	if err != nil {
		return pair, err
	}
	pair = collections.Join(reversePair.K2(), reversePair.K1())
	return pair, nil
}

// PrimaryKeys returns all the primary keys contained in the iterator.
func (m MultiPairIterator[K2, K1]) PrimaryKeys() (pairs []collections.Pair[K1, K2], err error) {
	defer m.Close()
	for ; m.Valid(); m.Next() {
		pair, err := m.PrimaryKey()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)
	}
	return pairs, err
}

func (m MultiPairIterator[K2, K1]) Next() {
	(collections.KeySetIterator[collections.Pair[K2, K1]])(m).Next()
}

func (m MultiPairIterator[K2, K1]) Valid() bool {
	return (collections.KeySetIterator[collections.Pair[K2, K1]])(m).Valid()
}

func (m MultiPairIterator[K2, K1]) Close() error {
	return (collections.KeySetIterator[collections.Pair[K2, K1]])(m).Close()
}
