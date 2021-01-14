package smt_test

import (
	"bytes"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/store/smt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
)

func TestIteration(t *testing.T) {
	pairs := []struct{ key, val []byte }{
		{[]byte("foo"), []byte("bar")},
		{[]byte("lorem"), []byte("ipsum")},
		{[]byte("alpha"), []byte("beta")},
		{[]byte("gamma"), []byte("delta")},
		{[]byte("epsilon"), []byte("zeta")},
		{[]byte("eta"), []byte("theta")},
		{[]byte("iota"), []byte("kappa")},
	}

	s := smt.NewStore(dbm.NewMemDB())

	for _, p := range pairs {
		s.Set(p.key, p.val)
	}

	// sort test data by key, to get "expected" ordering
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].key, pairs[j].key) < 0
	})

	iter := s.Iterator([]byte("alpha"), []byte("omega"))
	for _, p := range pairs {
		require.True(t, iter.Valid())
		assert.Equal(t, p.key, iter.Key())
		assert.Equal(t, p.val, iter.Value())
		iter.Next()
	}
	assert.False(t, iter.Valid())
	assert.NoError(t, iter.Error())
	assert.NoError(t, iter.Close())

	iter = s.Iterator(nil, nil)
	for _, p := range pairs {
		require.True(t, iter.Valid())
		assert.Equal(t, p.key, iter.Key())
		assert.Equal(t, p.val, iter.Value())
		iter.Next()
	}
	assert.False(t, iter.Valid())
	assert.NoError(t, iter.Error())
	assert.NoError(t, iter.Close())

	iter = s.Iterator([]byte("epsilon"), []byte("gamma"))
	for _, p := range pairs[1:4] {
		require.True(t, iter.Valid())
		assert.Equal(t, p.key, iter.Key())
		assert.Equal(t, p.val, iter.Value())
		iter.Next()
	}
	assert.False(t, iter.Valid())
	assert.NoError(t, iter.Error())
	assert.NoError(t, iter.Close())

	rIter := s.ReverseIterator(nil, nil)
	for i := len(pairs) - 1; i >= 0; i-- {
		require.True(t, rIter.Valid())
		assert.Equal(t, pairs[i].key, rIter.Key())
		assert.Equal(t, pairs[i].val, rIter.Value())
		rIter.Next()
	}
	assert.False(t, rIter.Valid())
	assert.NoError(t, rIter.Error())
	assert.NoError(t, rIter.Close())
}

func TestDomain(t *testing.T) {
	s := smt.NewStore(dbm.NewMemDB())

	iter := s.Iterator(nil, nil)
	start, end := iter.Domain()
	assert.Nil(t, start)
	assert.Nil(t, end)

	iter = s.Iterator([]byte("foo"), []byte("bar"))
	start, end = iter.Domain()
	assert.Equal(t, []byte("foo"), start)
	assert.Equal(t, []byte("bar"), end)
}
