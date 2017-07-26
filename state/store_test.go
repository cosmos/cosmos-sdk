package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetDBs() []SimpleDB {
	return []SimpleDB{
		NewMemKVStore(),
	}
}

// TestKVStore makes sure that get/set/remove operations work,
// as well as list
func TestKVStore(t *testing.T) {
	assert := assert.New(t)

	type listQuery struct {
		// this is the list query
		start, end []byte
		limit      int
		// expected result from List, first element also expected for First
		expected []Model
		// expected result from Last
		last Model
	}

	cases := []struct {
		toSet    []Model
		toRemove []Model
		toGet    []Model
		toList   []listQuery
	}{
		// simple add
		{
			[]Model{
				{[]byte{1}, []byte{2}},
				{[]byte{3}, []byte{4}},
			},
			nil,
			[]Model{{[]byte{1}, []byte{2}}},
			[]listQuery{
				{
					[]byte{1}, []byte{4}, 0,
					// all
					[]Model{
						{[]byte{1}, []byte{2}},
						{[]byte{3}, []byte{4}},
					},
					// last one
					Model{[]byte{3}, []byte{4}},
				},
				{
					[]byte{1}, []byte{3}, 10,
					// all
					[]Model{
						{[]byte{1}, []byte{2}},
					},
					// last one
					Model{[]byte{1}, []byte{2}},
				},
			},
		},
		// over-write data, remove
		{
			[]Model{
				{[]byte{1}, []byte{2}},
				{[]byte{2}, []byte{2}},
				{[]byte{3}, []byte{2}},
				{[]byte{2}, []byte{4}},
			},
			[]Model{{[]byte{3}, []byte{2}}},
			[]Model{
				{[]byte{1}, []byte{2}},
				{[]byte{2}, []byte{4}},
				{[]byte{3}, nil},
			},
			[]listQuery{
				{
					[]byte{0, 5}, []byte{10}, 1,
					// all
					[]Model{
						{[]byte{1}, []byte{2}},
					},
					// last
					Model{[]byte{2}, []byte{4}},
				},
				{
					[]byte{1, 4}, []byte{1, 7}, 10,
					[]Model{},
					Model{},
				},
				{
					[]byte{1, 5}, []byte{10}, 0,
					[]Model{
						{[]byte{2}, []byte{4}},
					},
					Model{[]byte{2}, []byte{4}},
				},
			},
		},
	}

	for i, tc := range cases {
		for j, db := range GetDBs() {
			for _, s := range tc.toSet {
				db.Set(s.Key, s.Value)
			}
			for k, r := range tc.toRemove {
				val := db.Remove(r.Key)
				assert.EqualValues(r.Value, val, "%d/%d/%d", i, j, k)
			}
			for k, g := range tc.toGet {
				val := db.Get(g.Key)
				assert.EqualValues(g.Value, val, "%d/%d/%d", i, j, k)
				has := db.Has(g.Key)
				assert.Equal(len(g.Value) != 0, has, "%d/%d/%d", i, j, k)
			}
			for k, lq := range tc.toList {
				list := db.List(lq.start, lq.end, lq.limit)
				if assert.EqualValues(lq.expected, list, "%d/%d/%d", i, j, k) {
					var first Model
					if len(lq.expected) > 0 {
						first = lq.expected[0]
					}
					f := db.First(lq.start, lq.end)
					assert.EqualValues(first, f, "%d/%d/%d", i, j, k)
					l := db.Last(lq.start, lq.end)
					assert.EqualValues(lq.last, l, "%d/%d/%d", i, j, k)
				}
			}
		}
	}
}
