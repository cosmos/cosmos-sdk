package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	corestore "cosmossdk.io/core/store"
)

type DBTestSuite struct {
	suite.Suite

	db corestore.KVStoreWithBatch
}

func (s *DBTestSuite) TearDownSuite() {
	s.Require().NoError(s.db.Close())
}

func (s *DBTestSuite) TestDBOperations() {
	// Batch Set
	b := s.db.NewBatch()
	s.Require().NoError(b.Set([]byte("key"), []byte("value")))
	s.Require().NoError(b.Set([]byte("key1"), []byte("value1")))
	s.Require().NoError(b.Set([]byte("key2"), []byte("value2")))
	s.Require().NoError(b.Write())

	// Get
	value, err := s.db.Get([]byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value"), value)

	// Has
	has, err := s.db.Has([]byte("key1"))
	s.Require().NoError(err)
	s.Require().True(has)
	has, err = s.db.Has([]byte("key3"))
	s.Require().NoError(err)
	s.Require().False(has)

	// Batch Delete
	b = s.db.NewBatch()
	s.Require().NoError(b.Delete([]byte("key1")))
	s.Require().NoError(b.Write())

	// Has
	has, err = s.db.Has([]byte("key1"))
	s.Require().NoError(err)
	s.Require().False(has)

	// Set & Delete
	s.Require().NoError(s.db.Set([]byte("key3"), []byte("value3")))
	has, err = s.db.Has([]byte("key3"))
	s.Require().NoError(err)
	s.Require().True(has)
	value, err = s.db.Get([]byte("key3"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value3"), value)
	s.Require().NoError(s.db.Delete([]byte("key3")))
	has, err = s.db.Has([]byte("key3"))
	s.Require().NoError(err)
	s.Require().False(has)
	value, err = s.db.Get([]byte("key3"))
	s.Require().NoError(err)
	s.Require().Nil(value)
}

func (s *DBTestSuite) TestIterator() {
	// Set
	b := s.db.NewBatch()
	for i := 0; i < 10; i++ {
		s.Require().NoError(b.Set([]byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i))))
	}
	s.Require().NoError(b.Write())

	// Iterator
	itr, err := s.db.Iterator(nil, nil)
	s.Require().NoError(err)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		value1, err := s.db.Get(key)
		s.Require().NoError(err)
		s.Require().Equal(value1, value)
	}

	// Reverse Iterator
	ritr, err := s.db.ReverseIterator([]byte("key0"), []byte("keys"))
	s.Require().NoError(err)
	defer ritr.Close()

	index := 9
	for ; ritr.Valid(); ritr.Next() {
		key := ritr.Key()
		value := ritr.Value()
		s.Require().Equal([]byte(fmt.Sprintf("key%d", index)), key)
		value1, err := s.db.Get(key)
		s.Require().NoError(err)
		s.Require().Equal(value1, value)
		index -= 1
	}
	s.Require().Equal(-1, index)
}

func TestMemDBSuite(t *testing.T) {
	suite.Run(t, &DBTestSuite{
		db: NewMemDB(),
	})
}

func TestPebbleDBSuite(t *testing.T) {
	db, err := NewPebbleDB("test", t.TempDir())
	require.NoError(t, err)

	suite.Run(t, &DBTestSuite{
		db: db,
	})
}

func TestGoLevelDBSuite(t *testing.T) {
	db, err := NewGoLevelDB("test", t.TempDir(), nil)
	require.NoError(t, err)
	suite.Run(t, &DBTestSuite{
		db: db,
	})
}

func TestPrefixDBSuite(t *testing.T) {
	suite.Run(t, &DBTestSuite{
		db: NewPrefixDB(NewMemDB(), []byte("prefix")),
	})
}
