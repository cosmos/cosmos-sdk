package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestCacheEntry_BasicOperations(t *testing.T) {
	loadFunc := func(ctx context.Context) (map[string][]string, error) {
		return map[string][]string{
			"key1": {"val1", "val2"},
			"key2": {"val3"},
		}, nil
	}
	cache := NewCacheEntry(100, loadFunc)

	// Initially empty (dirty=true, data=nil)
	data := cache.get()
	require.Empty(t, data)

	// Set some data
	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2", "val3"})

	retrieved := cache.get()
	require.Len(t, retrieved, 2)
	require.Equal(t, []string{"val1"}, retrieved["key1"])
	require.Equal(t, []string{"val2", "val3"}, retrieved["key2"])

	// Verify it's a copy (modifying returned data shouldn't affect cache)
	retrieved["key1"][0] = "modified"
	retrievedAgain := cache.get()
	require.Equal(t, "val1", retrievedAgain["key1"][0])
}

func TestCacheEntry_GetEntry(t *testing.T) {
	cache := NewCacheEntry[string, []string](100, nil)

	cache.setEntry("key1", []string{"val1", "val2"})
	cache.setEntry("key2", []string{"val3"})

	// Get specific entry
	entry := cache.getEntry("key1")
	require.Equal(t, []string{"val1", "val2"}, entry)

	// Get non-existent entry
	nonExistent := cache.getEntry("key3")
	require.Empty(t, nonExistent)

	// Verify it's a copy
	entry[0] = "modified"
	entryAgain := cache.getEntry("key1")
	require.Equal(t, "val1", entryAgain[0])
}

func TestCacheEntry_SetEntry(t *testing.T) {
	cache := NewCacheEntry[string, []string](100, nil)

	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2", "val3"})

	data := cache.get()
	require.Len(t, data, 2)
	require.Equal(t, []string{"val1"}, data["key1"])
	require.Equal(t, []string{"val2", "val3"}, data["key2"])
}

func TestCacheEntry_DeleteEntry(t *testing.T) {
	cache := NewCacheEntry[string, []string](100, nil)

	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2"})

	data := cache.get()
	require.Len(t, data, 2)

	cache.deleteEntry("key1")
	data = cache.get()
	require.Len(t, data, 1)
	require.NotContains(t, data, "key1")
	require.Contains(t, data, "key2")
}

func TestCacheEntry_UnlimitedSize(t *testing.T) {
	cache := NewCacheEntry[string, []string](0, nil)

	// With max=0, cache is unlimited (can store anything)
	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2"})
	cache.setEntry("key3", []string{"val3"})

	data := cache.get()
	require.Len(t, data, 3)
	require.False(t, cache.full.Load(), "unlimited cache should never be full")

	// Add many more entries to verify it's truly unlimited
	for i := 0; i < 1000; i++ {
		cache.setEntry(fmt.Sprintf("key%d", i+10), []string{fmt.Sprintf("val%d", i)})
	}

	data = cache.get()
	require.GreaterOrEqual(t, len(data), 1000)
	require.False(t, cache.full.Load(), "unlimited cache should never be full")
}

func TestCacheEntry_MaxSizeExceeded(t *testing.T) {
	cache := NewCacheEntry[string, []string](3, nil)

	// Add entries up to limit
	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2"})
	cache.setEntry("key3", []string{"val3"})

	require.True(t, cache.full.Load())
	require.Len(t, cache.get(), 3)

	// Try to add one more - should set full flag
	cache.setEntry("key4", []string{"val4"})

	require.True(t, cache.full.Load(), "cache should be marked as full")
	// key4 was not added because cache is now full
	data := cache.get()
	require.Len(t, data, 3)
	require.NotContains(t, data, "key4")
}

func TestCacheEntry_FullFlagPreventsWrites(t *testing.T) {
	cache := NewCacheEntry[string, []string](2, nil)

	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2"})
	require.True(t, cache.full.Load())

	// Try to add more - should be ignored
	cache.setEntry("key4", []string{"val4"})
	data := cache.get()
	require.Len(t, data, 2)
	require.NotContains(t, data, "key4")

	// Try to edit existing entry - should be ignored
	cache.setEntry("key2", []string{"val2", "val3"})
	data = cache.get()
	require.Len(t, data, 2)
}

