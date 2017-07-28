package state

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type pair struct {
	k []byte
	v []byte
}

type setCase struct {
	data []pair
	// these are the tests to try out
	gets []pair  // for each item check the query matches
	list KeyList // make sure the set returns the proper list
}

func TestSet(t *testing.T) {

	a, b, c, d := []byte{0xaa}, []byte{0xbb}, []byte{0xcc}, []byte{0xdd}

	cases := []setCase{

		// simplest queries
		{
			[]pair{{a, a}, {b, b}, {c, c}},
			[]pair{{c, c}, {d, nil}, {b, b}},
			KeyList{a, b, c},
		},
		// out of order
		{
			[]pair{{c, a}, {a, b}, {d, c}, {b, d}},
			[]pair{{a, b}, {b, d}},
			KeyList{a, b, c, d},
		},
		// duplicate and removing
		{
			[]pair{{c, a}, {c, c}, {a, d}, {d, d}, {b, b}, {d, nil}, {a, nil}, {a, a}, {b, nil}},
			[]pair{{a, a}, {c, c}, {b, nil}},
			KeyList{a, c},
		},
	}

	for i, tc := range cases {
		store := NewMemKVStore()

		// initialize a queue and add items
		s := NewSet(store)
		for _, x := range tc.data {
			s.Set(x.k, x.v)
		}

		testSet(t, i, s, tc)
		// reload and try the queries again
		s2 := NewSet(store)
		testSet(t, i+10, s2, tc)
	}
}

func testSet(t *testing.T, idx int, s *Set, tc setCase) {
	assert := assert.New(t)
	i := strconv.Itoa(idx)

	for _, g := range tc.gets {
		v := s.Get(g.k)
		assert.Equal(g.v, v, i)
		e := s.Exists(g.k)
		assert.Equal(e, (g.v != nil), i)
	}

	l := s.List()
	assert.True(tc.list.Equals(l), "%s: %v / %v", i, tc.list, l)
}
