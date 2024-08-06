package storage

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/store/v2"
)

const (
	storeKey1 = "store1"
)

var storeKey1Bytes = []byte(storeKey1)

// StorageTestSuite defines a reusable test suite for all storage backends.
type StorageTestSuite struct {
	suite.Suite

	NewDB          func(dir string) (*StorageStore, error)
	EmptyBatchSize int
	SkipTests      []string
}

func (s *StorageTestSuite) TestDatabase_Close() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	s.Require().NoError(db.Close())

	// close should not be idempotent
	s.Require().Panics(func() { _ = db.Close() })
}

func (s *StorageTestSuite) TestDatabase_LatestVersion() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	lv, err := db.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Zero(lv)

	for i := uint64(1); i <= 1001; i++ {
		err = db.SetLatestVersion(i)
		s.Require().NoError(err)

		lv, err = db.GetLatestVersion()
		s.Require().NoError(err)
		s.Require().Equal(i, lv)
	}
}

func (s *StorageTestSuite) TestDatabase_VersionedKeys() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	for i := uint64(1); i <= 100; i++ {
		s.Require().NoError(db.ApplyChangeset(i, corestore.NewChangesetWithPairs(
			map[string]corestore.KVPairs{
				storeKey1: {{Key: []byte("key"), Value: []byte(fmt.Sprintf("value%03d", i))}},
			},
		)))
	}

	for i := uint64(1); i <= 100; i++ {
		bz, err := db.Get(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("value%03d", i), string(bz))
	}
}

func (s *StorageTestSuite) TestDatabase_GetVersionedKey() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// store a key at version 1
	s.Require().NoError(db.ApplyChangeset(1, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{
			storeKey1: {{Key: []byte("key"), Value: []byte("value001")}},
		},
	)))

	// assume chain progresses to version 10 w/o any changes to key
	bz, err := db.Get(storeKey1Bytes, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value001"), bz)

	ok, err := db.Has(storeKey1Bytes, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().True(ok)

	// chain progresses to version 11 with an update to key
	s.Require().NoError(db.ApplyChangeset(11, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{
			storeKey1: {{Key: []byte("key"), Value: []byte("value011")}},
		},
	)))

	bz, err = db.Get(storeKey1Bytes, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value001"), bz)

	ok, err = db.Has(storeKey1Bytes, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().True(ok)

	for i := uint64(11); i <= 14; i++ {
		bz, err = db.Get(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().Equal([]byte("value011"), bz)

		ok, err = db.Has(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	// chain progresses to version 15 with a delete to key
	s.Require().NoError(db.ApplyChangeset(15, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{storeKey1: {{Key: []byte("key"), Remove: true}}},
	)))

	// all queries up to version 14 should return the latest value
	for i := uint64(1); i <= 14; i++ {
		bz, err = db.Get(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().NotNil(bz)

		ok, err = db.Has(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().True(ok)
	}

	// all queries after version 15 should return nil
	for i := uint64(15); i <= 17; i++ {
		bz, err = db.Get(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().Nil(bz)

		ok, err = db.Has(storeKey1Bytes, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().False(ok)
	}
}

func (s *StorageTestSuite) TestDatabase_ApplyChangeset() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
	for i := 0; i < 100; i++ {
		cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(fmt.Sprintf("key%03d", i)), Value: []byte("value")})
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(fmt.Sprintf("key%03d", i)), Remove: true})
		}
	}

	s.Require().NoError(db.ApplyChangeset(1, cs))

	lv, err := db.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), lv)

	for i := 0; i < 1; i++ {
		ok, err := db.Has(storeKey1Bytes, 1, []byte(fmt.Sprintf("key%03d", i)))
		s.Require().NoError(err)

		if i%10 == 0 {
			s.Require().False(ok)
		} else {
			s.Require().True(ok)
		}
	}
}

func (s *StorageTestSuite) TestDatabase_IteratorEmptyDomain() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	iter, err := db.Iterator(storeKey1Bytes, 1, []byte{}, []byte{})
	s.Require().Error(err)
	s.Require().Nil(iter)
}

func (s *StorageTestSuite) TestDatabase_IteratorClose() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	iter, err := db.Iterator(storeKey1Bytes, 1, []byte("key000"), nil)
	s.Require().NoError(err)
	iter.Close()

	s.Require().False(iter.Valid())
}

