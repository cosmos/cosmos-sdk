package branch_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/branch"
	"cosmossdk.io/store/v2/storage/sqlite"
	"github.com/stretchr/testify/suite"
)

const storeKey = "storeKey"

type StoreTestSuite struct {
	suite.Suite

	storage store.VersionedDatabase
	kvStore store.KVStore
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{})
}

func (s *StoreTestSuite) SetupTest() {
	storage, err := sqlite.New(s.T().TempDir())
	s.Require().NoError(err)

	batch, err := storage.NewBatch(1)
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		err = batch.Set(storeKey, []byte(key), []byte(val))
		s.Require().NoError(err)
	}

	err = batch.Write()
	s.Require().NoError(err)

	kvStore, err := branch.New(storeKey, storage)
	s.Require().NoError(err)

	s.storage = storage
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
	s.Require().NoError(s.kvStore.Reset())

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
