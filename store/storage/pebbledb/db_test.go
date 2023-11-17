package pebbledb

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/storage"
)

func TestStorageTestSuite(t *testing.T) {
	s := &storage.StorageTestSuite{
		NewDB: func(dir string) (store.VersionedDatabase, error) {
			db, err := New(dir)
			if err == nil && db != nil {
				// we set sync=false just to speed up CI tests
				db.SetSync(false)
			}

			return db, err
		},
		EmptyBatchSize: 12,
	}

	suite.Run(t, s)
}