func (s *StorageTestSuite) TestDatabase_IteratorDomain() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	testCases := map[string]struct {
		start, end []byte
	}{
		"start without end domain": {
			start: []byte("key010"),
		},
		"start and end domain": {
			start: []byte("key010"),
			end:   []byte("key020"),
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			iter, err := db.Iterator(storeKey1Bytes, 1, tc.start, tc.end)
			s.Require().NoError(err)

			defer iter.Close()

			start, end := iter.Domain()
			s.Require().Equal(tc.start, start)
			s.Require().Equal(tc.end, end)
		})
	}
}

func (s *StorageTestSuite) TestDatabase_Iterator() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val), Remove: false})
	}

	s.Require().NoError(db.ApplyChangeset(1, cs))

	// iterator without an end key over multiple versions
	for v := uint64(1); v < 5; v++ {
		itr, err := db.Iterator(storeKey1Bytes, v, []byte("key000"), nil)
		s.Require().NoError(err)

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i++
			count++
		}
		s.Require().NoError(itr.Error())
		s.Require().Equal(100, count)

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Valid())

		err = itr.Close()
		s.Require().NoError(err, "Failed to close iterator")
	}

	// iterator with a start and end domain over multiple versions
	for v := uint64(1); v < 5; v++ {
		itr2, err := db.Iterator(storeKey1Bytes, v, []byte("key010"), []byte("key019"))
		s.Require().NoError(err)

		i, count := 10, 0
		for ; itr2.Valid(); itr2.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr2.Key())
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr2.Value())

			i++
			count++
		}
		s.Require().Equal(9, count)
		s.Require().NoError(itr2.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr2.Valid())

		err = itr2.Close()
		if err != nil {
			return
		}
	}

	// start must be <= end
	iter3, err := db.Iterator(storeKey1Bytes, 1, []byte("key020"), []byte("key019"))
	s.Require().Error(err)
	s.Require().Nil(iter3)
}

func (s *StorageTestSuite) TestDatabase_Iterator_RangedDeletes() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	s.Require().NoError(db.ApplyChangeset(1, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{
			storeKey1: {
				{Key: []byte("key001"), Value: []byte("value001"), Remove: false},
				{Key: []byte("key002"), Value: []byte("value001"), Remove: false},
			},
		},
	)))

	s.Require().NoError(db.ApplyChangeset(5, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{
			storeKey1: {{Key: []byte("key002"), Value: []byte("value002"), Remove: false}},
		},
	)))

	s.Require().NoError(db.ApplyChangeset(10, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{
			storeKey1: {{Key: []byte("key002"), Remove: true}},
		},
	)))

	itr, err := db.Iterator(storeKey1Bytes, 11, []byte("key001"), nil)
	s.Require().NoError(err)

	defer itr.Close()

	// there should only be one valid key in the iterator -- key001
	var count int
	for ; itr.Valid(); itr.Next() {
		s.Require().Equal([]byte("key001"), itr.Key())
		count++
	}
	s.Require().Equal(1, count)
	s.Require().NoError(itr.Error())
}

func (s *StorageTestSuite) TestDatabase_IteratorMultiVersion() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// for versions 1-49, set all 10 keys
	for v := uint64(1); v < 50; v++ {
		cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val)})
		}

		s.Require().NoError(db.ApplyChangeset(v, cs))
	}

	// for versions 50-100, only update even keys
	for v := uint64(50); v <= 100; v++ {
		cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				key := fmt.Sprintf("key%03d", i)
				val := fmt.Sprintf("val%03d-%03d", i, v)

				cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val), Remove: false})
			}
		}

		s.Require().NoError(db.ApplyChangeset(v, cs))
	}

	itr, err := db.Iterator(storeKey1Bytes, 69, []byte("key000"), nil)
	s.Require().NoError(err)

	defer itr.Close()

	// All keys should be present; All odd keys should have a value that reflects
	// version 49, and all even keys should have a value that reflects the desired
	// version, 69.
	var i, count int
	for ; itr.Valid(); itr.Next() {
		s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

		if i%2 == 0 {
			s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, 69)), itr.Value())
		} else {
			s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, 49)), itr.Value())
		}

		i = (i + 1) % 10
		count++
	}

	s.Require().NoError(itr.Error())
	s.Require().Equal(10, count)
}

