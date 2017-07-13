package state

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type kv struct {
	k uint64
	v []byte
}

type bscase struct {
	data []kv
	// these are the tests to try out
	top    kv
	bottom kv
	gets   []kv // for each item check the query matches
	lte    []kv // value for lte queires...
	gte    []kv // value for gte
}

func TestBasicSpan(t *testing.T) {

	a, b, c := []byte{0xaa}, []byte{0xbb}, []byte{0xcc}

	lots := make([]kv, 1000)
	for i := range lots {
		lots[i] = kv{uint64(3 * i), []byte{byte(i / 100), byte(i % 100)}}
	}

	cases := []bscase{
		// simplest queries
		{
			[]kv{{1, a}, {3, b}, {5, c}},
			kv{5, c},
			kv{1, a},
			[]kv{{1, a}, {3, b}, {5, c}},
			[]kv{{2, a}, {77, c}, {3, b}, {0, nil}}, // lte
			[]kv{{6, nil}, {2, b}, {1, a}},          // gte
		},
		// add out of order
		{
			[]kv{{7, a}, {2, b}, {6, c}},
			kv{7, a},
			kv{2, b},
			[]kv{{2, b}, {6, c}, {7, a}},
			[]kv{{4, b}, {7, a}, {1, nil}}, // lte
			[]kv{{4, c}, {7, a}, {1, b}},   // gte
		},
		// add out of order and with duplicates
		{
			[]kv{{7, a}, {2, b}, {6, c}, {7, c}, {6, b}, {2, a}},
			kv{7, c},
			kv{2, a},
			[]kv{{2, a}, {6, b}, {7, c}},
			[]kv{{5, a}, {6, b}, {123, c}},         // lte
			[]kv{{0, a}, {3, b}, {7, c}, {8, nil}}, // gte
		},
		// try lots...
		{
			lots,
			lots[len(lots)-1],
			lots[0],
			lots,
			nil,
			nil,
		},
	}

	for i, tc := range cases {
		store := NewMemKVStore()

		// initialize a queue and add items
		s := NewSpan(store)
		for _, x := range tc.data {
			s.Set(x.k, x.v)
		}

		testSpan(t, i, s, tc)
		// reload and try the queries again
		s2 := NewSpan(store)
		testSpan(t, i+10, s2, tc)
	}
}

func testSpan(t *testing.T, idx int, s *Span, tc bscase) {
	assert := assert.New(t)
	i := strconv.Itoa(idx)

	v, k := s.Top()
	assert.Equal(tc.top.k, k, i)
	assert.Equal(tc.top.v, v, i)

	v, k = s.Bottom()
	assert.Equal(tc.bottom.k, k, i)
	assert.Equal(tc.bottom.v, v, i)

	for _, g := range tc.gets {
		v = s.Get(g.k)
		assert.Equal(g.v, v, i)
	}

	for _, l := range tc.lte {
		v, k = s.LTE(l.k)
		assert.Equal(l.v, v, i)
		if l.v != nil {
			assert.True(k <= l.k, i)
		}
	}

	for _, t := range tc.gte {
		v, k = s.GTE(t.k)
		assert.Equal(t.v, v, i)
		if t.v != nil {
			assert.True(k >= t.k, i)
		}
	}

}
