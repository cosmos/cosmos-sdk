package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/02-client/exported"
)

var _ exported.Height = (*Height)(nil)

// NewHeight is a constructor for the IBC height type
func NewHeight(epochNumber, epochHeight uint64) Height {
	return Height{
		EpochNumber: epochNumber,
		EpochHeight: epochHeight,
	}
}

/// Compare implements a method to compare two heights. When comparing two heights a, b
// we can call a.Compare(b) which will return
// -1 if a < b
// 0  if a = b
// 1  if a > b
//
// It first compares based on epoch numbers, whichever has the higher epoch number is the higher height
// If epoch number is the same, then the epoch height is compared
func (h Height) Compare(other Height) int64 {
	var cmp int64
	if h.EpochNumber != other.EpochNumber {
		cmp = int64(h.EpochNumber) - int64(other.EpochNumber)
	} else {
		cmp = int64(h.EpochHeight) - int64(other.EpochHeight)
	}
	if cmp < 0 {
		return -1
	} else if cmp > 0 {
		return 1
	}
	return 0
}

// LT Helper comparison function returns true if h < other
func (h Height) LT(other Height) bool {
	return h.Compare(other) == -1
}

// GT Helper comparison function returns true if h > other
func (h Height) GT(other Height) bool {
	return h.Compare(other) == 1
}

// EQ Helper comparison function returns true if h == other
func (h Height) EQ(other Height) bool {
	return h.Compare(other) == 0
}

// Decrement will return a decremented height from the given height. If this is not possible,
// an error is returned
// Decrement will return a new height with the EpochHeight decremented
// If the EpochHeight is already at lowest value (1), then false success flag is returend
func (h Height) Decrement() (decremented Height, success bool) {
	if h.EpochHeight <= 1 {
		return Height{}, false
	}
	return NewHeight(h.EpochNumber, h.EpochHeight-1), true
}

// Valid returns true if height is valid, false otherwise
// Returns false if EpochHeight is 0
func (h Height) Valid() bool {
	return h.EpochHeight != 0
}

// String implements the fmt.Stringer interface
func (h Height) String() string {
	return fmt.Sprintf("epoch-%d_height-%d", h.EpochNumber, h.EpochHeight)
}

// IsZero returns true if height epoch and epoch-height are both 0
func (h Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
