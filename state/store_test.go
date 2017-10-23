package state

import (
	"io/ioutil"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/stretchr/testify/assert"

	"github.com/tendermint/iavl"
	dbm "github.com/tendermint/tmlibs/db"
)

func GetDBs() []sdk.SimpleDB {
	// tree with persistence....
	tmpDir, err := ioutil.TempDir("", "state-tests")
	if err != nil {
		panic(err)
	}
	db := dbm.NewDB("test-get-dbs", dbm.LevelDBBackendStr, tmpDir)
	persist := iavl.NewVersionedTree(500, db)

	return []sdk.SimpleDB{
		NewMemKVStore(),
		NewBonsai(iavl.NewVersionedTree(0, dbm.NewMemDB())),
		NewBonsai(persist),
	}
}

func b(k string) []byte {
	if k == "" {
		return nil
	}
	return []byte(k)
}

func m(k, v string) sdk.Model {
	return sdk.Model{
		Key:   b(k),
		Value: b(v),
	}
}

type listQuery struct {
	// this is the list query
	start, end string
	limit      int
	// expected result from List, first element also expected for First
	expected []sdk.Model
	// expected result from Last
	last sdk.Model
}

// TestKVStore makes sure that get/set/remove operations work,
// as well as list
func TestKVStore(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		toSet    []sdk.Model
		toRemove []sdk.Model
		toGet    []sdk.Model
		toList   []listQuery
	}{
		// simple add
		{
			toSet:    []sdk.Model{m("a", "b"), m("c", "d")},
			toRemove: nil,
			toGet:    []sdk.Model{m("a", "b")},
			toList: []listQuery{
				{
					"a", "d", 0,
					[]sdk.Model{m("a", "b"), m("c", "d")},
					m("c", "d"),
				},
				{
					"a", "c", 10,
					[]sdk.Model{m("a", "b")},
					m("a", "b"),
				},
			},
		},
		// over-write data, remove
		{
			toSet: []sdk.Model{
				m("a", "1"),
				m("b", "2"),
				m("c", "3"),
				m("b", "4"),
			},
			toRemove: []sdk.Model{m("c", "3")},
			toGet: []sdk.Model{
				m("a", "1"),
				m("b", "4"),
				m("c", ""),
			},
			toList: []listQuery{
				{
					"0d", "h", 1,
					[]sdk.Model{m("a", "1")},
					m("b", "4"),
				},
				{
					"ad", "ak", 10,
					[]sdk.Model{},
					sdk.Model{},
				},
				{
					"ad", "k", 0,
					[]sdk.Model{m("b", "4")},
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
					var first sdk.Model
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
