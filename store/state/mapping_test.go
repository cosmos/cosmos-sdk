package state

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type mapping interface {
	Get(Context, interface{}, interface{})
	GetSafe(Context, interface{}, interface{}) error
	Set(Context, interface{}, interface{})
	Has(Context, interface{}) bool
	Delete(Context, interface{})
	RandomKey() interface{}
}

type mappingT struct {
	Mapping
}

var _ mapping = mappingT{}

func newMapping() mappingT {
	return mappingT{NewMapping(testkey, testcdc, nil)}
}

func (m mappingT) Get(ctx Context, key interface{}, ptr interface{}) {
	m.Mapping.Get(ctx, []byte(key.(string)), ptr)
}

func (m mappingT) GetSafe(ctx Context, key interface{}, ptr interface{}) error {
	return m.Mapping.GetSafe(ctx, []byte(key.(string)), ptr)
}

func (m mappingT) Set(ctx Context, key interface{}, o interface{}) {
	m.Mapping.Set(ctx, []byte(key.(string)), o)
}

func (m mappingT) Has(ctx Context, key interface{}) bool {
	return m.Mapping.Has(ctx, []byte(key.(string)))
}

func (m mappingT) Delete(ctx Context, key interface{}) {
	m.Mapping.Delete(ctx, []byte(key.(string)))
}

func (m mappingT) RandomKey() interface{} {
	bz := make([]byte, 64)
	rand.Read(bz)
	return string(bz)
}

type indexerT struct {
	Indexer
}

var _ mapping = indexerT{}

func newIndexer(enc IntEncoding) indexerT {
	return indexerT{NewIndexer(NewMapping(testkey, testcdc, nil), enc)}
}

func (m indexerT) Get(ctx Context, key interface{}, ptr interface{}) {
	m.Indexer.Get(ctx, key.(uint64), ptr)
}

func (m indexerT) GetSafe(ctx Context, key interface{}, ptr interface{}) error {
	return m.Indexer.GetSafe(ctx, key.(uint64), ptr)
}

func (m indexerT) Set(ctx Context, key interface{}, o interface{}) {
	m.Indexer.Set(ctx, key.(uint64), o)
}

func (m indexerT) Has(ctx Context, key interface{}) bool {
	return m.Indexer.Has(ctx, key.(uint64))
}

func (m indexerT) Delete(ctx Context, key interface{}) {
	m.Indexer.Delete(ctx, key.(uint64))
}

func (m indexerT) RandomKey() interface{} {
	return rand.Uint64()
}

func TestMapping(t *testing.T) {
	ctx := defaultComponents()
	table := []mapping{newMapping(), newIndexer(Dec), newIndexer(Hex), newIndexer(Bin)}

	for _, m := range table {
		exp := make(map[interface{}]uint64)
		for n := 0; n < 10e4; n++ {
			k, v := m.RandomKey(), rand.Uint64()
			require.False(t, m.Has(ctx, k))
			exp[k] = v
			m.Set(ctx, k, v)
		}

		for k, v := range exp {
			ptr := new(uint64)
			m.Get(ctx, k, ptr)
			require.Equal(t, v, indirect(ptr))

			ptr = new(uint64)
			err := m.GetSafe(ctx, k, ptr)
			require.NoError(t, err)
			require.Equal(t, v, indirect(ptr))

			require.True(t, m.Has(ctx, k))

			m.Delete(ctx, k)
			require.False(t, m.Has(ctx, k))
			ptr = new(uint64)
			err = m.GetSafe(ctx, k, ptr)
			require.Error(t, err)
			require.Equal(t, reflect.Zero(reflect.TypeOf(ptr).Elem()).Interface(), indirect(ptr))
		}
	}
}
