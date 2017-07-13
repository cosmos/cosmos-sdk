package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	assert := assert.New(t)

	lots := make([][]byte, 500)
	for i := range lots {
		lots[i] = []byte{1, 8, 7}
	}

	cases := []struct {
		pushes [][]byte
		pops   [][]byte
	}{
		// fill it up and empty it all
		{
			[][]byte{{1, 2, 3}, {44}, {3, 0}},
			[][]byte{{1, 2, 3}, {44}, {3, 0}},
		},
		// don't empty everything - size is 1 at the end
		{
			[][]byte{{77, 22}, {11, 9}, {121}},
			[][]byte{{77, 22}, {11, 9}},
		},
		// empty too much, just get nil, no negative size
		{
			[][]byte{{1}, {2}, {4}},
			[][]byte{{1}, {2}, {4}, nil, nil, nil},
		},
		// let's play with lots....
		{lots, append(lots, nil)},
	}

	for i, tc := range cases {
		store := NewMemKVStore()

		// initialize a queue and add items
		q := NewQueue(store)
		for j, in := range tc.pushes {
			cnt := q.Push(in)
			assert.Equal(uint64(j), cnt, "%d", i)
		}
		assert.EqualValues(len(tc.pushes), q.Size())

		// load from disk and pop them
		r := NewQueue(store)
		for _, out := range tc.pops {
			val := r.Pop()
			assert.Equal(out, val, "%d", i)
		}

		// it's empty in memory and on disk
		expected := len(tc.pushes) - len(tc.pops)
		if expected < 0 {
			expected = 0
		}
		assert.EqualValues(expected, r.Size())
		s := NewQueue(store)
		assert.EqualValues(expected, s.Size())
	}
}
