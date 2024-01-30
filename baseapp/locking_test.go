package baseapp_test

import (
	"encoding/binary"
	"fmt"
	"sync"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestLockAndCacheContextDecorator(t *testing.T) {
	tests := map[string]struct {
		err error
	}{
		"AnteHandle succeeds": {},
		"AnteHandle returns error": {
			err: fmt.Errorf("Fake test failure"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			key := []byte("key")
			db := dbm.NewMemDB()

			storeKey := storetypes.NewKVStoreKey("test")
			cms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
			cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
			ctx := sdk.NewContext(cms, cmtproto.Header{}, true, log.NewNopLogger())
			cms.LoadLatestVersion()

			cms.GetKVStore(storeKey).Set(key, encode(0))
			ctx.WithTxBytes([]byte("fake tx bytes"))

			l := baseapp.NewLockAndCacheContextAnteDecorator()

			_, err := l.AnteHandle(
				ctx,
				nil,
				false,
				func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
					store := ctx.MultiStore().GetKVStore(storeKey)
					store.Set(key, encode(decode(store.Get(key))+1))
					return ctx, tc.err
				},
			)

			actual := int(decode(cms.GetKVStore(storeKey).Get(key)))

			require.Equal(t, tc.err, err)
			if tc.err != nil {
				require.Equal(t, 0, actual)
			} else {
				require.Equal(t, 1, actual)
			}
		})
	}
}

func TestLockAndCacheContextDecorator_Concurrency(t *testing.T) {
	const numThreads = 999
	key := []byte("key")
	db := dbm.NewMemDB()

	errResp := fmt.Errorf("Fake test failure")

	storeKey := storetypes.NewKVStoreKey("test")
	cms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	cms.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	ctx := sdk.NewContext(cms, cmtproto.Header{}, true, log.NewNopLogger())
	cms.LoadLatestVersion()

	cms.GetKVStore(storeKey).Set(key, encode(0))
	ctx.WithTxBytes([]byte("fake tx bytes"))

	l := baseapp.NewLockAndCacheContextAnteDecorator()

	wg := sync.WaitGroup{}
	wg.Add(numThreads)
	for i := 0; i < numThreads; i++ {
		ii := i
		go func() {
			_, err := l.AnteHandle(
				ctx,
				nil,
				false,
				func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
					store := ctx.MultiStore().GetKVStore(storeKey)
					store.Set(key, encode(decode(store.Get(key))+1))
					if ii%3 == 0 {
						return ctx, nil
					}
					return ctx, errResp
				},
			)
			wg.Done()

			if ii%3 == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		}()
	}
	wg.Wait()

	// We run 1/3 the goroutines with failures and 2/3 with successes. Since each run is expected to execute
	// independently of each other after the locking decorator we know that exactly 2/3's must succeed.

	actual := int(decode(cms.GetKVStore(storeKey).Get(key)))
	require.Equal(t, numThreads/3, actual)
}

func encode(v uint32) []byte {
	rval := make([]byte, 4)
	binary.LittleEndian.PutUint32(rval, v)
	return rval
}

func decode(v []byte) uint32 {
	return binary.LittleEndian.Uint32(v)
}