func TestCacheEntry_DeleteClearsFull(t *testing.T) {
	cache := NewCacheEntry[string, []string](2, nil)

	// Fill to capacity
	cache.setEntry("key1", []string{"val1"})
	cache.setEntry("key2", []string{"val2"})
	cache.setEntry("key3", []string{"val3"}) // Triggers full
	require.True(t, cache.full.Load())

	// Delete one entry - should clear full flag
	cache.deleteEntry("key1")
	require.False(t, cache.full.Load(), "deleting should clear full flag")

	// Now we should be able to add again
	cache.setEntry("key4", []string{"val4"})
	data := cache.get()
	require.Len(t, data, 2)
	require.Contains(t, data, "key4")
}

// Test ValidatorsQueueCache
func TestValidatorsQueueCache_Initialization(t *testing.T) {
	validatorsLoader := func(ctx context.Context) (map[string][]string, error) {
		return map[string][]string{
			"time1": {"val1", "val2"},
			"time2": {"val3"},
		}, nil
	}
	delegationsLoader := func(ctx context.Context) (map[string][]types.DVPair, error) {
		return map[string][]types.DVPair{
			"time1": {{DelegatorAddress: "del1", ValidatorAddress: "val1"}},
		}, nil
	}
	redelegationsLoader := func(ctx context.Context) (map[string][]types.DVVTriplet, error) {
		return map[string][]types.DVVTriplet{
			"time1": {{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"}},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		validatorsLoader,
		delegationsLoader,
		redelegationsLoader,
	)

	require.NotNil(t, cache)
	require.NotNil(t, cache.unbondingValidatorsQueue)
	require.NotNil(t, cache.unbondingDelegationsQueue)
	require.NotNil(t, cache.redelegationsQueue)
}

func TestValidatorsQueueCache_LoadFromStore(t *testing.T) {
	ctx := context.Background()

	validatorsLoader := func(ctx context.Context) (map[string][]string, error) {
		return map[string][]string{
			"time1": {"val1", "val2"},
			"time2": {"val3"},
		}, nil
	}

	delegationsLoader := func(ctx context.Context) (map[string][]types.DVPair, error) {
		return map[string][]types.DVPair{
			"time1": {{DelegatorAddress: "del1", ValidatorAddress: "val1"}},
		}, nil
	}
	redelegationsLoader := func(ctx context.Context) (map[string][]types.DVVTriplet, error) {
		return map[string][]types.DVVTriplet{
			"time1": {{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"}},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		validatorsLoader,
		delegationsLoader,
		redelegationsLoader,
	)

	// Initially dirty, should load from store
	unbondingValidators, err := cache.GetUnbondingValidatorsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, unbondingValidators, 2)
	require.Equal(t, []string{"val1", "val2"}, unbondingValidators["time1"])
	require.Equal(t, []string{"val3"}, unbondingValidators["time2"])

	unbondingDelegations, err := cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, unbondingDelegations, 1)
	require.Equal(t, []types.DVPair{{DelegatorAddress: "del1", ValidatorAddress: "val1"}}, unbondingDelegations["time1"])

	redelgations, err := cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, redelgations, 1)
	require.Equal(t, []types.DVVTriplet{{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"}}, redelgations["time1"])

	// Cache should no longer be dirty
	require.False(t, cache.unbondingValidatorsQueue.dirty.Load())
	require.False(t, cache.unbondingDelegationsQueue.dirty.Load())
	require.False(t, cache.redelegationsQueue.dirty.Load())
}

func TestValidatorsQueueCache_DirtyReinitialization(t *testing.T) {
	ctx := context.Background()

	// Test unbonding validators
	valCallCount := 0
	validatorsLoader := func(ctx context.Context) (map[string][]string, error) {
		valCallCount++
		if valCallCount == 1 {
			return map[string][]string{
				"time1": {"val1"},
			}, nil
		}
		return map[string][]string{
			"time2": {"val2"},
		}, nil
	}

	// Test unbonding delegations
	delCallCount := 0
	delegationsLoader := func(ctx context.Context) (map[string][]types.DVPair, error) {
		delCallCount++
		if delCallCount == 1 {
			return map[string][]types.DVPair{
				"time1": {{DelegatorAddress: "del1", ValidatorAddress: "val1"}},
			}, nil
		}
		return map[string][]types.DVPair{
			"time2": {{DelegatorAddress: "del2", ValidatorAddress: "val2"}},
		}, nil
	}

	// Test redelegations
	redCallCount := 0
	redelegationsLoader := func(ctx context.Context) (map[string][]types.DVVTriplet, error) {
		redCallCount++
		if redCallCount == 1 {
			return map[string][]types.DVVTriplet{
				"time1": {{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"}},
			}, nil
		}
		return map[string][]types.DVVTriplet{
			"time2": {{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"}},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		validatorsLoader,
		delegationsLoader,
		redelegationsLoader,
	)

	// Test unbonding validators queue
	valData, err := cache.GetUnbondingValidatorsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, valData, 1)
	require.Contains(t, valData, "time1")
	require.Equal(t, 1, valCallCount)

	cache.unbondingValidatorsQueue.dirty.Store(true)
	valData, err = cache.GetUnbondingValidatorsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, valData, 1)
	require.Contains(t, valData, "time2")
	require.NotContains(t, valData, "time1", "old data should be cleared")
	require.Equal(t, 2, valCallCount)

	// Test unbonding delegations queue
	delData, err := cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, delData, 1)
	require.Contains(t, delData, "time1")
	require.Equal(t, 1, delCallCount)

	cache.unbondingDelegationsQueue.dirty.Store(true)
	delData, err = cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, delData, 1)
	require.Contains(t, delData, "time2")
	require.NotContains(t, delData, "time1", "old data should be cleared")
	require.Equal(t, 2, delCallCount)

	// Test redelegations queue
	redData, err := cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, redData, 1)
	require.Contains(t, redData, "time1")
	require.Equal(t, 1, redCallCount)

	cache.redelegationsQueue.dirty.Store(true)
	redData, err = cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, redData, 1)
	require.Contains(t, redData, "time2")
	require.NotContains(t, redData, "time1", "old data should be cleared")
	require.Equal(t, 2, redCallCount)
}

func TestValidatorsQueueCache_FullPreventsLoad(t *testing.T) {
	ctx := context.Background()

	validatorsLoader := func(ctx context.Context) (map[string][]string, error) {
		// Return too much data (4 keys > max 3)
		return map[string][]string{
			"time1": {"val1"},
			"time2": {"val2"},
			"time3": {"val3"},
			"time4": {"val4"},
		}, nil
	}

	delegationsLoader := func(ctx context.Context) (map[string][]types.DVPair, error) {
		// Return too much data (4 keys > max 3)
		return map[string][]types.DVPair{
			"time1": {{DelegatorAddress: "del1", ValidatorAddress: "val1"}},
			"time2": {{DelegatorAddress: "del2", ValidatorAddress: "val2"}},
			"time3": {{DelegatorAddress: "del3", ValidatorAddress: "val3"}},
			"time4": {{DelegatorAddress: "del4", ValidatorAddress: "val4"}},
		}, nil
	}

	redelegationsLoader := func(ctx context.Context) (map[string][]types.DVVTriplet, error) {
		// Return too much data (4 keys > max 3)
		return map[string][]types.DVVTriplet{
			"time1": {{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"}},
			"time2": {{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"}},
			"time3": {{DelegatorAddress: "del3", ValidatorSrcAddress: "val3", ValidatorDstAddress: "val4"}},
			"time4": {{DelegatorAddress: "del4", ValidatorSrcAddress: "val4", ValidatorDstAddress: "val5"}},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		3,
		logger,
		validatorsLoader,
		delegationsLoader,
		redelegationsLoader,
	)

	// Try to load unbonding validators - should fail due to exceeding max
	_, err := cache.GetUnbondingValidatorsQueue(ctx)
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.unbondingValidatorsQueue.full.Load())

	// Try to load unbonding delegations - should fail due to exceeding max
	_, err = cache.GetUnbondingDelegationsQueue(ctx)
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.unbondingDelegationsQueue.full.Load())

	// Try to load redelegations - should fail due to exceeding max
	_, err = cache.GetRedelegationsQueue(ctx)
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.redelegationsQueue.full.Load())
}

func TestValidatorsQueueCache_GetEntry(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flags to avoid loading
	cache.unbondingValidatorsQueue.dirty.Store(false)
	cache.unbondingDelegationsQueue.dirty.Store(false)
	cache.redelegationsQueue.dirty.Store(false)

	// Test unbonding validators queue entry
	endTime := time.Now().UTC()
	endHeight := int64(1000)
	valKey := types.GetCacheValidatorQueueKey(endTime, endHeight)

	cache.SetUnbondingValidatorQueueEntry(ctx, valKey, []string{"val1", "val2"})
	valEntry, err := cache.GetUnbondingValidatorsQueueEntry(ctx, endTime, endHeight)
	require.NoError(t, err)
	require.Equal(t, []string{"val1", "val2"}, valEntry)

	// Test unbonding delegations queue entry
	delKey := sdk.FormatTimeString(endTime)
	delPairs := []types.DVPair{
		{DelegatorAddress: "del1", ValidatorAddress: "val1"},
		{DelegatorAddress: "del2", ValidatorAddress: "val2"},
	}
	cache.SetUnbondingDelegationsQueueEntry(ctx, delKey, delPairs)
	delEntry, err := cache.GetUnbondingDelegationsQueueEntry(ctx, endTime)
	require.NoError(t, err)
	require.Equal(t, delPairs, delEntry)

	// Test redelegations queue entry
	redKey := sdk.FormatTimeString(endTime)
	redTriplets := []types.DVVTriplet{
		{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
		{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"},
	}
	cache.SetRedelegationsQueueEntry(ctx, redKey, redTriplets)
	redEntry, err := cache.GetRedelegationsQueueEntry(ctx, endTime)
	require.NoError(t, err)
	require.Equal(t, redTriplets, redEntry)
}

func TestValidatorsQueueCache_SetAndDelete(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flags to avoid loading
	cache.unbondingValidatorsQueue.dirty.Store(false)
	cache.unbondingDelegationsQueue.dirty.Store(false)
	cache.redelegationsQueue.dirty.Store(false)

	// Test unbonding validators queue
	err := cache.SetUnbondingValidatorQueueEntry(ctx, "key1", []string{"val1"})
	require.NoError(t, err)
	err = cache.SetUnbondingValidatorQueueEntry(ctx, "key2", []string{"val2"})
	require.NoError(t, err)

	valData, err := cache.GetUnbondingValidatorsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, valData, 2)

	cache.DeleteUnbondingValidatorQueueEntry("key1")
	valData, err = cache.GetUnbondingValidatorsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, valData, 1)
	require.NotContains(t, valData, "key1")

	// Test unbonding delegations queue
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time1", []types.DVPair{
		{DelegatorAddress: "del1", ValidatorAddress: "val1"},
	})
	require.NoError(t, err)
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time2", []types.DVPair{
		{DelegatorAddress: "del2", ValidatorAddress: "val2"},
	})
	require.NoError(t, err)

	delData, err := cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, delData, 2)

	cache.DeleteUnbondingDelegationQueueEntry("time1")
	delData, err = cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, delData, 1)
	require.NotContains(t, delData, "time1")

	// Test redelegations queue
	err = cache.SetRedelegationsQueueEntry(ctx, "time1", []types.DVVTriplet{
		{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
	})
	require.NoError(t, err)
	err = cache.SetRedelegationsQueueEntry(ctx, "time2", []types.DVVTriplet{
		{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"},
	})
	require.NoError(t, err)

	redData, err := cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, redData, 2)

	cache.DeleteRedelegationsQueueEntry("time1")
	redData, err = cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, redData, 1)
	require.NotContains(t, redData, "time1")
}

func TestValidatorsQueueCache_FullMarkedDirty(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		2,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flags
	cache.unbondingValidatorsQueue.dirty.Store(false)
	cache.unbondingDelegationsQueue.dirty.Store(false)
	cache.redelegationsQueue.dirty.Store(false)

	// Test unbonding validators queue
	err := cache.SetUnbondingValidatorQueueEntry(ctx, "key1", []string{"val1"})
	require.NoError(t, err)
	err = cache.SetUnbondingValidatorQueueEntry(ctx, "key2", []string{"val2"})
	require.NoError(t, err)

	// Try to add one more - should mark as full and dirty
	err = cache.SetUnbondingValidatorQueueEntry(ctx, "key3", []string{"val3"})
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.unbondingValidatorsQueue.full.Load())
	require.True(t, cache.unbondingValidatorsQueue.dirty.Load())

	// Test unbonding delegations queue
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time1", []types.DVPair{
		{DelegatorAddress: "del1", ValidatorAddress: "val1"},
	})
	require.NoError(t, err)
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time2", []types.DVPair{
		{DelegatorAddress: "del2", ValidatorAddress: "val2"},
	})
	require.NoError(t, err)

	// Try to add one more - should mark as full and dirty
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time3", []types.DVPair{
		{DelegatorAddress: "del3", ValidatorAddress: "val3"},
	})
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.unbondingDelegationsQueue.full.Load())
	require.True(t, cache.unbondingDelegationsQueue.dirty.Load())

	// Test redelegations queue
	err = cache.SetRedelegationsQueueEntry(ctx, "time1", []types.DVVTriplet{
		{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
	})
	require.NoError(t, err)
	err = cache.SetRedelegationsQueueEntry(ctx, "time2", []types.DVVTriplet{
		{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"},
	})
	require.NoError(t, err)

	// Try to add one more - should mark as full and dirty
	err = cache.SetRedelegationsQueueEntry(ctx, "time3", []types.DVVTriplet{
		{DelegatorAddress: "del3", ValidatorSrcAddress: "val3", ValidatorDstAddress: "val4"},
	})
	require.Error(t, err)
	require.Equal(t, types.ErrCacheMaxSizeReached, err)
	require.True(t, cache.redelegationsQueue.full.Load())
	require.True(t, cache.redelegationsQueue.dirty.Load())
}

func TestValidatorsQueueCache_UnbondingDelegations(t *testing.T) {
	ctx := context.Background()

	delegationsLoader := func(ctx context.Context) (map[string][]types.DVPair, error) {
		return map[string][]types.DVPair{
			"time1": {
				{DelegatorAddress: "del1", ValidatorAddress: "val1"},
				{DelegatorAddress: "del2", ValidatorAddress: "val2"},
			},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		delegationsLoader,
		nil,
	)

	// Load from store
	data, err := cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.Len(t, data["time1"], 2)

	// Set individual entry
	err = cache.SetUnbondingDelegationsQueueEntry(ctx, "time2", []types.DVPair{
		{DelegatorAddress: "del3", ValidatorAddress: "val3"},
	})
	require.NoError(t, err)

	data, err = cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 2)

	// Delete entry
	cache.DeleteUnbondingDelegationQueueEntry("time1")
	data, err = cache.GetUnbondingDelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.NotContains(t, data, "time1")
}

func TestValidatorsQueueCache_UnbondingDelegationsEntry(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flag
	cache.unbondingDelegationsQueue.dirty.Store(false)

	endTime := time.Now().UTC()
	key := sdk.FormatTimeString(endTime)

	// Set entry
	pairs := []types.DVPair{
		{DelegatorAddress: "del1", ValidatorAddress: "val1"},
	}
	err := cache.SetUnbondingDelegationsQueueEntry(ctx, key, pairs)
	require.NoError(t, err)

	// Get specific entry
	entry, err := cache.GetUnbondingDelegationsQueueEntry(ctx, endTime)
	require.NoError(t, err)
	require.Equal(t, pairs, entry)
}

func TestValidatorsQueueCache_Redelegations(t *testing.T) {
	ctx := context.Background()

	redelegationsLoader := func(ctx context.Context) (map[string][]types.DVVTriplet, error) {
		return map[string][]types.DVVTriplet{
			"time1": {
				{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
			},
		}, nil
	}

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		nil,
		redelegationsLoader,
	)

	// Load from store
	data, err := cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)

	// Set individual entry
	err = cache.SetRedelegationsQueueEntry(ctx, "time2", []types.DVVTriplet{
		{DelegatorAddress: "del2", ValidatorSrcAddress: "val2", ValidatorDstAddress: "val3"},
	})
	require.NoError(t, err)

	data, err = cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 2)

	// Delete entry
	cache.DeleteRedelegationsQueueEntry("time1")
	data, err = cache.GetRedelegationsQueue(ctx)
	require.NoError(t, err)
	require.Len(t, data, 1)
	require.NotContains(t, data, "time1")
}

