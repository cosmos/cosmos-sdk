package mem_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/kv/mem"
)

const storeKey = "storeKey"

type StoreTestSuite struct {
	suite.Suite

	kvStore store.KVStore
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{})
}

func (s *StoreTestSuite) SetupTest() {
	s.kvStore = mem.New(storeKey)
}

func (s *StoreTestSuite) TestGetStoreType() {
	s.Require().Equal(store.StoreTypeMem, s.kvStore.GetStoreType())
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

func (s *StoreTestSuite) TestCRUD() {
	bz := s.kvStore.Get([]byte("key000"))
	s.Require().Nil(bz)
	s.Require().False(s.kvStore.Has([]byte("key000")))

	s.kvStore.Set([]byte("key000"), []byte("val000"))

	bz = s.kvStore.Get([]byte("key000"))
	s.Require().Equal([]byte("val000"), bz)
	s.Require().True(s.kvStore.Has([]byte("key000")))

	s.kvStore.Delete([]byte("key000"))

	bz = s.kvStore.Get([]byte("key000"))
	s.Require().Nil(bz)
	s.Require().False(s.kvStore.Has([]byte("key000")))
}

func (s *StoreTestSuite) TestIterator() {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		s.kvStore.Set([]byte(key), []byte(val))
	}

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

func (s *StoreTestSuite) TestReverseIterator() {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		s.kvStore.Set([]byte(key), []byte(val))
	}

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
		s.Require().False(itr.Valid())
		s.Require().False(itr.Next())
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
