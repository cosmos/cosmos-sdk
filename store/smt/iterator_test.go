package smt_test

import (
	"bytes"
	"sort"
	"testing"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/store/smt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIteration(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

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
		require.True(iter.Valid())
		require.Equal(p.key, iter.Key())
		require.Equal(p.val, iter.Value())
		iter.Next()
	}
	assert.False(iter.Valid())
	assert.NoError(iter.Error())
	assert.NoError(iter.Close())

	iter = s.Iterator(nil, nil)
	for _, p := range pairs {
		require.True(iter.Valid())
		require.Equal(p.key, iter.Key())
		require.Equal(p.val, iter.Value())
		iter.Next()
	}
	assert.False(iter.Valid())
	assert.NoError(iter.Error())
	assert.NoError(iter.Close())

	iter = s.Iterator([]byte("epsilon"), []byte("gamma"))
	for _, p := range pairs[1:4] {
		require.True(iter.Valid())
		require.Equal(p.key, iter.Key())
		require.Equal(p.val, iter.Value())
		iter.Next()
	}
	assert.False(iter.Valid())
	assert.NoError(iter.Error())
	assert.NoError(iter.Close())

	rIter := s.ReverseIterator(nil, nil)
	for i := len(pairs) - 1; i >= 0; i-- {
		require.True(rIter.Valid())
		require.Equal(pairs[i].key, rIter.Key())
		require.Equal(pairs[i].val, rIter.Value())
		rIter.Next()
	}
	assert.False(rIter.Valid())
	assert.NoError(rIter.Error())
	assert.NoError(rIter.Close())

	// delete something, and ensure that iteration still works
	s.Delete([]byte("eta"))

	iter = s.Iterator(nil, nil)
	for _, p := range pairs {
		if !bytes.Equal([]byte("eta"), p.key) {
			require.True(iter.Valid())
			require.Equal(p.key, iter.Key())
			require.Equal(p.val, iter.Value())
			iter.Next()
		}
	}
	assert.False(iter.Valid())
	assert.NoError(iter.Error())
	assert.NoError(iter.Close())
}

func TestDomain(t *testing.T) {
	assert := assert.New(t)
	s := smt.NewStore(dbm.NewMemDB())

	iter := s.Iterator(nil, nil)
	start, end := iter.Domain()
	assert.Nil(start)
	assert.Nil(end)

	iter = s.Iterator([]byte("foo"), []byte("bar"))
	start, end = iter.Domain()
	assert.Equal([]byte("foo"), start)
	assert.Equal([]byte("bar"), end)
}
