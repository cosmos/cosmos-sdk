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