func TestValidatorsQueueCache_RedelegationsEntry(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		100,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flag
	cache.redelegationsQueue.dirty.Store(false)

	endTime := time.Now().UTC()
	key := sdk.FormatTimeString(endTime)

	// Set entry
	triplets := []types.DVVTriplet{
		{DelegatorAddress: "del1", ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
	}
	err := cache.SetRedelegationsQueueEntry(ctx, key, triplets)
	require.NoError(t, err)

	// Get specific entry
	entry, err := cache.GetRedelegationsQueueEntry(ctx, endTime)
	require.NoError(t, err)
	require.Equal(t, triplets, entry)
}

// Concurrent operations tests
func TestCacheEntry_ConcurrentReads(t *testing.T) {
	cache := NewCacheEntry[string, []string](1000, nil)
	cache.dirty.Store(false) // Skip loading

	for i := 0; i < 100; i++ {
		cache.setEntry(fmt.Sprintf("key%d", i), []string{fmt.Sprintf("val%d", i)})
	}

	var wg sync.WaitGroup
	numReaders := 50
	readsPerReader := 100

	wg.Add(numReaders)
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerReader; j++ {
				data := cache.get()
				require.NotEmpty(t, data)
			}
		}()
	}

	wg.Wait()
}

func TestCacheEntry_ConcurrentWrites(t *testing.T) {
	cache := NewCacheEntry[string, []string](10000, nil)

	var wg sync.WaitGroup
	numWriters := 50
	writesPerWriter := 100

	wg.Add(numWriters)
	for i := 0; i < numWriters; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.setEntry(key, []string{fmt.Sprintf("val_%d", id)})
			}
		}(i)
	}

	wg.Wait()

	// Verify data integrity
	data := cache.get()
	require.NotEmpty(t, data)
	// Each writer creates unique keys, so we expect exactly numWriters * writesPerWriter entries
	expectedKeys := numWriters * writesPerWriter
	require.Equal(t, expectedKeys, len(data), "cache should contain exactly %d keys", expectedKeys)
}

