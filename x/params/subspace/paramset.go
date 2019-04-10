package subspace

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Used for associating paramsubspace key and field of param structs
type ParamSetPair struct {
	Key   []byte
	Value interface{}
}

// Slice of KeyFieldPair
type ParamSetPairs []ParamSetPair

// Interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
	ValidateKV(key string, value string) (interface{}, sdk.Error)
}
