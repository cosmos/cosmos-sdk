package bank

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/tendermint/tmlibs/common"
)

//----------------------------------------
// Error constructors

// error when provided inputs are invalid
func ErrInvalidInput(msg string) sdk.Error {
	return sdk.ErrGeneric(msg, []cmn.KVPair{
		{[]byte("module"), []byte("bank")},
		{[]byte("cause"), []byte("INVALID_INPUT")},
	})
}

// error when no inputs are provided
func ErrNoInputs() sdk.Error {
	return sdk.ErrGeneric("", []cmn.KVPair{
		{[]byte("module"), []byte("bank")},
		{[]byte("cause"), []byte("NO_INPUTS")},
	})
}

// error when provided outputs are invalid
func ErrInvalidOutput(msg string) sdk.Error {
	return sdk.ErrGeneric(msg, []cmn.KVPair{
		{[]byte("module"), []byte("bank")},
		{[]byte("cause"), []byte("INVALID_OUTPUT")},
	})
}

// error when no outputs are provided
func ErrNoOutputs() sdk.Error {
	return sdk.ErrGeneric("", []cmn.KVPair{
		{[]byte("module"), []byte("bank")},
		{[]byte("cause"), []byte("NO_OUTPUTS")},
	})
}
