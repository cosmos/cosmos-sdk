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
	ValSet Validators  `json:"valset" yaml:"valset"`
}

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header abci.Header, valSet Validators) HistoricalInfo {
	sort.Sort(valSet)
	return HistoricalInfo{
		Header: header,
		ValSet: valSet,
	}
}

// ToProto converts a HistoricalInfo into a HistoricalInfoProto type.
func (hi HistoricalInfo) ToProto() HistoricalInfoProto {
	valsetProto := make([]ValidatorProto, len(hi.ValSet))
	for i, val := range hi.ValSet {
		valsetProto[i] = val.ToProto()
	}

	return HistoricalInfoProto{
		Header: hi.Header,
		Valset: valsetProto,
	}
}

// ToHistoricalInfo converts a HistoricalInfoProto to a HistoricalInfo type.
func (hip HistoricalInfoProto) ToHistoricalInfo() HistoricalInfo {
	valset := make(Validators, len(hip.Valset))
	for i, valProto := range hip.Valset {
		valset[i] = valProto.ToValidator()
	}

	return NewHistoricalInfo(hip.Header, valset)
}

// MustMarshalHistoricalInfo wll marshal historical info and panic on error
func MustMarshalHistoricalInfo(cdc codec.Marshaler, hi HistoricalInfo) []byte {
	hiProto := hi.ToProto()
	return cdc.MustMarshalBinaryLengthPrefixed(&hiProto)
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
	hip := HistoricalInfoProto{}
	if err := cdc.UnmarshalBinaryLengthPrefixed(value, &hip); err != nil {
		return HistoricalInfo{}, err
	}

	return hip.ToHistoricalInfo(), nil
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
