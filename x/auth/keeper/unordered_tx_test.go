package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
)

func TestManager(t *testing.T) {
	var (
		mgr *keeper.UnorderedTxManager
		ctx sdk.Context
	)
	reset := func() {
		mockStoreKey := storetypes.NewKVStoreKey("test")
		storeService := runtime.NewKVStoreService(mockStoreKey)
		ctx = testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx
		mgr = keeper.NewUnorderedTxManager(storeService)
	}

	type utxSequence struct {
		sender  string
		timeout time.Time
	}
	testCases := map[string]struct {
		addFunc           []utxSequence
		blockTime         time.Time
		expectContains    []utxSequence
		expectNotContains []utxSequence
	}{
		"transactions are not removed when block time is before every utx": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
				{
					"cosmos3",
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(5, 0),
			expectContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
				{
					"cosmos3",
					time.Unix(10, 0),
				},
			},
		},
		"transactions are removed if their timeout is equal to the block time": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
				{
					"cosmos3",
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(10, 10),
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
				{
					"cosmos3",
					time.Unix(10, 0),
				},
			},
		},
		"only some txs are removed": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(15, 0),
				},
				{
					"cosmos3",
					time.Unix(20, 0),
				},
			},
			blockTime: time.Unix(16, 10),
			expectContains: []utxSequence{
				{
					"cosmos3",
					time.Unix(20, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(15, 0),
				},
			},
		},
		"empty state - no transactions to remove": {
			addFunc:           []utxSequence{},
			blockTime:         time.Unix(10, 0),
			expectContains:    []utxSequence{},
			expectNotContains: []utxSequence{},
		},

		"multiple senders with same timestamp": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(10, 1),
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos2",
					time.Unix(10, 0),
				},
			},
		},

		"same sender with multiple timestamps": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos1",
					time.Unix(15, 0),
				},
				{
					"cosmos1",
					time.Unix(20, 0),
				},
			},
			blockTime: time.Unix(16, 0),
			expectContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(20, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos1",
					time.Unix(15, 0),
				},
			},
		},

		"duplicate transaction (same sender and timestamp)": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
				{
					"cosmos1",
					time.Unix(10, 0), // Duplicate entry
				},
			},
			blockTime: time.Unix(5, 0),
			expectContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 0),
				},
			},
		},
		"nanosecond precision boundary test": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 999999998),
				},
				{
					"cosmos2",
					time.Unix(10, 999999999),
				},
				{
					"cosmos3",
					time.Unix(11, 0),
				},
			},
			blockTime: time.Unix(10, 999999999),
			expectContains: []utxSequence{
				{
					"cosmos3",
					time.Unix(11, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(10, 999999998),
				},
				{
					"cosmos2",
					time.Unix(10, 999999999),
				},
			},
		},

		"zero timestamp test": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(0, 0),
				},
			},
			blockTime: time.Unix(1, 0),
			expectNotContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(0, 0),
				},
			},
		},
		"far future timestamp": {
			addFunc: []utxSequence{
				{
					"cosmos1",
					time.Unix(2^30, 0), // Very far in the future
				},
			},
			blockTime: time.Unix(10, 0),
			expectContains: []utxSequence{
				{
					"cosmos1",
					time.Unix(2^30, 0),
				},
			},
		},
	}

	for name, tc := range testCases {
		reset()
		t.Run(name, func(t *testing.T) {
			ctx = ctx.WithBlockTime(tc.blockTime)
			for _, seq := range tc.addFunc {
				err := mgr.Add(ctx, seq.sender, uint64(seq.timeout.UnixNano()))
				t.Logf("added transaction: %d/%s", seq.timeout.UnixNano(), seq.sender)
				require.NoError(t, err)
			}
			t.Logf("removing txs. block_time: %d", tc.blockTime.UnixNano())
			err := mgr.RemoveExpired(ctx)
			require.NoError(t, err)

			for _, seq := range tc.expectNotContains {
				has, err := mgr.Contains(ctx, seq.sender, uint64(seq.timeout.UnixNano()))
				require.NoError(t, err)
				require.False(t, has, "should not contain %s", seq.sender)
			}
			for _, seq := range tc.expectContains {
				has, err := mgr.Contains(ctx, seq.sender, uint64(seq.timeout.UnixNano()))
				require.NoError(t, err)
				require.True(t, has, "expected to contain %d/%s", uint64(seq.timeout.UnixNano()), seq.sender)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(mockStoreKey)
	testCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	mgr := keeper.NewUnorderedTxManager(storeService)

	sender := "cosmos1"
	timeout := time.Unix(100, 0)
	require.NoError(t, mgr.Add(testCtx.Ctx, sender, uint64(timeout.UnixNano())))
	require.NoError(t, mgr.Add(testCtx.Ctx, sender, uint64(timeout.UnixNano()+1)))
	require.NoError(t, mgr.Add(testCtx.Ctx, sender, uint64(timeout.UnixNano()+2)))

	// it should have the tx in state.
	ok, err := mgr.Contains(testCtx.Ctx, sender, uint64(timeout.UnixNano()))
	require.NoError(t, err)
	require.True(t, ok)

	// lets test calling remove with a block time that is before it's expiry.
	blockTime := time.Unix(90, 0)
	testCtx.Ctx = testCtx.Ctx.WithBlockTime(blockTime)

	// it should still have it, since 90 is before 100.
	err = mgr.RemoveExpired(testCtx.Ctx)
	require.NoError(t, err)
	ok, err = mgr.Contains(testCtx.Ctx, sender, uint64(timeout.UnixNano()))
	require.NoError(t, err)
	require.True(t, ok)

	// lets test with the block time after the tx.
	blockTime = time.Unix(101, 0)
	testCtx.Ctx = testCtx.Ctx.WithBlockTime(blockTime)

	// it should be removed since 101 is greater than 100.
	err = mgr.RemoveExpired(testCtx.Ctx)
	require.NoError(t, err)
	ok, err = mgr.Contains(testCtx.Ctx, sender, uint64(timeout.UnixNano()))
	require.NoError(t, err)
	require.False(t, ok)
}
