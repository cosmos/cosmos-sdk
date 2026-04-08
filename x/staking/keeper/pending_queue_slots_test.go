package keeper_test

import (
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestGetValidatorQueuePendingSlots_NoEntries() {
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Nil(slots)
}

func (s *KeeperTestSuite) TestGetValidatorQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime, testHeight)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0].Time)
	s.Require().Equal(testHeight, slots[0].Height)
}

func (s *KeeperTestSuite) TestGetValidatorQueuePendingSlots_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}
	testHeights := []int64{100, 200, 300}

	for i, t := range testTimes {
		err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, t, testHeights[i])
		s.Require().NoError(err)
	}

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)
	// Should be sorted by time, then height
	for i, slot := range slots {
		s.Require().Equal(testTimes[i], slot.Time)
		s.Require().Equal(testHeights[i], slot.Height)
	}
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_EmptySlice() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	// Add a slot first
	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime, testHeight)
	s.Require().NoError(err)

	// Set empty slice (should delete)
	err = s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, []stakingtypes.TimeHeightQueueSlot{})
	s.Require().NoError(err)

	// Verify it's deleted
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_DuplicateEntries() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	// Set with duplicates
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: testTime, Height: testHeight},
		{Time: testTime, Height: testHeight}, // duplicate
		{Time: testTime, Height: testHeight}, // duplicate
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	// Verify only one entry persisted
	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0].Time)
	s.Require().Equal(testHeight, retrievedSlots[0].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: testTime, Height: testHeight},
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0].Time)
	s.Require().Equal(testHeight, retrievedSlots[0].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_MultipleEntries() {
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Height: 300},
		{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Height: 100},
		{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Height: 200},
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 3)
	// Should be sorted by time, then height
	s.Require().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), retrievedSlots[0].Time)
	s.Require().Equal(int64(100), retrievedSlots[0].Height)
	s.Require().Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), retrievedSlots[1].Time)
	s.Require().Equal(int64(200), retrievedSlots[1].Height)
	s.Require().Equal(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), retrievedSlots[2].Time)
	s.Require().Equal(int64(300), retrievedSlots[2].Height)
}

func (s *KeeperTestSuite) TestAddValidatorQueuePendingSlot() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)

	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime, testHeight)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0].Time)
	s.Require().Equal(testHeight, slots[0].Height)

	// Adding the same slot again should not create a duplicate
	err = s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime, testHeight)
	s.Require().NoError(err)

	slots, err = s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1) // Still only one entry
}

func (s *KeeperTestSuite) TestRemoveValidatorQueuePendingSlot() {
	testTime1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight1 := int64(100)
	testTime2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	testHeight2 := int64(200)

	// Add two slots
	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime1, testHeight1)
	s.Require().NoError(err)
	err = s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime2, testHeight2)
	s.Require().NoError(err)

	// Remove one
	err = s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, testTime1, testHeight1)
	s.Require().NoError(err)

	// Verify only one remains
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime2, slots[0].Time)
	s.Require().Equal(testHeight2, slots[0].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_SortingEdgeCases_SameTimeDifferentHeights() {
	// Same time, different heights - should sort by height
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: testTime, Height: 300},
		{Time: testTime, Height: 100},
		{Time: testTime, Height: 200},
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 3)
	// Should be sorted by height when times are equal
	s.Require().Equal(testTime, retrievedSlots[0].Time)
	s.Require().Equal(int64(100), retrievedSlots[0].Height)
	s.Require().Equal(testTime, retrievedSlots[1].Time)
	s.Require().Equal(int64(200), retrievedSlots[1].Height)
	s.Require().Equal(testTime, retrievedSlots[2].Time)
	s.Require().Equal(int64(300), retrievedSlots[2].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_SortingEdgeCases_SameHeightDifferentTimes() {
	// Same height, different times - should sort by time first
	testHeight := int64(100)
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), Height: testHeight},
		{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Height: testHeight},
		{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Height: testHeight},
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 3)
	// Should be sorted by time first
	s.Require().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), retrievedSlots[0].Time)
	s.Require().Equal(testHeight, retrievedSlots[0].Height)
	s.Require().Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), retrievedSlots[1].Time)
	s.Require().Equal(testHeight, retrievedSlots[1].Height)
	s.Require().Equal(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), retrievedSlots[2].Time)
	s.Require().Equal(testHeight, retrievedSlots[2].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_Deduplication_SameTimeDifferentHeight() {
	// Same time but different height should NOT deduplicate
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: testTime, Height: 100},
		{Time: testTime, Height: 200}, // Different height, should NOT be deduplicated
		{Time: testTime, Height: 100}, // Same time+height, should be deduplicated
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 2) // Should have 2 unique entries (100 and 200)

	// Should be sorted
	s.Require().Equal(testTime, retrievedSlots[0].Time)
	s.Require().Equal(int64(100), retrievedSlots[0].Height)
	s.Require().Equal(testTime, retrievedSlots[1].Time)
	s.Require().Equal(int64(200), retrievedSlots[1].Height)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_Deduplication_SameHeightDifferentTime() {
	// Same height but different time should NOT deduplicate
	testHeight := int64(100)
	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Height: testHeight},
		{Time: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), Height: testHeight}, // Different time, should NOT be deduplicated
		{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Height: testHeight}, // Same time+height, should be deduplicated
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 2) // Should have 2 unique entries (different times)

	// Should be sorted by time first
	s.Require().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), retrievedSlots[0].Time)
	s.Require().Equal(testHeight, retrievedSlots[0].Height)
	s.Require().Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), retrievedSlots[1].Time)
	s.Require().Equal(testHeight, retrievedSlots[1].Height)
}

