package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/gogo/protobuf/proto"
)

type (
	ValueValidatorFn func(value proto.Message) error

	// ParamSetPair is used for associating paramSubspace key and field of param
	// structs.
	ParamSetPair struct {
		Key         []byte
		Value       codec.ProtoMarshaler
		ValidatorFn ValueValidatorFn
	}
)

// NewParamSetPair creates a new ParamSetPair instance.
func NewParamSetPair(key []byte, value codec.ProtoMarshaler, vfn ValueValidatorFn) ParamSetPair {
	return ParamSetPair{key, value, vfn}
}

// ParamSetPairs Slice of KeyFieldPair
type ParamSetPairs []ParamSetPair

// ParamSet defines an interface for structs containing parameters for a module
type ParamSet interface {
	ParamSetPairs() ParamSetPairs
}
