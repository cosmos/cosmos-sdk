package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ exported.Height = (*Height)(nil)

// NewHeight is a constructor for the IBC height type
func NewHeight(epochNumber, epochHeight uint64) *Height {
	return &Height{
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
func (h *Height) Compare(other exported.Height) int64 {
	height, ok := other.(*Height)
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
func (h *Height) LT(other exported.Height) bool {
	return h.Compare(other) == -1
}

// GT Helper comparison function returns true if h > other
func (h *Height) GT(other exported.Height) bool {
	return h.Compare(other) == 1
}

// EQ Helper comparison function returns true if h == other
func (h *Height) EQ(other exported.Height) bool {
	return h.Compare(other) == 0
}

// Decrement will return a decremented height from the given height. If this is not possible,
// an error is returned
// Decrement will return a new height with the EpochHeight decremented
// If the EpochHeight is already at lowest value (1), then false success flag is returend
func (h *Height) Decrement() (decremented exported.Height, success bool) {
	if h.EpochHeight <= 1 {
		return &Height{}, false
	}
	return NewHeight(h.EpochNumber, h.EpochHeight-1), true
}

// Increment will return an incremented height from the given height.
// Increment will return a height with the same epoch number but an
// incremented epoch height
func (h *Height) Increment() exported.Height {
	return NewHeight(h.EpochNumber, h.EpochHeight+1)
}

// IsValid returns true if height is valid, false otherwise
// Returns false if EpochHeight is 0
func (h *Height) IsValid() bool {
	if h == nil {
		return false
	}
	return h.EpochHeight != 0
}

// String implements the fmt.Stringer interface
// func (h Height) String() string {
// 	return fmt.Sprintf("epoch-%d_height-%d", h.EpochNumber, h.EpochHeight)
// }

// IsZero returns true if height epoch and epoch-height are both 0
func (h *Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