func (s *StorageTestSuite) TestDatabaseIterator_SkipVersion() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)

	defer db.Close()

	dbApplyChangeset(s.T(), db, 58827506, storeKey1, [][]byte{[]byte("keyC")}, [][]byte{[]byte("value003")})
	dbApplyChangeset(s.T(), db, 58827506, storeKey1, [][]byte{[]byte("keyE")}, [][]byte{[]byte("value000")})
	dbApplyChangeset(s.T(), db, 58827506, storeKey1, [][]byte{[]byte("keyF")}, [][]byte{[]byte("value000")})
	dbApplyChangeset(s.T(), db, 58833605, storeKey1, [][]byte{[]byte("keyC")}, [][]byte{[]byte("value004")})
	dbApplyChangeset(s.T(), db, 58833606, storeKey1, [][]byte{[]byte("keyD")}, [][]byte{[]byte("value006")})

	itr, err := db.Iterator(storeKey1Bytes, 58831525, []byte("key"), nil)
	s.Require().NoError(err)
	defer itr.Close()

	count := make(map[string]struct{})
	for ; itr.Valid(); itr.Next() {
		count[string(itr.Key())] = struct{}{}
	}

	s.Require().NoError(itr.Error())
	s.Require().Equal(3, len(count))
}

func (s *StorageTestSuite) TestDatabaseIterator_ForwardIteration() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	dbApplyChangeset(s.T(), db, 8, storeKey1, [][]byte{[]byte("keyA")}, [][]byte{[]byte("value001")})
	dbApplyChangeset(s.T(), db, 9, storeKey1, [][]byte{[]byte("keyB")}, [][]byte{[]byte("value002")})
	dbApplyChangeset(s.T(), db, 10, storeKey1, [][]byte{[]byte("keyC")}, [][]byte{[]byte("value003")})
	dbApplyChangeset(s.T(), db, 11, storeKey1, [][]byte{[]byte("keyD")}, [][]byte{[]byte("value004")})

	dbApplyChangeset(s.T(), db, 2, storeKey1, [][]byte{[]byte("keyD")}, [][]byte{[]byte("value007")})
	dbApplyChangeset(s.T(), db, 3, storeKey1, [][]byte{[]byte("keyE")}, [][]byte{[]byte("value008")})
	dbApplyChangeset(s.T(), db, 4, storeKey1, [][]byte{[]byte("keyF")}, [][]byte{[]byte("value009")})
	dbApplyChangeset(s.T(), db, 5, storeKey1, [][]byte{[]byte("keyH")}, [][]byte{[]byte("value010")})

	itr, err := db.Iterator(storeKey1Bytes, 6, nil, []byte("keyZ"))
	s.Require().NoError(err)

	defer itr.Close()
	count := 0
	for ; itr.Valid(); itr.Next() {
		count++
	}

	s.Require().NoError(itr.Error())
	s.Require().Equal(4, count)
}

func (s *StorageTestSuite) TestDatabaseIterator_ForwardIterationHigher() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	dbApplyChangeset(s.T(), db, 9, storeKey1, [][]byte{[]byte("keyB")}, [][]byte{[]byte("value002")})
	dbApplyChangeset(s.T(), db, 10, storeKey1, [][]byte{[]byte("keyC")}, [][]byte{[]byte("value003")})
	dbApplyChangeset(s.T(), db, 11, storeKey1, [][]byte{[]byte("keyD")}, [][]byte{[]byte("value004")})

	dbApplyChangeset(s.T(), db, 12, storeKey1, [][]byte{[]byte("keyD")}, [][]byte{[]byte("value007")})
	dbApplyChangeset(s.T(), db, 13, storeKey1, [][]byte{[]byte("keyE")}, [][]byte{[]byte("value008")})
	dbApplyChangeset(s.T(), db, 14, storeKey1, [][]byte{[]byte("keyF")}, [][]byte{[]byte("value009")})
	dbApplyChangeset(s.T(), db, 15, storeKey1, [][]byte{[]byte("keyH")}, [][]byte{[]byte("value010")})

	itr, err := db.Iterator(storeKey1Bytes, 6, nil, []byte("keyZ"))
	s.Require().NoError(err)

	defer itr.Close()

	count := 0
	for ; itr.Valid(); itr.Next() {
		count++
	}

	s.Require().Equal(0, count)
}

