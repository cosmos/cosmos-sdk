package branch_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/kv/branch"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/sqlite"
)

const storeKey = "storeKey"

type StoreTestSuite struct {
	suite.Suite

	storage store.VersionedDatabase
	kvStore store.BranchedKVStore
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{})
}

func (s *StoreTestSuite) SetupTest() {
	sqliteDB, err := sqlite.New(s.T().TempDir())
	ss := storage.NewStorageStore(sqliteDB)
	s.Require().NoError(err)

	cs := store.NewChangeset(map[string]store.KVPairs{storeKey: {}})
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099

		cs.AddKVPair(storeKey, store.KVPair{Key: []byte(key), Value: []byte(val)})
	}

	s.Require().NoError(ss.ApplyChangeset(1, cs))

	kvStore, err := branch.New(storeKey, ss)
	s.Require().NoError(err)

	s.storage = ss
	s.kvStore = kvStore
}

func (s *StoreTestSuite) TestGetStoreType() {
	s.Require().Equal(store.StoreTypeBranch, s.kvStore.GetStoreType())
}

func (s *StoreTestSuite) TestGetChangeset() {
	// initial store with no writes should have an empty changeset
	cs := s.kvStore.GetChangeset()
	s.Require().Zero(cs.Size())

	// perform some writes
	s.kvStore.Set([]byte("key000"), []byte("updated_val000"))
	s.kvStore.Delete([]byte("key001"))

	cs = s.kvStore.GetChangeset()
	s.Require().Equal(cs.Size(), 2)
}

func (s *StoreTestSuite) TestReset() {
	s.Require().NoError(s.kvStore.Reset(1))

	cs := s.kvStore.GetChangeset()
	s.Require().Zero(cs.Size())
}

