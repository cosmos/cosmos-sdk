package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
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
func (h Height) Compare(other exported.Height) int64 {
	height, ok := other.(Height)
	if !ok {
		panic(fmt.Sprintf("cannot compare against invalid height type: %T. expected height type: %T", other, h))
	}
	var cmp int64
	if h.EpochNumber != height.EpochNumber {
		cmp = int64(h.EpochNumber) - int64(height.EpochNumber)
	} else {
		cmp = int64(h.EpochHeight) - int64(height.EpochHeight)
	}
	if cmp < 0 {
		return -1
	} else if cmp > 0 {
		return 1
	}
	return 0
}

// LT Helper comparison function returns true if h < other
func (h Height) LT(other exported.Height) bool {
	return h.Compare(other) == -1
}

// LTE Helper comparison function returns true if h <= other
func (h Height) LTE(other exported.Height) bool {
	cmp := h.Compare(other)
	return cmp <= 0
}

// GT Helper comparison function returns true if h > other
func (h Height) GT(other exported.Height) bool {
	return h.Compare(other) == 1
}

// GTE Helper comparison function returns true if h >= other
func (h Height) GTE(other exported.Height) bool {
	cmp := h.Compare(other)
	return cmp >= 0
}

// EQ Helper comparison function returns true if h == other
func (h Height) EQ(other exported.Height) bool {
	return h.Compare(other) == 0
}

// String returns a string representation of Height
func (h Height) String() string {
	return fmt.Sprintf("epoch-%d-height-%d", h.EpochNumber, h.EpochHeight)
}

// Decrement will return a decremented height from the given height. If this is not possible,
// an error is returned
// Decrement will return a new height with the EpochHeight decremented
// If the EpochHeight is already at lowest value (1), then false success flag is returend
func (h Height) Decrement() (decremented Height, success bool) {
	if h.EpochHeight == 0 {
		return Height{}, false
	}
	return NewHeight(h.EpochNumber, h.EpochHeight-1), true
}

// Increment will return an incremented height from the given height.
// Increment will return a height with the same epoch number but an
// incremented epoch height
func (h Height) Increment() Height {
	return NewHeight(h.EpochNumber, h.EpochHeight+1)
}

// IsZero returns true if height epoch and epoch-height are both 0
func (h Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