func (s *StorageTestSuite) TestDatabase_IteratorNoDomain() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// for versions 1-50, set all 10 keys
	for v := uint64(1); v <= 50; v++ {
		cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val), Remove: false})
		}

		s.Require().NoError(db.ApplyChangeset(v, cs))
	}

	// create an iterator over the entire domain
	itr, err := db.Iterator(storeKey1Bytes, 50, nil, nil)
	s.Require().NoError(err)

	defer itr.Close()

	var i, count int
	for ; itr.Valid(); itr.Next() {
		s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
		s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, 50)), itr.Value())

		i++
		count++
	}
	s.Require().NoError(itr.Error())
	s.Require().Equal(10, count)
}

func (s *StorageTestSuite) TestDatabase_Prune() {
	if slices.Contains(s.SkipTests, s.T().Name()) {
		s.T().SkipNow()
	}

	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// for versions 1-50, set 10 keys
	for v := uint64(1); v <= 50; v++ {
		cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val)})
		}

		s.Require().NoError(db.ApplyChangeset(v, cs))
	}

	// prune the first 25 versions
	s.Require().NoError(db.Prune(25))

	latestVersion, err := db.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(50), latestVersion)

	// Ensure all keys are no longer present up to and including version 25 and
	// all keys are present after version 25.
	for v := uint64(1); v <= 50; v++ {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			bz, err := db.Get(storeKey1Bytes, v, []byte(key))
			if v <= 25 {
				s.Require().Error(err)
				s.Require().Nil(bz)
			} else {
				s.Require().NoError(err)
				s.Require().Equal([]byte(val), bz)
			}
		}
	}

	itr, err := db.Iterator(storeKey1Bytes, 25, []byte("key000"), nil)
	s.Require().NoError(err)
	s.Require().False(itr.Valid())

	// prune the latest version which should prune the entire dataset
	s.Require().NoError(db.Prune(50))

	for v := uint64(1); v <= 50; v++ {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)

			bz, err := db.Get(storeKey1Bytes, v, []byte(key))
			s.Require().Error(err)
			s.Require().Nil(bz)
		}
	}
}

func (s *StorageTestSuite) TestDatabase_Prune_KeepRecent() {
	if slices.Contains(s.SkipTests, s.T().Name()) {
		s.T().SkipNow()
	}

	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	key := []byte("key")

	// write a key at three different versions
	s.Require().NoError(db.ApplyChangeset(1, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{storeKey1: {{Key: key, Value: []byte("val001"), Remove: false}}},
	)))
	s.Require().NoError(db.ApplyChangeset(100, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{storeKey1: {{Key: key, Value: []byte("val100"), Remove: false}}},
	)))
	s.Require().NoError(db.ApplyChangeset(200, corestore.NewChangesetWithPairs(
		map[string]corestore.KVPairs{storeKey1: {{Key: key, Value: []byte("val200"), Remove: false}}},
	)))

	// prune version 50
	s.Require().NoError(db.Prune(50))

	// ensure queries for versions 50 and older return nil
	bz, err := db.Get(storeKey1Bytes, 49, key)
	s.Require().Error(err)
	s.Require().Nil(bz)

	itr, err := db.Iterator(storeKey1Bytes, 49, nil, nil)
	s.Require().NoError(err)
	s.Require().False(itr.Valid())

	defer itr.Close()

	// ensure the value previously at version 1 is still there for queries greater than 50
	bz, err = db.Get(storeKey1Bytes, 51, key)
	s.Require().NoError(err)
	s.Require().Equal([]byte("val001"), bz)

	// ensure the correct value at a greater height
	bz, err = db.Get(storeKey1Bytes, 200, key)
	s.Require().NoError(err)
	s.Require().Equal([]byte("val200"), bz)

	// prune latest height and ensure we have the previous version when querying above it
	s.Require().NoError(db.Prune(200))

	bz, err = db.Get(storeKey1Bytes, 201, key)
	s.Require().NoError(err)
	s.Require().Equal([]byte("val200"), bz)
}

