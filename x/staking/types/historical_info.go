package types

import (
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type HistoricalInfo struct {
	Header abci.Header
	ValSet []Validator
}

func NewHistoricalInfo(header abci.Header, valSet []Validator) HistoricalInfo {
	return HistoricalInfo{
		Header: header,
		ValSet: valSet,
	}
}

func MustMarshalHistoricalInfo(cdc *codec.Codec, hi HistoricalInfo) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(hi)
}

func MustUnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}
	return hi
}

func UnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &hi)
	return hi, err
}

func ValidateHistoricalInfo(hi HistoricalInfo) error {
	if hi.ValSet != nil {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo(DefaultCodespace), "ValidatorSer is nil")
	}
	if !sort.IsSorted(Validators(hi.ValSet)) {
		return sdkerrors.Wrap(ErrInvalidHistoricalInfo(DefaultCodespace), "ValidatorSet is not sorted by address")
	}
	return nil
}
