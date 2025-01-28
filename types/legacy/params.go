// Package legacy contains types and interfaces that have support removed, but may need to be exported for legacy
// and migration purposes.
package legacy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// The following types are from the removed x/params modules and can be used to satisfy
// legacy interfaces that may use these.

type (
	ValueValidatorFn func(value interface{}) error

	// ParamSetPair is used for associating paramsubspace key and field of param
	// structs.
	ParamSetPair struct {
		Key         []byte
		Value       interface{}
		ValidatorFn ValueValidatorFn
	}
)

// NewParamSetPair creates a new ParamSetPair instance.
func NewParamSetPair(key []byte, value interface{}, vfn ValueValidatorFn) ParamSetPair {
	return ParamSetPair{key, value, vfn}
}

// ParamSetPairs Slice of KeyFieldPair
type ParamSetPairs []ParamSetPair

// ParamSet defines an interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
}

type (
	// Subspace defines an interface that implements the legacy x/params Subspace
	// type.
	//
	// NOTE: This is used solely for migration of x/params managed parameters.
	Subspace interface {
		GetParamSet(ctx sdk.Context, ps ParamSet)
	}
)