func TestCacheEntry_ConcurrentReadWrite(t *testing.T) {
	cache := NewCacheEntry[string, []string](10000, nil)
	cache.dirty.Store(false)

	for i := 0; i < 100; i++ {
		cache.setEntry(fmt.Sprintf("key%d", i), []string{fmt.Sprintf("val%d", i)})
	}

	var wg sync.WaitGroup
	numRoutines := 50
	operationsPerRoutine := 100

	// Readers
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				_ = cache.get()
			}
		}()
	}

	// Writers
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("new_key_%d_%d", id, j)
				cache.setEntry(key, []string{fmt.Sprintf("val_%d", id)})
			}
		}(i)
	}

	wg.Wait()

	// Should complete without race conditions
	data := cache.get()
	require.NotEmpty(t, data)
}

func TestCacheEntry_ConcurrentSetAndDelete(t *testing.T) {
	cache := NewCacheEntry[string, []string](10000, nil)

	var wg sync.WaitGroup
	numRoutines := 30
	operationsPerRoutine := 100

	// Writers
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("key_%d", id%10) // Reuse some keys
				cache.setEntry(key, []string{fmt.Sprintf("val_%d_%d", id, j)})
			}
		}(i)
	}

	// Deleters
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("key_%d", id%10)
				cache.deleteEntry(key)
			}
		}(i)
	}

	wg.Wait()

	// Should complete without panic
	_ = cache.get()
}

