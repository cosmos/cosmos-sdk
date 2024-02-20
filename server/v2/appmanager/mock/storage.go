package mock

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/pebbledb"
)

func StateCommitment(_ *testing.T) storev2.Committer {
	db := db.NewMemDB()
	tree := iavl.NewIavlTree(db, logger{}, iavl.DefaultConfig())

	sc, _ := commitment.NewCommitStore(map[string]commitment.Tree{"": tree}, db, nil, logger{})
	return sc
}

func StateStorage(t *testing.T) storev2.VersionedDatabase {
	t.Helper()
	db, err := pebbledb.New(t.TempDir())
	require.NoError(t, err)
	return storage.NewStorageStore(db, nil, logger{})
}

type logger struct{}

func (l logger) Info(msg string, keyVals ...any) {}

func (l logger) Error(msg string, keyVals ...any) {}

func (l logger) Debug(msg string, keyVals ...any) {}

func (l logger) Warn(msg string, keyVals ...any) {}

func (l logger) With(keyVals ...any) log.Logger { return l }

func (l logger) Impl() any { return l }
