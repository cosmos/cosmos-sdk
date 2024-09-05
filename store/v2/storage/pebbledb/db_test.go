package pebbledb

import (
	"testing"

	"github.com/stretchr/testify/suite"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/storage"
)

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (*storage.StorageStore, error) {
			db, err := New(dir)
			if err == nil && db != nil {
				// We set sync=false just to speed up CI tests. Operators should take
				// careful consideration when setting this value in production environments.
				db.SetSync(false)
			}

			return storage.NewStorageStore(db, coretesting.NewNopLogger()), err
		},
		EmptyBatchSize: 12,
	}

	suite.Run(t, s)
}
