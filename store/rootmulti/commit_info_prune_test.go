package rootmulti

import (
	"fmt"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	pruningtypes "github.com/cosmos/cosmos-sdk/store/v2/pruning/types"
)

func countCommitInfoKeys(t *testing.T, db dbm.DB) int {
	t.Helper()

	iter, err := db.Iterator([]byte("s/0"), []byte("s/:"))
	require.NoError(t, err)
	defer iter.Close()

	count := 0
	for ; iter.Valid(); iter.Next() {
		count++
	}
	require.NoError(t, iter.Error())

	return count
}

func TestPruneCommitInfoDeletesStaleRecords(t *testing.T) {
	db := dbm.NewMemDB()
	ms := newMultiStoreWithMounts(db, pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))
	require.NoError(t, ms.LoadLatestVersion())

	const versions = 50
	for range versions {
		ms.Commit()
	}
	require.Equal(t, versions, countCommitInfoKeys(t, db))

	require.NoError(t, ms.pruneCommitInfo(40))

	require.Equal(t, 11, countCommitInfoKeys(t, db))

	for _, version := range []int64{1, 39} {
		has, err := db.Has([]byte(fmt.Sprintf("s/%d", version)))
		require.NoError(t, err)
		require.False(t, has)
	}
	for _, version := range []int64{40, 50} {
		has, err := db.Has([]byte(fmt.Sprintf("s/%d", version)))
		require.NoError(t, err)
		require.True(t, has)
	}

	has, err := db.Has([]byte(latestVersionKey))
	require.NoError(t, err)
	require.True(t, has)
}

func TestPruneCommitInfoRespectsBatchCapAndIgnoresOtherKeys(t *testing.T) {
	db := dbm.NewMemDB()
	ms := newMultiStoreWithMounts(db, pruningtypes.NewPruningOptions(pruningtypes.PruningNothing))

	const total = commitInfoPruneBatch + 5000
	for version := 1; version <= total; version++ {
		require.NoError(t, db.Set([]byte(fmt.Sprintf("s/%d", version)), []byte{1}))
	}

	decoys := []string{latestVersionKey, earliestVersionKey, "s/k:bank/value", "s/_/value"}
	for _, key := range decoys {
		require.NoError(t, db.Set([]byte(key), []byte{9}))
	}
	require.Equal(t, total, countCommitInfoKeys(t, db))

	require.NoError(t, ms.pruneCommitInfo(int64(total)+1))
	require.Equal(t, total-commitInfoPruneBatch, countCommitInfoKeys(t, db))

	require.NoError(t, ms.pruneCommitInfo(int64(total)+1))
	require.Equal(t, 0, countCommitInfoKeys(t, db))

	for _, key := range decoys {
		has, err := db.Has([]byte(key))
		require.NoError(t, err)
		require.True(t, has)
	}
}
