package rootmulti

import (
	"fmt"
	"sync"
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/store/v2/cache"
	"github.com/cosmos/cosmos-sdk/store/v2/iavl"
	"github.com/cosmos/cosmos-sdk/store/v2/types"
)

// TestResetAndReloadSerializedWithQueryReaders exercises the Restore tail
// (Reset + LoadLatestVersion under interBlockCacheMu) against the query
// paths that can overlap it: CacheMultiStoreWithVersion (CreateQueryContext)
// and Query/GetStoreByName (Unwrap + *iavl.Store type assertion). Unlike a
// full Restore, this does not import a snapshot, so it does not race inside
// the IAVL tree during import.
func TestResetAndReloadSerializedWithQueryReaders(t *testing.T) {
	db := dbm.NewMemDB()
	rs := NewStore(db, log.NewNopLogger())
	key := types.NewKVStoreKey("test")
	rs.MountStoreWithDB(key, types.StoreTypeIAVL, nil)
	rs.SetInterBlockCache(cache.NewCommitKVStoreCacheManager(1000))
	require.NoError(t, rs.LoadLatestVersion())

	testKey := []byte("auth-params")
	testValue := []byte("max_memo_characters=256")
	rs.GetKVStore(key).Set(testKey, testValue)
	rs.Commit()
	version := rs.LastCommitID().Version

	errCh := make(chan error, 1)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	reader := func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				select {
				case errCh <- fmt.Errorf("query reader panic: %v", r):
				default:
				}
			}
		}()

		for {
			select {
			case <-stop:
				return
			default:
			}

			_, err := rs.CacheMultiStoreWithVersion(version)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("CacheMultiStoreWithVersion: %w", err):
				default:
				}
				return
			}

			s := rs.GetStoreByName("test")
			if s != nil {
				if _, ok := s.(*iavl.Store); !ok {
					select {
					case errCh <- fmt.Errorf("GetStoreByName: expected *iavl.Store, got %T", s):
					default:
					}
					return
				}
			}

			_, _ = rs.Query(&types.RequestQuery{Path: "/test/key", Data: testKey})
		}
	}

	const readers = 8
	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go reader()
	}

	// Repeat the Restore tail: exclusive reset + complete reload.
	for i := 0; i < 20; i++ {
		rs.interBlockCacheMu.Lock()
		rs.interBlockCache.Reset()
		err := rs.LoadLatestVersion()
		rs.interBlockCacheMu.Unlock()
		require.NoError(t, err)
	}

	close(stop)
	wg.Wait()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	default:
	}

	require.Equal(t, testValue, rs.GetKVStore(key).Get(testKey))
}
