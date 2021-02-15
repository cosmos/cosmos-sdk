package types

import (
	"sort"

	"github.com/gogo/protobuf/proto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header tmproto.Header, valSet Validators) HistoricalInfo {
	// Must sort in the same way that tendermint does
	sort.Sort(ValidatorsByVotingPower(valSet))

	return HistoricalInfo{
		Header: header,
		Valset: valSet,
	}
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc codec.BinaryMarshaler, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc codec.BinaryMarshaler, value []byte) (hi HistoricalInfo, err error) {
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

// Equal checks if receiver is equal to the parameter
func (hi *HistoricalInfo) Equal(hi2 *HistoricalInfo) bool {
	if !proto.Equal(&hi.Header, &hi2.Header) {
		return false
	}
	if len(hi.Valset) != len(hi2.Valset) {
		return false
	}
	for i := range hi.Valset {
		if !hi.Valset[i].Equal(&hi2.Valset[i]) {
			return false
		}
	}
	return true
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (hi HistoricalInfo) UnpackInterfaces(c codectypes.AnyUnpacker) error {
	for i := range hi.Valset {
		if err := hi.Valset[i].UnpackInterfaces(c); err != nil {
			return err
		}
	}
	return nil
}
