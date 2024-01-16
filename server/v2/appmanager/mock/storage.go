package mock

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/pebbledb"
)

func StateCommitment(_ *testing.T) storev2.Committer {
	db := dbm.NewMemDB()
	tree := iavl.NewIavlTree(db, logger{}, iavl.DefaultConfig())

	sc, _ := commitment.NewCommitStore(map[string]commitment.Tree{"": tree}, logger{})
	return sc
}

func StateStorage(t *testing.T) storev2.VersionedDatabase {
	db, err := pebbledb.New(t.TempDir())
	require.NoError(t, err)
	return storage.NewStorageStore(db)
}

type logger struct{}

func (l logger) Info(msg string, keyVals ...any) {}

func (l logger) Error(msg string, keyVals ...any) {}

func (l logger) Debug(msg string, keyVals ...any) {}

func (l logger) Warn(msg string, keyVals ...any) {}

func (l logger) With(keyVals ...any) log.Logger { return l }

func (l logger) Impl() any { return l }
