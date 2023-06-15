package types

import "cosmossdk.io/errors"

var (
	// ErrGovInvalidSigner this has been duplicated from the x/gov module to prevent a cyclic dependency
	// see: https://github.com/cosmos/cosmos-sdk/blob/91d14c04accdd5ded86888514401f1cdd0949eb2/x/gov/types/errors.go#L20
	ErrGovInvalidSigner = errors.Register(ModuleName, 1, "expected gov account as only signer for proposal message")
)
