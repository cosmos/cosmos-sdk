//go:build rocksdb
// +build rocksdb

package db

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestRocksDBSuite(t *testing.T) {
	db, err := NewRocksDB("test", t.TempDir())
	require.NoError(t, err)

	suite.Run(t, &DBTestSuite{
		db: db,
	})
}