func (s *StoreTestSuite) TestGet() {
	// perform read of key000, which is not dirty
	bz := s.kvStore.Get([]byte("key000"))
	s.Require().Equal([]byte("val000"), bz)

	// update key000 and perform a read which should reflect the new value
	s.kvStore.Set([]byte("key000"), []byte("updated_val000"))

	bz = s.kvStore.Get([]byte("key000"))
	s.Require().Equal([]byte("updated_val000"), bz)

	// ensure the primary SS backend is not modified
	bz, err := s.storage.Get(storeKey, 1, []byte("key000"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("val000"), bz)
}

func (s *StoreTestSuite) TestHas() {
	// perform read of key000, which is not dirty thus falling back to SS
	ok := s.kvStore.Has([]byte("key000"))
	s.Require().True(ok)

	ok = s.kvStore.Has([]byte("key100"))
	s.Require().False(ok)

	// perform a write of a brand new key not in SS, but in the changeset
	s.kvStore.Set([]byte("key100"), []byte("val100"))

	ok = s.kvStore.Has([]byte("key100"))
	s.Require().True(ok)
}

func (s *StoreTestSuite) TestBranch() {
	// perform a few writes on the original store
	s.kvStore.Set([]byte("key000"), []byte("updated_val000"))
	s.kvStore.Set([]byte("key001"), []byte("updated_val001"))

	// create a new branch
	b := s.kvStore.Branch()

	// update an existing dirty write
	b.Set([]byte("key001"), []byte("branched_updated_val001"))

	// perform reads on the branched store without writing first

	// key000 is dirty in the original store, but not in the branched store
	s.Require().Equal([]byte("updated_val000"), b.Get([]byte("key000")))

	// key001 is dirty in both the original and branched store, but branched store
	// should reflect the branched write.
	s.Require().Equal([]byte("branched_updated_val001"), b.Get([]byte("key001")))

	// key002 is not dirty in either store, so should fall back to SS
	s.Require().Equal([]byte("val002"), b.Get([]byte("key002")))

	// ensure the original store is not modified
	s.Require().Equal([]byte("updated_val001"), s.kvStore.Get([]byte("key001")))

	s.Require().Equal(1, b.GetChangeset().Size())
	s.Require().Equal([]byte("key001"), b.GetChangeset().Pairs[storeKey][0].Key)

	// write the branched store and ensure all writes are flushed to the parent
	b.Write()
	s.Require().Equal([]byte("branched_updated_val001"), s.kvStore.Get([]byte("key001")))

	s.Require().Equal(2, s.kvStore.GetChangeset().Size())
}

func (s *StoreTestSuite) TestIterator_NoWrites() {
	// iterator without an end domain
	s.Run("start_only", func() {
		itr := s.kvStore.Iterator([]byte("key000"), nil)
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i++
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator without a start domain
	s.Run("end_only", func() {
		itr := s.kvStore.Iterator(nil, []byte("key100"))
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i++
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with with a start and end domain
	s.Run("start_and_end", func() {
		itr := s.kvStore.Iterator([]byte("key000"), []byte("key050"))
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i++
			count++
		}
		s.Require().Equal(50, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with an open domain
	s.Run("open_domain", func() {
		itr := s.kvStore.Iterator(nil, nil)
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i++
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})
}

func (s *StoreTestSuite) TestIterator_DirtyWrites() {
	// modify all even keys
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			key := fmt.Sprintf("key%03d", i)         // key000, key002, ...
			val := fmt.Sprintf("updated_val%03d", i) // updated_val000, updated_val002, ...
			s.kvStore.Set([]byte(key), []byte(val))
		}
	}

	// add some new keys to ensure we cover those as well
	for i := 100; i < 150; i++ {
		key := fmt.Sprintf("key%03d", i) // key100, key101, ...
		val := fmt.Sprintf("val%03d", i) // val100, val101, ...
		s.kvStore.Set([]byte(key), []byte(val))
	}

	// iterator without an end domain
	s.Run("start_only", func() {
		itr := s.kvStore.Iterator([]byte("key000"), nil)
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i++
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator without a start domain
	s.Run("end_only", func() {
		itr := s.kvStore.Iterator(nil, []byte("key150"))
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i++
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with with a start and end domain
	s.Run("start_and_end", func() {
		itr := s.kvStore.Iterator([]byte("key000"), []byte("key050"))
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i++
			count++
		}
		s.Require().Equal(50, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with an open domain
	s.Run("open_domain", func() {
		itr := s.kvStore.Iterator(nil, nil)
		defer itr.Close()

		var i, count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i++
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})
}

func (s *StoreTestSuite) TestReverseIterator_NoWrites() {
	// reverse iterator without an end domain
	s.Run("start_only", func() {
		itr := s.kvStore.ReverseIterator([]byte("key000"), nil)
		defer itr.Close()

		i := 99
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i--
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// reverse iterator without a start domain
	s.Run("end_only", func() {
		itr := s.kvStore.ReverseIterator(nil, []byte("key100"))
		defer itr.Close()

		i := 99
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i--
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// reverse iterator with with a start and end domain
	s.Run("start_and_end", func() {
		itr := s.kvStore.ReverseIterator([]byte("key000"), []byte("key050"))
		defer itr.Close()

		i := 49
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i--
			count++
		}
		s.Require().Equal(50, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// reverse iterator with an open domain
	s.Run("open_domain", func() {
		itr := s.kvStore.ReverseIterator(nil, nil)
		defer itr.Close()

		i := 99
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))
			s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())

			i--
			count++
		}
		s.Require().Equal(100, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})
}

func (s *StoreTestSuite) TestReverseIterator_DirtyWrites() {
	// modify all even keys
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			key := fmt.Sprintf("key%03d", i)         // key000, key002, ...
			val := fmt.Sprintf("updated_val%03d", i) // updated_val000, updated_val002, ...
			s.kvStore.Set([]byte(key), []byte(val))
		}
	}

	// add some new keys to ensure we cover those as well
	for i := 100; i < 150; i++ {
		key := fmt.Sprintf("key%03d", i) // key100, key101, ...
		val := fmt.Sprintf("val%03d", i) // val100, val101, ...
		s.kvStore.Set([]byte(key), []byte(val))
	}

	// reverse iterator without an end domain
	s.Run("start_only", func() {
		itr := s.kvStore.ReverseIterator([]byte("key000"), nil)
		defer itr.Close()

		i := 149
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), "itr_key: %s, count: %d", string(itr.Key()), count)

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i--
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator without a start domain
	s.Run("end_only", func() {
		itr := s.kvStore.ReverseIterator(nil, []byte("key150"))
		defer itr.Close()

		i := 149
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i--
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with with a start and end domain
	s.Run("start_and_end", func() {
		itr := s.kvStore.ReverseIterator([]byte("key000"), []byte("key050"))
		defer itr.Close()

		i := 49
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i--
			count++
		}
		s.Require().Equal(50, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})

	// iterator with an open domain
	s.Run("open_domain", func() {
		itr := s.kvStore.ReverseIterator(nil, nil)
		defer itr.Close()

		i := 149
		var count int
		for ; itr.Valid(); itr.Next() {
			s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr.Key(), string(itr.Key()))

			if i%2 == 0 && i < 100 {
				s.Require().Equal([]byte(fmt.Sprintf("updated_val%03d", i)), itr.Value())
			} else {
				s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr.Value())
			}

			i--
			count++
		}
		s.Require().Equal(150, count)
		s.Require().NoError(itr.Error())

		// seek past domain, which should make the iterator invalid and produce an error
		s.Require().False(itr.Next())
		s.Require().False(itr.Valid())
	})
}
