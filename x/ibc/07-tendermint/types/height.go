package types

import (
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

var _ clientexported.Height = Height{}

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

func (h Height) String() string {
	return fmt.Sprintf("epoch:%d-height:%d", h.EpochNumber, h.EpochHeight)
}

func (h Height) IsZero() bool {
	return h.EpochNumber == 0 && h.EpochHeight == 0
}
