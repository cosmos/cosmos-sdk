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
			return New(dir)
		},
		EmptyBatchSize: 12,
		SkipTests: []string{
			"TestStorageTestSuite/TestDatabase_Prune",
		},
	}

	suite.Run(t, s)
}
