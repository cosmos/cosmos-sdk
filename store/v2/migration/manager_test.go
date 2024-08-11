package migration

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/store/v2/commitment"
	"cosmossdk.io/store/v2/commitment/iavl"
	dbm "cosmossdk.io/store/v2/db"
	"cosmossdk.io/store/v2/snapshots"
	"cosmossdk.io/store/v2/storage"
	"cosmossdk.io/store/v2/storage/pebbledb"
)

var storeKeys = []string{"store1", "store2"}

func setupMigrationManager(t *testing.T, noCommitStore bool) (*Manager, *commitment.CommitStore) {
	t.Helper()

	db := dbm.NewMemDB()
	multiTrees := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(db, []byte(storeKey))
		multiTrees[storeKey] = iavl.NewIavlTree(prefixDB, coretesting.NewNopLogger(), iavl.DefaultConfig())
	}

	commitStore, err := commitment.NewCommitStore(multiTrees, nil, db, coretesting.NewNopLogger())
	require.NoError(t, err)

	snapshotsStore, err := snapshots.NewStore(t.TempDir())
	require.NoError(t, err)

	snapshotsManager := snapshots.NewManager(snapshotsStore, snapshots.NewSnapshotOptions(1500, 2), commitStore, nil, nil, coretesting.NewNopLogger())

	storageDB, err := pebbledb.New(t.TempDir())
	require.NoError(t, err)
	newStorageStore := storage.NewStorageStore(storageDB, coretesting.NewNopLogger()) // for store/v2

	db1 := dbm.NewMemDB()
	multiTrees1 := make(map[string]commitment.Tree)
	for _, storeKey := range storeKeys {
		prefixDB := dbm.NewPrefixDB(db1, []byte(storeKey))
		multiTrees1[storeKey] = iavl.NewIavlTree(prefixDB, coretesting.NewNopLogger(), iavl.DefaultConfig())
	}

	newCommitStore, err := commitment.NewCommitStore(multiTrees1, nil, db1, coretesting.NewNopLogger()) // for store/v2
	require.NoError(t, err)
	if noCommitStore {
		newCommitStore = nil
	}

	return NewManager(db, snapshotsManager, newStorageStore, newCommitStore, coretesting.NewNopLogger()), commitStore
}

func TestMigrateState(t *testing.T) {
	for _, noCommitStore := range []bool{false, true} {
		t.Run(fmt.Sprintf("Migrate noCommitStore=%v", noCommitStore), func(t *testing.T) {
			m, orgCommitStore := setupMigrationManager(t, noCommitStore)

			// apply changeset
			toVersion := uint64(100)
			keyCount := 10
			for version := uint64(1); version <= toVersion; version++ {
				cs := corestore.NewChangeset()
				for _, storeKey := range storeKeys {
					for i := 0; i < keyCount; i++ {
						cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
					}
				}
				require.NoError(t, orgCommitStore.WriteChangeset(cs))
				_, err := orgCommitStore.Commit(version)
				require.NoError(t, err)
			}

			err := m.Migrate(toVersion - 1)
			require.NoError(t, err)

			// expecting error for conflicting process, since Migrate trigger snapshotter create migration,
			// which start a snapshot process already.
			_, err = m.snapshotsManager.Create(toVersion - 1)
			require.Error(t, err)

			if m.stateCommitment != nil {
				// check the migrated state
				for version := uint64(1); version < toVersion; version++ {
					for _, storeKey := range storeKeys {
						for i := 0; i < keyCount; i++ {
							val, err := m.stateCommitment.Get([]byte(storeKey), toVersion-1, []byte(fmt.Sprintf("key-%d-%d", version, i)))
							require.NoError(t, err)
							require.Equal(t, []byte(fmt.Sprintf("value-%d-%d", version, i)), val)
						}
					}
				}
				// check the latest state
				val, err := m.stateCommitment.Get([]byte("store1"), toVersion-1, []byte("key-100-1"))
				require.NoError(t, err)
				require.Nil(t, val)
				val, err = m.stateCommitment.Get([]byte("store2"), toVersion-1, []byte("key-100-0"))
				require.NoError(t, err)
				require.Nil(t, val)
			}

			// check the storage
			for version := uint64(1); version < toVersion; version++ {
				for _, storeKey := range storeKeys {
					for i := 0; i < keyCount; i++ {
						val, err := m.stateStorage.Get([]byte(storeKey), toVersion-1, []byte(fmt.Sprintf("key-%d-%d", version, i)))
						require.NoError(t, err)
						require.Equal(t, []byte(fmt.Sprintf("value-%d-%d", version, i)), val)
					}
				}
			}
		})
	}
}

