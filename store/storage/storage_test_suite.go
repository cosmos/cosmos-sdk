package storage

import (
	"fmt"

	"cosmossdk.io/store/v2"
	"github.com/stretchr/testify/suite"
)

const (
	storeKey1 = "store1"
)

// StorageTestSuite defines a reusable test suite for all storage backends.
type StorageTestSuite struct {
	suite.Suite

	NewDB func(dir string) (store.VersionedDatabase, error)
}

func (s *StorageTestSuite) TestDatabase_Close() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	s.Require().NoError(db.Close())
}

func (s *StorageTestSuite) TestDatabase_LatestVersion() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	lv, err := db.GetLatestVersion()
	s.Require().Error(err)
	s.Require().Zero(lv)

	for i := uint64(1); i <= 1001; i++ {
		err = db.SetLatestVersion(i)
		s.Require().NoError(err)

		lv, err = db.GetLatestVersion()
		s.Require().NoError(err)
		s.Require().Equal(i, lv)
	}
}

func (s *StorageTestSuite) TestDatabase_CRUD() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	ok, err := db.Has(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)
	s.Require().False(ok)

	err = db.Set(storeKey1, 1, []byte("key"), []byte("value"))
	s.Require().NoError(err)

	ok, err = db.Has(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)
	s.Require().True(ok)

	val, err := db.Get(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value"), val)

	err = db.Delete(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)

	ok, err = db.Has(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)
	s.Require().False(ok)

	val, err = db.Get(storeKey1, 1, []byte("key"))
	s.Require().NoError(err)
	s.Require().Nil(val)
}

func (s *StorageTestSuite) TestDatabase_VersionedKeys() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	for i := uint64(1); i <= 100; i++ {
		err := db.Set(storeKey1, i, []byte("key"), []byte(fmt.Sprintf("value%03d", i)))
		s.Require().NoError(err)
	}

	for i := uint64(1); i <= 100; i++ {
		bz, err := db.Get(storeKey1, i, []byte("key"))
		s.Require().NoError(err)
		s.Require().Equal(fmt.Sprintf("value%03d", i), string(bz))
	}
}

func (s *StorageTestSuite) TestDatabase_GetVersionedKey() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// store a key at version 1
	err = db.Set(storeKey1, 1, []byte("key"), []byte("value001"))
	s.Require().NoError(err)

	// assume chain progresses to version 10 w/o any changes to key
	bz, err := db.Get(storeKey1, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value001"), bz)

	ok, err := db.Has(storeKey1, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().True(ok)

	// chain progresses to version 11 with an update to key
	err = db.Set(storeKey1, 11, []byte("key"), []byte("value011"))
	s.Require().NoError(err)

	bz, err = db.Get(storeKey1, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().Equal([]byte("value001"), bz)

	ok, err = db.Has(storeKey1, 10, []byte("key"))
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *StorageTestSuite) TestDatabase_Batch() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		err = batch.Set(storeKey1, []byte(fmt.Sprintf("key%03d", i)), []byte("value"))
		s.Require().NoError(err)
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			err = batch.Delete(storeKey1, []byte(fmt.Sprintf("key%03d", i)))
			s.Require().NoError(err)
		}
	}

	s.Require().NotZero(batch.Size())

	err = batch.Write()
	s.Require().NoError(err)

	lv, err := db.GetLatestVersion()
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), lv)

	for i := 0; i < 1; i++ {
		ok, err := db.Has(storeKey1, 1, []byte(fmt.Sprintf("key%03d", i)))
		s.Require().NoError(err)

		if i%10 == 0 {
			s.Require().False(ok)
		} else {
			s.Require().True(ok)
		}
	}
}

func (s *StorageTestSuite) TestDatabase_ResetBatch() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	batch, err := db.NewBatch(1)
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		err = batch.Set(storeKey1, []byte(fmt.Sprintf("key%d", i)), []byte("value"))
		s.Require().NoError(err)
	}

	for i := 0; i < 100; i++ {
		if i%10 == 0 {
			err = batch.Delete(storeKey1, []byte(fmt.Sprintf("key%d", i)))
			s.Require().NoError(err)
		}
	}

	s.Require().NotZero(batch.Size())
	batch.Reset()
	s.Require().NotPanics(func() { batch.Reset() })

	// There is an initial cost of 12 bytes for the batch header
	s.Require().LessOrEqual(batch.Size(), 12)
}

func (s *StorageTestSuite) TestDatabase_IteratorEmptyDomain() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	iter, err := db.NewIterator(storeKey1, 1, []byte{}, []byte{})
	s.Require().Error(err)
	s.Require().Nil(iter)
}

func (s *StorageTestSuite) TestDatabase_IteratorClose() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	iter, err := db.NewIterator(storeKey1, 1, []byte("key000"), nil)
	s.Require().NoError(err)
	iter.Close()

	s.Require().False(iter.Valid())
	s.Require().Panics(func() { iter.Close() })
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
			iter, err := db.NewIterator(storeKey1, 1, tc.start, tc.end)
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

	batch, err := db.NewBatch(1)
	s.Require().NoError(err)

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i) // key000, key001, ..., key099
		val := fmt.Sprintf("val%03d", i) // val000, val001, ..., val099
		err = batch.Set(storeKey1, []byte(key), []byte(val))
		s.Require().NoError(err)
	}

	err = batch.Write()
	s.Require().NoError(err)

	// iterator without an end key
	itr, err := db.NewIterator(storeKey1, 1, []byte("key000"), nil)
	s.Require().NoError(err)

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

	// iterator with with a start and end domain
	itr2, err := db.NewIterator(storeKey1, 1, []byte("key010"), []byte("key019"))
	s.Require().NoError(err)

	defer itr2.Close()

	i, count = 10, 0
	for ; itr2.Valid(); itr2.Next() {
		s.Require().Equal([]byte(fmt.Sprintf("key%03d", i)), itr2.Key())
		s.Require().Equal([]byte(fmt.Sprintf("val%03d", i)), itr2.Value())

		i++
		count++
	}
	s.Require().Equal(9, count)
	s.Require().NoError(itr2.Error())

	// seek past domain, which should make the iterator invalid and produce an error
	s.Require().False(itr2.Next())
	s.Require().False(itr2.Valid())

	// start must be <= end
	iter3, err := db.NewIterator(storeKey1, 1, []byte("key020"), []byte("key019"))
	s.Require().Error(err)
	s.Require().Nil(iter3)
}

func (s *StorageTestSuite) TestDatabase_IteratorMultiVersion() {
	db, err := s.NewDB(s.T().TempDir())
	s.Require().NoError(err)
	defer db.Close()

	// for versions 1-49, set all 10 keys
	for v := uint64(1); v < 50; v++ {
		b, err := db.NewBatch(v)
		s.Require().NoError(err)

		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key%03d", i)
			val := fmt.Sprintf("val%03d-%03d", i, v)

			s.Require().NoError(b.Set(storeKey1, []byte(key), []byte(val)))
		}

		s.Require().NoError(b.Write())
	}

	// for versions 50-100, only update even keys
	for v := uint64(50); v <= 100; v++ {
		b, err := db.NewBatch(v)
		s.Require().NoError(err)

		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				key := fmt.Sprintf("key%03d", i)
				val := fmt.Sprintf("val%03d-%03d", i, v)

				s.Require().NoError(b.Set(storeKey1, []byte(key), []byte(val)))
			}
		}

		s.Require().NoError(b.Write())
	}

	itr, err := db.NewIterator(storeKey1, 69, []byte("key000"), nil)
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
	s.Require().Equal(10, count)
	s.Require().NoError(itr.Error())
}
