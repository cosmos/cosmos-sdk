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
	s.Require().Equal([]byte("key001"), b.GetChangeset().Pairs[0].Key)

	// write the branched store and ensure all writes are flushed to the parent
	b.Write()
	s.Require().Equal([]byte("branched_updated_val001"), s.kvStore.Get([]byte("key001")))

	s.Require().Equal(2, s.kvStore.GetChangeset().Size())
}
