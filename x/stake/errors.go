// nolint
package stake

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tmlibs/common"
)

//----------------------------------------
// Error constructors

func ErrNoDescription() sdk.Error {
	return sdk.ErrGeneric("description must be included", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_INPUT")},
	})
}

func ErrNotEnoughBondShares(shares string) sdk.Error {
	return sdk.ErrGeneric(fmt.Sprintf("not enough shares only have %v", shares), []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_BOND")},
	})
}

func ErrCandidateEmpty() sdk.Error {
	return sdk.ErrGeneric("Cannot bond to an empty candidate", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrBadBondingDenom() sdk.Error {
	return sdk.ErrGeneric("Invalid coin denomination", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_BOND")},
	})
}

func ErrBadBondingAmount() sdk.Error {
	return sdk.ErrGeneric("Amount must be > 0", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_BOND")},
	})
}

func ErrNoBondingAcct() sdk.Error {
	return sdk.ErrGeneric("No bond account for this (address, validator) pair", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrCommissionNegative() sdk.Error {
	return sdk.ErrGeneric("Commission must be positive", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrCommissionHuge() sdk.Error {
	return sdk.ErrGeneric("Commission cannot be more than 100%", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrBadValidatorAddr() sdk.Error {
	return sdk.ErrGeneric("Validator does not exist for that address", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrBadCandidateAddr() sdk.Error {
	return sdk.ErrGeneric("Candidate does not exist for that address", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrBadDelegatorAddr() sdk.Error {
	return sdk.ErrGeneric("Delegator does not exist for that address", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrCandidateExistsAddr() sdk.Error {
	return sdk.ErrGeneric("Candidate already exists, cannot re-declare candidacy", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrMissingSignature() sdk.Error {
	return sdk.ErrGeneric("Missing signature", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrBondNotNominated() sdk.Error {
	return sdk.ErrGeneric("Cannot bond to non-nominated account", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrNoCandidateForAddress() sdk.Error {
	return sdk.ErrGeneric("Validator does not exist for that address", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrNoDelegatorForAddress() sdk.Error {
	return sdk.ErrGeneric("Delegator does not contain validator bond", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}

func ErrInsufficientFunds() sdk.Error {
	return sdk.ErrGeneric("Insufficient bond shares", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_INPUT")},
	})
}

func ErrBadShares() sdk.Error {
	return sdk.ErrGeneric("bad shares provided as input, must be MAX or decimal", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_INPUT")},
	})
}

func ErrBadRemoveValidator() sdk.Error {
	return sdk.ErrGeneric("Error removing validator", []cmn.KVPair{
		cmn.KVPair{[]byte("module"), []byte("stake")},
		cmn.KVPair{[]byte("cause"), []byte("INVALID_VALIDATOR")},
	})
}
