package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		init   []Model
		toGet  []Model
		toList []listQuery

		setCache    []Model
		removeCache []Model
		getCache    []Model
		listCache   []listQuery
	}{
		// simple add
		{
			init:  []Model{m("a", "1"), m("c", "2")},
			toGet: []Model{m("a", "1"), m("c", "2"), m("d", "")},
			toList: []listQuery{{
				"a", "e", 0,
				[]Model{m("a", "1"), m("c", "2")},
				m("c", "2"),
			}},
			setCache:    []Model{m("d", "3")},
			removeCache: []Model{m("a", "1")},
			getCache:    []Model{m("a", ""), m("c", "2"), m("d", "3")},
			listCache: []listQuery{{
				"a", "e", 0,
				[]Model{m("c", "2"), m("d", "3")},
				m("d", "3"),
			}},
		},
	}

	checkGet := func(db SimpleDB, m Model, msg string) {
		val := db.Get(m.Key)
		assert.EqualValues(m.Value, val, msg)
		has := db.Has(m.Key)
		assert.Equal(len(m.Value) != 0, has, msg)
	}

	checkList := func(db SimpleDB, lq listQuery, msg string) {
		start, end := []byte(lq.start), []byte(lq.end)
		list := db.List(start, end, lq.limit)
		if assert.EqualValues(lq.expected, list, msg) {
			var first Model
			if len(lq.expected) > 0 {
				first = lq.expected[0]
			}
			f := db.First(start, end)
			assert.EqualValues(first, f, msg)
			l := db.Last(start, end)
			assert.EqualValues(lq.last, l, msg)
		}
	}

	for i, tc := range cases {
		for j, db := range GetDBs() {
			for _, s := range tc.init {
				db.Set(s.Key, s.Value)
			}
			for k, g := range tc.toGet {
				msg := fmt.Sprintf("%d/%d/%d: %#v", i, j, k, g)
				checkGet(db, g, msg)
			}
			for k, lq := range tc.toList {
				msg := fmt.Sprintf("%d/%d/%d", i, j, k)
				checkList(db, lq, msg)
			}

			// make cache
			cache := db.Checkpoint()

			for _, s := range tc.setCache {
				cache.Set(s.Key, s.Value)
			}
			for k, r := range tc.removeCache {
				val := cache.Remove(r.Key)
				assert.EqualValues(r.Value, val, "%d/%d/%d: %#v", i, j, k, r)
			}

			// make sure data is in cache
			for k, g := range tc.getCache {
				msg := fmt.Sprintf("%d/%d/%d: %#v", i, j, k, g)
				checkGet(cache, g, msg)
			}
			for k, lq := range tc.listCache {
				msg := fmt.Sprintf("%d/%d/%d", i, j, k)
				checkList(cache, lq, msg)
			}

			// data not in basic store
			for k, g := range tc.toGet {
				msg := fmt.Sprintf("%d/%d/%d: %#v", i, j, k, g)
				checkGet(db, g, msg)
			}
			for k, lq := range tc.toList {
				msg := fmt.Sprintf("%d/%d/%d", i, j, k)
				checkList(db, lq, msg)
			}

			// commit
			db.Commit(cache)

			// make sure data is in cache
			for k, g := range tc.getCache {
				msg := fmt.Sprintf("%d/%d/%d", i, j, k)
				checkGet(db, g, msg)
			}
			for k, lq := range tc.listCache {
				msg := fmt.Sprintf("%d/%d/%d", i, j, k)
				checkList(db, lq, msg)
			}
		}
	}
}