func (s *StorageTestSuite) TestDatabase_Restore() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	toVersion := uint64(10)
	keyCount := 10

	// for versions 1-10, set 10 keys
	for v := uint64(1); v <= toVersion; v++ {
		cs := corestore.NewChangesetWithPairs(map[string]corestore.KVPairs{storeKey1: {}})
		for i := 0; i < keyCount; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			cs.AddKVPair(storeKey1Bytes, corestore.KVPair{Key: []byte(key), Value: []byte(val)})
		}

		s.Require().NoError(db.ApplyChangeset(v, cs))
	}

	latestVersion, err := db.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(10), latestVersion)

	chStorage := make(chan *corestore.StateChanges, 5)

	go func() {
		for i := uint64(11); i <= 15; i++ {
			kvPairs := []corestore.KVPair{}
			for j := 0; j < keyCount; j++ {
				key := fmt.Sprintf("key%03d-%03d", j, i)
				val := fmt.Sprintf("val%03d-%03d", j, i)

				kvPairs = append(kvPairs, corestore.KVPair{Key: []byte(key), Value: []byte(val)})
			}
			chStorage <- &corestore.StateChanges{
				Actor:        storeKey1Bytes,
				StateChanges: kvPairs,
			}
		}
		close(chStorage)
	}()

	// restore with snapshot version smaller than latest version
	// should return an error
	err = db.Restore(9, chStorage)
	s.Require().Error(err)

	// restore
	err = db.Restore(11, chStorage)
	s.Require().NoError(err)

	// check the storage
	for i := uint64(11); i <= 15; i++ {
		for j := 0; j < keyCount; j++ {
			key := fmt.Sprintf("key%03d-%03d", j, i)
			val := fmt.Sprintf("val%03d-%03d", j, i)

			v, err := db.Get(storeKey1Bytes, 11, []byte(key))
			s.Require().NoError(err)
			s.Require().Equal([]byte(val), v)
		}
	}
}

func (s *StorageTestSuite) TestUpgradable() {
	ss, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer ss.Close()

	// Ensure the database is upgradable.
	if _, ok := ss.db.(store.UpgradableDatabase); !ok {
		s.T().Skip("database is not upgradable")
	}

	storeKeys := []string{"store1", "store2", "store3"}
	uptoVersion := uint64(50)
	keyCount := 10
	for _, storeKey := range storeKeys {
		for v := uint64(1); v <= uptoVersion; v++ {
			keys := make([][]byte, keyCount)
			vals := make([][]byte, keyCount)
			for i := 0; i < keyCount; i++ {
				keys[i] = []byte(fmt.Sprintf("key%03d", i))
				vals[i] = []byte(fmt.Sprintf("val%03d-%03d", i, v))
			}
			dbApplyChangeset(s.T(), ss, v, storeKey, keys, vals)
		}
	}

	// prune storekeys (`store2`, `store3`)
	removedStoreKeys := []string{storeKeys[1], storeKeys[2]}
	err = ss.PruneStoreKeys(removedStoreKeys, uptoVersion)
	s.Require().NoError(err)
	// should be able to query before Prune for removed storeKeys
	for _, storeKey := range removedStoreKeys {
		for v := uint64(1); v <= uptoVersion; v++ {
			for i := 0; i < keyCount; i++ {
				bz, err := ss.Get([]byte(storeKey), v, []byte(fmt.Sprintf("key%03d", i)))
				s.Require().NoError(err)
				s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, v)), bz)
			}
		}
	}
	s.Require().NoError(ss.Prune(uptoVersion))
	// should not be able to query after Prune
	// skip the test of RocksDB
	if !slices.Contains(s.SkipTests, "TestUpgradable_Prune") {
		for _, storeKey := range removedStoreKeys {
			// it will return error ErrVersionPruned
			for v := uint64(1); v <= uptoVersion; v++ {
				for i := 0; i < keyCount; i++ {
					_, err := ss.Get([]byte(storeKey), v, []byte(fmt.Sprintf("key%03d", i)))
					s.Require().Error(err)
				}
			}
			v := uptoVersion + 1
			for i := 0; i < keyCount; i++ {
				val, err := ss.Get([]byte(storeKey), v, []byte(fmt.Sprintf("key%03d", i)))
				s.Require().NoError(err)
				s.Require().Nil(val)
			}
		}
	}
}

