package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

var _ clientexported.Height = Height{}

// Height is the tendermint implementation of clientexported.Height
// Each Epoch represents a continous range of heights, when the chain height gets
// reset on update, the epoch number is incremented
// EpochHeight must start at 1
type Height struct {
	EpochNumber uint64
	EpochHeight uint64
}

func NewHeight(epochNumber, epochHeight uint64) Height {
	return Height{
		EpochNumber: epochNumber,
		EpochHeight: epochHeight,
	}
}

// Compare implements clientexported.Height
// It first compares based on epoch numbers, whichever has the higher epoch number is the higher height
// If epoch number is the same, then the epoch height is compared
// If the comparing height is not tendermint height, error is returned
func (h Height) Compare(other clientexported.Height) (int64, error) {
	tmHeight, ok := other.(Height)
	if !ok {
		return 0, sdkerrors.Wrapf(ErrInvalidHeightComparison, "cannot compare tendermint.Height to non tendermint height: %v", other)
	}
	if h.EpochNumber != tmHeight.EpochNumber {
		return int64(h.EpochNumber) - int64(tmHeight.EpochNumber), nil
	}
	return int64(h.EpochHeight) - int64(tmHeight.EpochHeight), nil
}

// Decrement implements clientexported.Height
// Decrement will return a new height with the EpochHeight decremented
// If the EpochHeight is already at lowest value (1), then error is returend
func (h Height) Decrement() (clientexported.Height, error) {
	if h.EpochHeight <= 1 {
		return Height{}, sdkerrors.Wrapf(clienttypes.ErrInvalidHeight,
			"cannot decrement EpochHeight below 1, reached lowest height for this epoch. height: %s", h.String())
	}
	return NewHeight(h.EpochNumber, h.EpochHeight-1), nil
}

// Valid implements clientexported.Height
// Returns false if EpochHeight is 0
func (h Height) Valid() bool {
	return h.EpochHeight != 0
}

func (h Height) String() string {
	return fmt.Sprintf("epoch:%d-height:%d", h.EpochNumber, h.EpochHeight)
}

func (h Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
