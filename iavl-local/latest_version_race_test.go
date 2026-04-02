package iavl

import (
	"fmt"
	"sync"
	"testing"

	dbm "github.com/cosmos/iavl/db"
	"github.com/stretchr/testify/require"
)

func TestLatestVersionRaceScenarios(t *testing.T) {
	t.Run("ConcurrentSaveVersionAndImmutableGet", func(t *testing.T) {
		// ARRANGE
		tree := setupMutableTree(false)
		_, err := tree.Set([]byte("seed-key"), []byte("seed-value"))
		require.NoError(t, err)
		_, baseVersion, err := tree.SaveVersion()
		require.NoError(t, err)

		immutable, err := tree.GetImmutable(baseVersion)
		require.NoError(t, err)

		// ACT
		var wg sync.WaitGroup
		wg.Add(2)

		done := make(chan struct{})
		go func() {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				_, setErr := tree.Set([]byte(fmt.Sprintf("a-key-%d", i)), []byte("v"))
				require.NoError(t, setErr)
				_, _, saveErr := tree.SaveVersion()
				require.NoError(t, saveErr)
			}
			close(done)
		}()

		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_, getErr := immutable.Get([]byte("a-missing-key"))
					require.NoError(t, getErr)
				}
			}
		}()

		wg.Wait()

		// ASSERT
		_, err = immutable.Get([]byte("a-missing-key"))
		require.NoError(t, err)
	})

	t.Run("ConcurrentDeleteOverwriteAndImmutableGet", func(t *testing.T) {
		// ARRANGE
		tree := setupMutableTree(false)
		for i := 0; i < 3; i++ {
			_, err := tree.Set([]byte(fmt.Sprintf("b-seed-%d", i)), []byte("v"))
			require.NoError(t, err)
			_, _, err = tree.SaveVersion()
			require.NoError(t, err)
		}

		immutable, err := tree.GetImmutable(1)
		require.NoError(t, err)

		// ACT
		var wg sync.WaitGroup
		wg.Add(2)

		done := make(chan struct{})
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				_, setErr := tree.Set([]byte(fmt.Sprintf("b-live-%d", i)), []byte("v"))
				require.NoError(t, setErr)
				_, _, saveErr := tree.SaveVersion()
				require.NoError(t, saveErr)

				// Exercise DeleteVersionsFrom() via overwrite flow.
				overwriteErr := tree.LoadVersionForOverwriting(1)
				require.NoError(t, overwriteErr)
			}
			close(done)
		}()

		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_, getErr := immutable.Get([]byte("b-missing-key"))
					require.NoError(t, getErr)
				}
			}
		}()

		wg.Wait()

		// ASSERT
		_, err = immutable.Get([]byte("b-missing-key"))
		require.NoError(t, err)
	})

	t.Run("LazyLatestVersionRefreshAndImmutableGet", func(t *testing.T) {
		// ARRANGE
		db := dbm.NewMemDB()
		writerTree := NewMutableTree(db, 0, false, NewNopLogger())

		for i := 0; i < 4; i++ {
			_, err := writerTree.Set([]byte(fmt.Sprintf("c-seed-%d", i)), []byte("v"))
			require.NoError(t, err)
			_, _, err = writerTree.SaveVersion()
			require.NoError(t, err)
		}

		coldTree := NewMutableTree(db, 0, false, NewNopLogger())
		coldTree.ndb.resetLatestVersion(0)

		immutable, err := coldTree.GetImmutable(1)
		require.NoError(t, err)

		// ACT
		var wg sync.WaitGroup
		wg.Add(2)

		done := make(chan struct{})
		go func() {
			defer wg.Done()
			for i := 0; i < 400; i++ {
				coldTree.ndb.resetLatestVersion(0)
				_, latestErr := coldTree.ndb.getLatestVersion()
				require.NoError(t, latestErr)
			}
			close(done)
		}()

		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
					_, getErr := immutable.Get([]byte("c-missing-key"))
					require.NoError(t, getErr)
				}
			}
		}()

		wg.Wait()

		// ASSERT
		_, err = immutable.Get([]byte("c-missing-key"))
		require.NoError(t, err)
	})
}