func (s *StorageTestSuite) TestRemovingOldStoreKey() {
	ss, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer ss.Close()

	// Ensure the database is upgradable.
	if _, ok := ss.db.(store.UpgradableDatabase); !ok {
		s.T().Skip("database is not upgradable")
	}

	storeKeys := []string{"store1", "store2", "store3"}
	uptoVersion := uint64(50)
	keyCount := 10
	for _, storeKey := range storeKeys {
		for v := uint64(1); v <= uptoVersion; v++ {
			keys := make([][]byte, keyCount)
			vals := make([][]byte, keyCount)
			for i := 0; i < keyCount; i++ {
				keys[i] = []byte(fmt.Sprintf("key%03d-%03d", i, v))
				vals[i] = []byte(fmt.Sprintf("val%03d-%03d", i, v))
			}
			dbApplyChangeset(s.T(), ss, v, storeKey, keys, vals)
		}
	}

	// remove `store1` and `store3`
	removedStoreKeys := []string{storeKeys[0], storeKeys[2]}
	err = ss.PruneStoreKeys(removedStoreKeys, uptoVersion)
	s.Require().NoError(err)
	// should be able to query before Prune for removed storeKeys
	for _, storeKey := range removedStoreKeys {
		for v := uint64(1); v <= uptoVersion; v++ {
			for i := 0; i < keyCount; i++ {
				bz, err := ss.Get([]byte(storeKey), v, []byte(fmt.Sprintf("key%03d-%03d", i, v)))
				s.Require().NoError(err)
				s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, v)), bz)
			}
		}
	}
	// add `store1` back
	newStoreKeys := []string{storeKeys[0], storeKeys[1]}
	newVersion := uptoVersion + 10
	for _, storeKey := range newStoreKeys {
		for v := uptoVersion + 1; v <= newVersion; v++ {
			keys := make([][]byte, keyCount)
			vals := make([][]byte, keyCount)
			for i := 0; i < keyCount; i++ {
				keys[i] = []byte(fmt.Sprintf("key%03d-%03d", i, v))
				vals[i] = []byte(fmt.Sprintf("val%03d-%03d", i, v))
			}
			dbApplyChangeset(s.T(), ss, v, storeKey, keys, vals)
		}
	}

	s.Require().NoError(ss.Prune(newVersion))
	// skip the test of RocksDB
	if !slices.Contains(s.SkipTests, "TestUpgradable_Prune") {
		for _, storeKey := range removedStoreKeys {
			queryVersion := newVersion + 1
			// should not be able to query after Prune during 1 ~ uptoVersion
			for v := uint64(1); v <= uptoVersion; v++ {
				for i := 0; i < keyCount; i++ {
					val, err := ss.Get([]byte(storeKey), queryVersion, []byte(fmt.Sprintf("key%03d", i)))
					s.Require().NoError(err)
					s.Require().Nil(val)
				}
			}
			// should be able to query after Prune during uptoVersion + 1 ~ newVersion
			// for `store1` added back
			for v := uptoVersion + 1; v <= newVersion; v++ {
				for i := 0; i < keyCount; i++ {
					val, err := ss.Get([]byte(storeKey), queryVersion, []byte(fmt.Sprintf("key%03d-%03d", i, v)))
					s.Require().NoError(err)
					if storeKey == storeKeys[0] {
						// `store1` is added back
						s.Require().Equal([]byte(fmt.Sprintf("val%03d-%03d", i, v)), val)
					} else {
						// `store3` is removed
						s.Require().Nil(val)
					}
				}
			}
		}
	}
}

func dbApplyChangeset(
	t *testing.T,
	db store.VersionedDatabase,
	version uint64,
	storeKey string,
	keys, vals [][]byte,
) {
	t.Helper()

	require.Greater(t, version, uint64(0))
	require.Equal(t, len(keys), len(vals))

	cs := corestore.NewChangeset()
	for i := 0; i < len(keys); i++ {
		remove := false
		if vals[i] == nil {
			remove = true
		}

		cs.AddKVPair([]byte(storeKey), corestore.KVPair{Key: keys[i], Value: vals[i], Remove: remove})
	}

	require.NoError(t, db.ApplyChangeset(version, cs))
}
