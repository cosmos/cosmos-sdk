package state

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/merkleeyes/iavl"
	// dbm "github.com/tendermint/tmlibs/db"
)

func GetDBs() []SimpleDB {
	// // tree with persistence....
	// tmpDir, err := ioutil.TempDir("", "state-tests")
	// if err != nil {
	// 	panic(err)
	// }
	// db := dbm.NewDB("test-get-dbs", dbm.LevelDBBackendStr, tmpDir)
	// persist := iavl.NewIAVLTree(500, db)

	return []SimpleDB{
		NewMemKVStore(),
		NewBonsai(iavl.NewIAVLTree(0, nil)),
		// NewBonsai(persist),
	}
}

func b(k string) []byte {
	if k == "" {
		return nil
	}
	return []byte(k)
}

func m(k, v string) Model {
	return Model{
		Key:   b(k),
		Value: b(v),
	}
}

type listQuery struct {
	// this is the list query
	start, end string
	limit      int
	// expected result from List, first element also expected for First
	expected []Model
	// expected result from Last
	last Model
}

// TestKVStore makes sure that get/set/remove operations work,
// as well as list
func TestKVStore(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		toSet    []Model
		toRemove []Model
		toGet    []Model
		toList   []listQuery
	}{
		// simple add
		{
			toSet:    []Model{m("a", "b"), m("c", "d")},
			toRemove: nil,
			toGet:    []Model{m("a", "b")},
			toList: []listQuery{
				{
					"a", "d", 0,
					[]Model{m("a", "b"), m("c", "d")},
					m("c", "d"),
				},
				{
					"a", "c", 10,
					[]Model{m("a", "b")},
					m("a", "b"),
				},
			},
		},
		// over-write data, remove
		{
			toSet: []Model{
				m("a", "1"),
				m("b", "2"),
				m("c", "3"),
				m("b", "4"),
			},
			toRemove: []Model{m("c", "3")},
			toGet: []Model{
				m("a", "1"),
				m("b", "4"),
				m("c", ""),
			},
			toList: []listQuery{
				{
					"0d", "h", 1,
					[]Model{m("a", "1")},
					m("b", "4"),
				},
				{
					"ad", "ak", 10,
					[]Model{},
					Model{},
				},
				{
					"ad", "k", 0,
					[]Model{m("b", "4")},
					m("b", "4"),
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
				start, end := []byte(lq.start), []byte(lq.end)
				list := db.List(start, end, lq.limit)
				if assert.EqualValues(lq.expected, list, "%d/%d/%d", i, j, k) {
					var first Model
					if len(lq.expected) > 0 {
						first = lq.expected[0]
					}
					f := db.First(start, end)
					assert.EqualValues(first, f, "%d/%d/%d", i, j, k)
					l := db.Last(start, end)
					assert.EqualValues(lq.last, l, "%d/%d/%d", i, j, k)
				}
			}
		}
	}
}
