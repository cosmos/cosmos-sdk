package types

import (
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header abci.Header, valSet Validators) HistoricalInfo {
	sort.Sort(valSet)
	return HistoricalInfo{
		Header: header,
		Valset: valSet,
	}
}

// MustMarshalHistoricalInfo wll marshal historical info and panic on error
func MustMarshalHistoricalInfo(cdc codec.Marshaler, hi HistoricalInfo) []byte {
	return cdc.MustMarshalBinaryBare(&hi)
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc codec.Marshaler, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}
	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc codec.Marshaler, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.UnmarshalBinaryBare(value, &hi)
	return hi, err
}

// ValidateBasic will ensure HistoricalInfo is not nil and sorted
func ValidateBasic(hi HistoricalInfo) error {
	if len(hi.Valset) == 0 {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is empty")
	}
	if !sort.IsSorted(Validators(hi.Valset)) {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is not sorted by address")
	}

	return nil
}