func TestValidatorsQueueCache_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()

	logger := func(ctx context.Context) log.Logger {
		return log.NewNopLogger()
	}

	cache := NewValidatorsQueueCache(
		10000,
		logger,
		nil,
		nil,
		nil,
	)

	// Clear dirty flags
	cache.unbondingValidatorsQueue.dirty.Store(false)
	cache.unbondingDelegationsQueue.dirty.Store(false)
	cache.redelegationsQueue.dirty.Store(false)

	var wg sync.WaitGroup
	numRoutines := 30
	operationsPerRoutine := 100

	// Concurrent unbonding validators operations
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("v_%d_%d", id, j)
				cache.SetUnbondingValidatorQueueEntry(ctx, key, []string{fmt.Sprintf("addr_%d", id)})
			}
		}(i)
	}

	// Concurrent unbonding delegations operations
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("d_%d_%d", id, j)
				cache.SetUnbondingDelegationsQueueEntry(ctx, key, []types.DVPair{
					{DelegatorAddress: fmt.Sprintf("del_%d", id), ValidatorAddress: "val"},
				})
			}
		}(i)
	}

	// Concurrent redelegations operations
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				key := fmt.Sprintf("r_%d_%d", id, j)
				cache.SetRedelegationsQueueEntry(ctx, key, []types.DVVTriplet{
					{DelegatorAddress: fmt.Sprintf("del_%d", id), ValidatorSrcAddress: "val1", ValidatorDstAddress: "val2"},
				})
			}
		}(i)
	}

	// Concurrent readers
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerRoutine; j++ {
				_, _ = cache.GetUnbondingValidatorsQueue(ctx)
				_, _ = cache.GetUnbondingDelegationsQueue(ctx)
				_, _ = cache.GetRedelegationsQueue(ctx)
			}
		}()
	}

	wg.Wait()

	// Verify all caches have data
	vData, _ := cache.GetUnbondingValidatorsQueue(ctx)
	dData, _ := cache.GetUnbondingDelegationsQueue(ctx)
	rData, _ := cache.GetRedelegationsQueue(ctx)

	require.NotEmpty(t, vData)
	require.NotEmpty(t, dData)
	require.NotEmpty(t, rData)
}
