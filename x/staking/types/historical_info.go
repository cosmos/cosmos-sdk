package types

import (
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// HistoricalInfo contains the historical information that gets stored at each height
type HistoricalInfo struct {
	Header abci.Header `json:"header" yaml:"header"`
	ValSet []Validator `json:"valset" yaml:"valset"`
}

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header abci.Header, valSet []Validator) HistoricalInfo {
	sort.Sort(Validators(valSet))
	return HistoricalInfo{
		Header: header,
		ValSet: valSet,
	}
}

// MustMarshalHistoricalInfo wll marshal historical info and panic on error
func MustMarshalHistoricalInfo(cdc *codec.Codec, hi HistoricalInfo) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(hi)
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}
	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &hi)
	return hi, err
}

// ValidateBasic will ensure HistoricalInfo is not nil and sorted
func ValidateBasic(hi HistoricalInfo) error {
	if len(hi.ValSet) == 0 {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is empty")
	}
	if !sort.IsSorted(Validators(hi.ValSet)) {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo, "validator set is not sorted by address")
	}
	return nil
}
