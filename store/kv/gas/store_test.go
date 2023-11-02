package gas_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/kv/gas"
	"cosmossdk.io/store/v2/kv/mem"
)

const (
	storeKey = "storeKey"
	gasLimit = store.Gas(1_000_000)
)

type StoreTestSuite struct {
	suite.Suite

	parent     store.KVStore
	gasKVStore store.BranchedKVStore
	gasMeter   store.GasMeter
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, &StoreTestSuite{})
}

func (s *StoreTestSuite) SetupTest() {
	s.parent = mem.New(storeKey)
	s.gasMeter = store.NewGasMeter(gasLimit)
	s.gasKVStore = gas.New(s.parent, s.gasMeter, store.DefaultGasConfig())
}

func (s *StoreTestSuite) TearDownTest() {
	err := s.gasKVStore.Reset(1)
	s.Require().NoError(err)
}

func (s *StoreTestSuite) TestGetStoreKey() {
	s.Require().Equal(s.parent.GetStoreKey(), s.gasKVStore.GetStoreKey())
}

func (s *StoreTestSuite) TestGetStoreType() {
	s.Require().Equal(s.parent.GetStoreType(), s.gasKVStore.GetStoreType())
}

func (s *StoreTestSuite) TestGet() {
	key, value := []byte("key"), []byte("value")
	s.parent.Set(key, value)

	s.Require().Equal(value, s.gasKVStore.Get(key))
	s.Require().Equal(store.Gas(1024), s.gasMeter.GasConsumed())
}

func (s *StoreTestSuite) TestHas() {
	key, value := []byte("key"), []byte("value")
	s.parent.Set(key, value)

	s.Require().True(s.gasKVStore.Has(key))
	s.Require().Equal(store.Gas(1000), s.gasMeter.GasConsumed())
}

func (s *StoreTestSuite) TestSet() {
	s.gasKVStore.Set([]byte("key"), []byte("value"))
	s.Require().Equal(store.Gas(2240), s.gasMeter.GasConsumed())
}

func (s *StoreTestSuite) TestDelete() {
	key, value := []byte("key"), []byte("value")
	s.parent.Set(key, value)

	s.gasKVStore.Delete(key)
	s.Require().Equal(store.Gas(1500), s.gasMeter.GasConsumed())
}

func (s *StoreTestSuite) TestIterator() {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		s.parent.Set([]byte(key), []byte(val))
	}

	itr := s.gasKVStore.Iterator(nil, nil)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		_ = itr.Key()
		_ = itr.Value()
	}
	s.Require().Equal(store.Gas(6600), s.gasMeter.GasConsumed())
}

func (s *StoreTestSuite) TestReverseIterator() {
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		s.parent.Set([]byte(key), []byte(val))
	}

	itr := s.gasKVStore.ReverseIterator(nil, nil)
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		_ = itr.Key()
		_ = itr.Value()
	}
	s.Require().Equal(store.Gas(6600), s.gasMeter.GasConsumed())
}