func TestStartMigrateState(t *testing.T) {
	for _, noCommitStore := range []bool{false, true} {
		t.Run(fmt.Sprintf("Migrate noCommitStore=%v", noCommitStore), func(t *testing.T) {
			m, orgCommitStore := setupMigrationManager(t, noCommitStore)

			chDone := make(chan struct{})
			chChangeset := make(chan *VersionedChangeset, 1)

			// apply changeset
			toVersion := uint64(10)
			keyCount := 5
			changesets := []corestore.Changeset{}

			for version := uint64(1); version <= toVersion; version++ {
				cs := corestore.NewChangeset()
				for _, storeKey := range storeKeys {
					for i := 0; i < keyCount; i++ {
						cs.Add([]byte(storeKey), []byte(fmt.Sprintf("key-%d-%d", version, i)), []byte(fmt.Sprintf("value-%d-%d", version, i)), false)
					}
				}
				changesets = append(changesets, *cs)
				require.NoError(t, orgCommitStore.WriteChangeset(cs))
				_, err := orgCommitStore.Commit(version)
				require.NoError(t, err)
			}

			// feed changesets to channel
			go func() {
				for version := uint64(1); version <= toVersion; version++ {
					chChangeset <- &VersionedChangeset{
						Version:   version,
						Changeset: &changesets[version-1],
					}
				}
			}()

			// check if migrate process complete
			go func() {
				for {
					migrateVersion := m.GetMigratedVersion()
					if migrateVersion == toVersion-1 {
						break
					}
				}

				chDone <- struct{}{}
			}()

			err := m.Start(toVersion-1, chChangeset, chDone)
			require.NoError(t, err)

			// expecting error for conflicting process, since Migrate trigger snapshotter create migration,
			// which start a snapshot process already.
			_, err = m.snapshotsManager.Create(toVersion - 1)
			require.Error(t, err)

			if m.stateCommitment != nil {
				// check the migrated state
				for version := uint64(1); version < toVersion; version++ {
					for _, storeKey := range storeKeys {
						for i := 0; i < keyCount; i++ {
							val, err := m.stateCommitment.Get([]byte(storeKey), toVersion-1, []byte(fmt.Sprintf("key-%d-%d", version, i)))
							require.NoError(t, err)
							require.Equal(t, []byte(fmt.Sprintf("value-%d-%d", version, i)), val)
						}
					}
				}
				// check the latest state
				val, err := m.stateCommitment.Get([]byte("store1"), toVersion-1, []byte("key-100-1"))
				require.NoError(t, err)
				require.Nil(t, val)
				val, err = m.stateCommitment.Get([]byte("store2"), toVersion-1, []byte("key-100-0"))
				require.NoError(t, err)
				require.Nil(t, val)
			}

			// check the storage
			for version := uint64(1); version < toVersion; version++ {
				for _, storeKey := range storeKeys {
					for i := 0; i < keyCount; i++ {
						val, err := m.stateStorage.Get([]byte(storeKey), toVersion-1, []byte(fmt.Sprintf("key-%d-%d", version, i)))
						require.NoError(t, err)
						require.Equal(t, []byte(fmt.Sprintf("value-%d-%d", version, i)), val)
					}
				}
			}

			// check if migration db write change set to storage
			for version := uint64(1); version < toVersion; version++ {
				buf := make([]byte, 8)
				binary.BigEndian.PutUint64(buf, version)
				csKey := []byte(fmt.Sprintf(migrateChangesetKeyFmt, buf))
				csVal, err := m.db.Get(csKey)
				require.NoError(t, err)
				require.NotEmpty(t, csVal)
			}
		})
	}
}
