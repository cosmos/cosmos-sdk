package types

import (
	"sort"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header cmtproto.Header, valSet Validators, powerReduction math.Int) HistoricalInfo {
	// Must sort in the same way that CometBFT does
	sort.SliceStable(valSet.Validators, func(i, j int) bool {
		return ValidatorsByVotingPower(valSet.Validators).Less(i, j, powerReduction)
	})

	return HistoricalInfo{
		Header: header,
		Valset: valSet.Validators,
	}
}

// ValidateBasic will ensure HistoricalInfo is not nil and sorted
func ValidateBasic(hi HistoricalInfo, valAc address.Codec) error {
	if len(hi.Valset) == 0 {
		return errors.Wrap(ErrInvalidHistoricalInfo, "validator set is empty")
	}

	if !sort.IsSorted(Validators{Validators: hi.Valset, ValidatorCodec: valAc}) {
		return errors.Wrap(ErrInvalidHistoricalInfo, "validator set is not sorted by address")
	}

	return nil
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