func (s *KeeperTestSuite) TestRemoveValidatorQueuePendingSlot_FromEmptyList() {
	// Remove from empty list should be no-op, not error
	err := s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 100)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

func (s *KeeperTestSuite) TestRemoveValidatorQueuePendingSlot_NonExistentEntry() {
	// Add one slot
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight := int64(100)
	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime, testHeight)
	s.Require().NoError(err)

	// Try to remove non-existent entry
	err = s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), 200)
	s.Require().NoError(err) // Should be no-op, not error

	// Original slot should still be there
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0].Time)
	s.Require().Equal(testHeight, slots[0].Height)
}

func (s *KeeperTestSuite) TestRemoveValidatorQueuePendingSlot_MiddleEntry() {
	// Add three slots
	testTime1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight1 := int64(100)
	testTime2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	testHeight2 := int64(200)
	testTime3 := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	testHeight3 := int64(300)

	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime1, testHeight1)
	s.Require().NoError(err)
	err = s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime2, testHeight2)
	s.Require().NoError(err)
	err = s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime3, testHeight3)
	s.Require().NoError(err)

	// Remove middle entry
	err = s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, testTime2, testHeight2)
	s.Require().NoError(err)

	// Verify only first and last remain
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 2)
	s.Require().Equal(testTime1, slots[0].Time)
	s.Require().Equal(testHeight1, slots[0].Height)
	s.Require().Equal(testTime3, slots[1].Time)
	s.Require().Equal(testHeight3, slots[1].Height)
}

func (s *KeeperTestSuite) TestRemoveValidatorQueuePendingSlot_AllEntries() {
	// Add two slots
	testTime1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testHeight1 := int64(100)
	testTime2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	testHeight2 := int64(200)

	err := s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime1, testHeight1)
	s.Require().NoError(err)
	err = s.stakingKeeper.AddValidatorQueuePendingSlot(s.ctx, testTime2, testHeight2)
	s.Require().NoError(err)

	// Remove all entries
	err = s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, testTime1, testHeight1)
	s.Require().NoError(err)
	err = s.stakingKeeper.RemoveValidatorQueuePendingSlot(s.ctx, testTime2, testHeight2)
	s.Require().NoError(err)

	// Key should be deleted (empty list)
	slots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

func (s *KeeperTestSuite) TestSetValidatorQueuePendingSlots_TimeWithNanosecondPrecision() {
	// Test with nanosecond precision
	testTime := time.Date(2024, 1, 1, 12, 34, 56, 123456789, time.UTC)
	testHeight := int64(100)

	slots := []stakingtypes.TimeHeightQueueSlot{
		{Time: testTime, Height: testHeight},
	}

	err := s.stakingKeeper.SetValidatorQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetValidatorQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0].Time) // Should preserve nanosecond precision
	s.Require().Equal(testHeight, retrievedSlots[0].Height)
}

// --- UBD Queue Tests ---

func (s *KeeperTestSuite) TestGetUBDQueuePendingSlots_NoEntries() {
	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Nil(slots)
}

func (s *KeeperTestSuite) TestGetUBDQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	err := s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])
}

func (s *KeeperTestSuite) TestGetUBDQueuePendingSlots_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	for _, t := range testTimes {
		err := s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, t)
		s.Require().NoError(err)
	}

	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)
	// Should be sorted
	for i, slot := range slots {
		s.Require().Equal(testTimes[i], slot)
	}
}

func (s *KeeperTestSuite) TestSetUBDQueuePendingSlots_EmptySlice() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add a slot first
	err := s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	// Set empty slice (should delete)
	err = s.stakingKeeper.SetUBDQueuePendingSlots(s.ctx, []time.Time{})
	s.Require().NoError(err)

	// Verify it's deleted
	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

func (s *KeeperTestSuite) TestSetUBDQueuePendingSlots_DuplicateEntries() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Set with duplicates
	slots := []time.Time{
		testTime,
		testTime, // duplicate
		testTime, // duplicate
	}

	err := s.stakingKeeper.SetUBDQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	// Verify only one entry persisted
	retrievedSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0])
}

func (s *KeeperTestSuite) TestSetUBDQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	slots := []time.Time{testTime}

	err := s.stakingKeeper.SetUBDQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0])
}

