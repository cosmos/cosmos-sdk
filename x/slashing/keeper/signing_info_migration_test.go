package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// TestMigrateSignedBlocksWindow tests the migration logic when SignedBlocksWindow parameter changes
func TestMigrateSignedBlocksWindow(t *testing.T) {
	suite := new(KeeperTestSuite)
	suite.SetT(t)
	suite.SetupTest()
	require := suite.Require()
	ctx := suite.ctx
	keeper := suite.slashingKeeper

	// Create a test validator
	_, pubKey, addr := testdata.KeyTestPubAddr()
	_ = pubKey // unused but needed for testdata.KeyTestPubAddr()
	consAddr := sdk.ConsAddress(addr)

	// Test Case 1: Window shrinks (10 -> 5), more misses than proportional
	t.Run("shrink window with excess misses", func(t *testing.T) {
		// Setup: 10 block window with 4 misses
		oldWindow := int64(10)
		newWindow := int64(5)

		// Create signing info
		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         8,
			JailedUntil:         time.Time{},
			Tombstoned:          false,
			MissedBlocksCounter: 4,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		// Set missed blocks at positions: 1, 2, 3, 6 (4 total)
		// After truncation to window 5, positions 0-4 exist
		// Positions 1, 2, 3 are kept (3 misses)
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 1, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 2, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 3, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 6, true)) // Will be discarded

		// Run migration
		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Verify results
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)

		// Proportional counter: 4 * 5/10 = 2
		require.Equal(int64(2), updatedInfo.MissedBlocksCounter)

		// IndexOffset should wrap: 8 % 5 = 3
		require.Equal(int64(3), updatedInfo.IndexOffset)

		// Count actual misses in bitmap
		actualMisses := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr, func(index int64, missed bool) (stop bool) {
			if missed {
				actualMisses++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(2), actualMisses, "bitmap should have 2 misses matching counter")
	})

	// Test Case 2: Window shrinks (10 -> 5), fewer misses than proportional
	t.Run("shrink window with insufficient misses", func(t *testing.T) {
		// Clear previous test data
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		oldWindow := int64(10)
		newWindow := int64(5)

		// Setup: 4 misses but all at positions > 5 (will be truncated away)
		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         8,
			JailedUntil:         time.Time{},
			Tombstoned:          false,
			MissedBlocksCounter: 4,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		// Set misses at positions 6, 7, 8, 9 (all will be truncated)
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 6, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 7, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 8, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 9, true))

		// Run migration
		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Verify results
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)

		// Proportional counter: 4 * 5/10 = 2
		require.Equal(int64(2), updatedInfo.MissedBlocksCounter)

		// Count actual misses - should have fabricated 2 misses
		actualMisses := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr, func(index int64, missed bool) (stop bool) {
			if missed {
				actualMisses++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(2), actualMisses, "bitmap should have 2 fabricated misses")
	})

	// Test Case 3: Window increases (5 -> 10)
	t.Run("increase window", func(t *testing.T) {
		// Clear previous test data
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		oldWindow := int64(5)
		newWindow := int64(10)

		// Setup: 3 misses in 5 block window
		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         3,
			JailedUntil:         time.Time{},
			Tombstoned:          false,
			MissedBlocksCounter: 3,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		// Set misses at positions 0, 2, 4
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 0, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 2, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 4, true))

		// Run migration
		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Verify results
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)

		// Proportional counter: 3 * 10/5 = 6
		require.Equal(int64(6), updatedInfo.MissedBlocksCounter)

		// IndexOffset stays: 3 % 10 = 3
		require.Equal(int64(3), updatedInfo.IndexOffset)

		// Count actual misses - should have 6 total (3 original + 3 fabricated)
		actualMisses := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr, func(index int64, missed bool) (stop bool) {
			if missed {
				actualMisses++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(6), actualMisses, "bitmap should have 6 misses after scaling up")
	})

	// Test Case 4: No change in window (edge case)
	t.Run("no window change", func(t *testing.T) {
		// Clear previous test data
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		window := int64(10)

		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         5,
			JailedUntil:         time.Time{},
			Tombstoned:          false,
			MissedBlocksCounter: 3,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 1, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 3, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 7, true))

		// Run migration with same window
		err := keeper.MigrateSignedBlocksWindow(ctx, window, window)
		require.NoError(err)

		// Everything should remain unchanged
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)
		require.Equal(int64(3), updatedInfo.MissedBlocksCounter)
		require.Equal(int64(5), updatedInfo.IndexOffset)
	})

	// Test Case 5: Zero counter edge case
	t.Run("zero counter", func(t *testing.T) {
		// Clear previous test data
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		oldWindow := int64(10)
		newWindow := int64(5)

		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         3,
			JailedUntil:         time.Time{},
			Tombstoned:          false,
			MissedBlocksCounter: 0,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		// Run migration
		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Verify results
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)
		require.Equal(int64(0), updatedInfo.MissedBlocksCounter)

		// No misses should exist
		actualMisses := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr, func(index int64, missed bool) (stop bool) {
			if missed {
				actualMisses++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(0), actualMisses)
	})

	// Test Case 6: Multiple validators migrated together
	t.Run("multiple validators", func(t *testing.T) {
		// Clear previous test data for consAddr
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		// Create a second validator
		_, pubKey2, addr2 := testdata.KeyTestPubAddr()
		_ = pubKey2
		consAddr2 := sdk.ConsAddress(addr2)

		oldWindow := int64(10)
		newWindow := int64(5)

		// Validator 1: 6 misses, all in positions 0-4
		signInfo1 := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         7,
			MissedBlocksCounter: 6,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo1))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 0, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 1, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 2, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 3, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 4, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr, 8, true))

		// Validator 2: 2 misses, scattered
		signInfo2 := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr2.String(),
			StartHeight:         0,
			IndexOffset:         4,
			MissedBlocksCounter: 2,
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr2, signInfo2))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr2, 1, true))
		require.NoError(keeper.SetMissedBlockBitmapValue(ctx, consAddr2, 7, true))

		// Run migration
		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Verify validator 1: proportional counter = 6 * 5/10 = 3
		updated1, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)
		require.Equal(int64(3), updated1.MissedBlocksCounter)
		require.Equal(int64(2), updated1.IndexOffset) // 7 % 5 = 2

		// Verify validator 1 bitmap matches counter
		misses1 := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr, func(index int64, missed bool) (stop bool) {
			if missed {
				misses1++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(3), misses1)

		// Verify validator 2: proportional counter = 2 * 5/10 = 1
		updated2, err := keeper.GetValidatorSigningInfo(ctx, consAddr2)
		require.NoError(err)
		require.Equal(int64(1), updated2.MissedBlocksCounter)
		require.Equal(int64(4), updated2.IndexOffset) // 4 % 5 = 4

		// Verify validator 2 bitmap matches counter
		misses2 := int64(0)
		err = keeper.IterateMissedBlockBitmap(ctx, consAddr2, func(index int64, missed bool) (stop bool) {
			if missed {
				misses2++
			}
			return false
		})
		require.NoError(err)
		require.Equal(int64(1), misses2)

		// Cleanup
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr2))
	})

	// Test Case 7: Counter exceeds old window (corrupted state guard)
	t.Run("corrupted counter exceeding window", func(t *testing.T) {
		require.NoError(keeper.DeleteMissedBlockBitmap(ctx, consAddr))

		oldWindow := int64(10)
		newWindow := int64(5)

		// Corrupted state: counter is larger than window
		signInfo := slashingtypes.ValidatorSigningInfo{
			Address:             consAddr.String(),
			StartHeight:         0,
			IndexOffset:         3,
			MissedBlocksCounter: 15, // corrupted: > oldWindow
		}
		require.NoError(keeper.SetValidatorSigningInfo(ctx, consAddr, signInfo))

		err := keeper.MigrateSignedBlocksWindow(ctx, oldWindow, newWindow)
		require.NoError(err)

		// Should clamp to oldWindow first: min(15, 10) = 10
		// Then proportional: 10 * 5/10 = 5
		updatedInfo, err := keeper.GetValidatorSigningInfo(ctx, consAddr)
		require.NoError(err)
		require.Equal(int64(5), updatedInfo.MissedBlocksCounter)
	})
}
