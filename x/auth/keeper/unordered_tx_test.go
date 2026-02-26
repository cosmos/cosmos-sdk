package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestManager(t *testing.T) {
	var (
		mgr keeper.AccountKeeper
		ctx sdk.Context
	)
	encCfg := moduletestutil.MakeTestEncodingConfig()
	reset := func() {
		mockStoreKey := storetypes.NewKVStoreKey("test")
		storeService := runtime.NewKVStoreService(mockStoreKey)
		ctx = testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx
		mgr = keeper.NewAccountKeeper(
			encCfg.Codec,
			storeService,
			types.ProtoBaseAccount,
			nil,
			authcodec.NewBech32Codec("cosmos"),
			"cosmos",
			types.NewModuleAddress("gov").String(),
			keeper.WithUnorderedTransactions(true),
		)
	}

	type utxSequence struct {
		sender  []byte
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
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos3"),
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(5, 0),
			expectContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos3"),
					time.Unix(10, 0),
				},
			},
		},
		"transactions are removed if their timeout is equal to the block time": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos3"),
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(10, 10),
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos3"),
					time.Unix(10, 0),
				},
			},
		},
		"only some txs are removed": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(15, 0),
				},
				{
					[]byte("cosmos3"),
					time.Unix(20, 0),
				},
			},
			blockTime: time.Unix(16, 10),
			expectContains: []utxSequence{
				{
					[]byte("cosmos3"),
					time.Unix(20, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
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
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
			},
			blockTime: time.Unix(10, 1),
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 0),
				},
			},
		},

		"same sender with multiple timestamps": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos1"),
					time.Unix(15, 0),
				},
				{
					[]byte("cosmos1"),
					time.Unix(20, 0),
				},
			},
			blockTime: time.Unix(16, 0),
			expectContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(20, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 0),
				},
				{
					[]byte("cosmos1"),
					time.Unix(15, 0),
				},
			},
		},
		"nanosecond precision boundary test": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 999999998),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 999999999),
				},
				{
					[]byte("cosmos3"),
					time.Unix(11, 0),
				},
			},
			blockTime: time.Unix(10, 999999999),
			expectContains: []utxSequence{
				{
					[]byte("cosmos3"),
					time.Unix(11, 0),
				},
			},
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(10, 999999998),
				},
				{
					[]byte("cosmos2"),
					time.Unix(10, 999999999),
				},
			},
		},

		"zero timestamp test": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(0, 0),
				},
			},
			blockTime: time.Unix(1, 0),
			expectNotContains: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(0, 0),
				},
			},
		},
		"far future timestamp": {
			addFunc: []utxSequence{
				{
					[]byte("cosmos1"),
					time.Unix(2^30, 0), // Very far in the future
				},
			},
			blockTime: time.Unix(10, 0),
			expectContains: []utxSequence{
				{
					[]byte("cosmos1"),
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
				err := mgr.TryAddUnorderedNonce(ctx, seq.sender, seq.timeout)
				t.Logf("added transaction: %d/%s", seq.timeout.UnixNano(), seq.sender)
				require.NoError(t, err)
			}
			t.Logf("removing txs. block_time: %d", tc.blockTime.UnixNano())
			err := mgr.RemoveExpiredUnorderedNonces(ctx)
			require.NoError(t, err)

			for _, seq := range tc.expectNotContains {
				has, err := mgr.ContainsUnorderedNonce(ctx, seq.sender, seq.timeout)
				require.NoError(t, err)
				require.False(t, has, "should not contain %s", seq.sender)
			}
			for _, seq := range tc.expectContains {
				has, err := mgr.ContainsUnorderedNonce(ctx, seq.sender, seq.timeout)
				require.NoError(t, err)
				require.True(t, has, "expected to contain %d/%s", uint64(seq.timeout.UnixNano()), seq.sender)
			}
		})
	}
}

func TestCannotAddDuplicate(t *testing.T) {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	storeService := runtime.NewKVStoreService(mockStoreKey)
	ctx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx
	mgr := keeper.NewAccountKeeper(
		moduletestutil.MakeTestEncodingConfig().Codec,
		storeService,
		types.ProtoBaseAccount,
		nil,
		authcodec.NewBech32Codec("cosmos"),
		"cosmos",
		types.NewModuleAddress("gov").String(),
		keeper.WithUnorderedTransactions(true),
	)

	addUser := []byte("foo")
	timeout := time.Unix(10, 0)
	err := mgr.TryAddUnorderedNonce(ctx, addUser, timeout)
	require.NoError(t, err)

	err = mgr.TryAddUnorderedNonce(ctx, addUser, timeout)
	require.ErrorContains(t, err, "already used timeout")
}