func (s *KeeperTestSuite) TestSetUBDQueuePendingSlots_MultipleEntries() {
	slots := []time.Time{
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	err := s.stakingKeeper.SetUBDQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 3)
	// Should be sorted
	s.Require().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), retrievedSlots[0])
	s.Require().Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), retrievedSlots[1])
	s.Require().Equal(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), retrievedSlots[2])
}

func (s *KeeperTestSuite) TestAddUBDQueuePendingSlot() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	err := s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])

	// Adding the same time again should not create a duplicate
	// Add calls Set internally, which deduplicates, so there should still be only 1 entry
	err = s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err = s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1) // Still only one entry due to deduplication in Set
	s.Require().Equal(testTime, slots[0])
}

func (s *KeeperTestSuite) TestAddUBDQueuePendingSlot_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	// Add multiple entries
	for _, t := range testTimes {
		err := s.stakingKeeper.AddUBDQueuePendingSlot(s.ctx, t)
		s.Require().NoError(err)
	}

	// Verify all entries were added and sorted
	slots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)
	// Should be sorted
	for i, slot := range slots {
		s.Require().Equal(testTimes[i], slot)
	}
}

func (s *KeeperTestSuite) TestSetUBDQueuePendingSlots_TimeWithNanosecondPrecision() {
	// Test with nanosecond precision
	testTime := time.Date(2024, 1, 1, 12, 34, 56, 123456789, time.UTC)

	slots := []time.Time{testTime}

	err := s.stakingKeeper.SetUBDQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetUBDQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0]) // Should preserve nanosecond precision
}

// --- Redelegation Queue Tests ---

func (s *KeeperTestSuite) TestGetRedelegationQueuePendingSlots_NoEntries() {
	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Nil(slots)
}

func (s *KeeperTestSuite) TestGetRedelegationQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	err := s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])
}

func (s *KeeperTestSuite) TestGetRedelegationQueuePendingSlots_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	for _, t := range testTimes {
		err := s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, t)
		s.Require().NoError(err)
	}

	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)
	// Should be sorted
	for i, slot := range slots {
		s.Require().Equal(testTimes[i], slot)
	}
}

func (s *KeeperTestSuite) TestSetRedelegationQueuePendingSlots_EmptySlice() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Add a slot first
	err := s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	// Set empty slice (should delete)
	err = s.stakingKeeper.SetRedelegationQueuePendingSlots(s.ctx, []time.Time{})
	s.Require().NoError(err)

	// Verify it's deleted
	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Empty(slots)
}

func (s *KeeperTestSuite) TestSetRedelegationQueuePendingSlots_DuplicateEntries() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Set with duplicates
	slots := []time.Time{
		testTime,
		testTime, // duplicate
		testTime, // duplicate
	}

	err := s.stakingKeeper.SetRedelegationQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	// Verify only one entry persisted
	retrievedSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0])
}

func (s *KeeperTestSuite) TestSetRedelegationQueuePendingSlots_SingleEntry() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	slots := []time.Time{testTime}

	err := s.stakingKeeper.SetRedelegationQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0])
}

func (s *KeeperTestSuite) TestSetRedelegationQueuePendingSlots_MultipleEntries() {
	slots := []time.Time{
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	err := s.stakingKeeper.SetRedelegationQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 3)
	// Should be sorted
	s.Require().Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), retrievedSlots[0])
	s.Require().Equal(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC), retrievedSlots[1])
	s.Require().Equal(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC), retrievedSlots[2])
}

func (s *KeeperTestSuite) TestAddRedelegationQueuePendingSlot() {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	err := s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1)
	s.Require().Equal(testTime, slots[0])

	// Adding the same time again should not create a duplicate
	// Add calls Set internally, which deduplicates, so there should still be only 1 entry
	err = s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, testTime)
	s.Require().NoError(err)

	slots, err = s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 1) // Still only one entry due to deduplication in Set
	s.Require().Equal(testTime, slots[0])
}

func (s *KeeperTestSuite) TestAddRedelegationQueuePendingSlot_MultipleEntries() {
	testTimes := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
	}

	// Add multiple entries
	for _, t := range testTimes {
		err := s.stakingKeeper.AddRedelegationQueuePendingSlot(s.ctx, t)
		s.Require().NoError(err)
	}

	// Verify all entries were added and sorted
	slots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(slots, 3)
	// Should be sorted
	for i, slot := range slots {
		s.Require().Equal(testTimes[i], slot)
	}
}

func (s *KeeperTestSuite) TestSetRedelegationQueuePendingSlots_TimeWithNanosecondPrecision() {
	// Test with nanosecond precision
	testTime := time.Date(2024, 1, 1, 12, 34, 56, 123456789, time.UTC)

	slots := []time.Time{testTime}

	err := s.stakingKeeper.SetRedelegationQueuePendingSlots(s.ctx, slots)
	s.Require().NoError(err)

	retrievedSlots, err := s.stakingKeeper.GetRedelegationQueuePendingSlots(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(retrievedSlots, 1)
	s.Require().Equal(testTime, retrievedSlots[0]) // Should preserve nanosecond precision
}
