package types

import (
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/gogo/protobuf/proto"

	amino "github.com/tendermint/go-amino"
)

type aminoCompat struct {
	bz     []byte
	jsonBz []byte
	err    error
}

var Debug = false

func aminoCompatError(errType string, x interface{}) error {
	if Debug {
		debug.PrintStack()
	}
	return fmt.Errorf(
		"amino %s Any marshaling error for %+v, this is likely because "+
			"amino is being used directly (instead of codec.Codec which is preferred) "+
			"or UnpackInterfacesMessage is not defined for some type which contains "+
			"a protobuf Any either directly or via one of its members. To see a "+
			"stacktrace of where the error is coming from, set the var Debug = true "+
			"in codec/types/amino_compat.go",
		errType, x,
	)
}

func (any Any) MarshalAmino() ([]byte, error) {
	ac := any.aminoCompat
	if ac == nil {
		return nil, aminoCompatError("binary unmarshal", any)
	}
	return ac.bz, ac.err
}

func (any *Any) UnmarshalAmino(bz []byte) error {
	any.aminoCompat = &aminoCompat{
		bz:  bz,
		err: nil,
	}
	return nil
}

func (any Any) MarshalJSON() ([]byte, error) {
	ac := any.aminoCompat
	if ac == nil {
		return nil, aminoCompatError("JSON marshal", any)
	}
	return ac.jsonBz, ac.err
}

func (any *Any) UnmarshalJSON(bz []byte) error {
	any.aminoCompat = &aminoCompat{
		jsonBz: bz,
		err:    nil,
	}
	return nil
}

// AminoUnpacker is an AnyUnpacker provided for backwards compatibility with
// amino for the binary un-marshaling phase
type AminoUnpacker struct {
	Cdc *amino.Codec
}

var _ AnyUnpacker = AminoUnpacker{}

func (a AminoUnpacker) UnpackAny(any *Any, iface interface{}) error {
	ac := any.aminoCompat
	if ac == nil {
		return aminoCompatError("binary unmarshal", reflect.TypeOf(iface))
	}
	err := a.Cdc.UnmarshalBinaryBare(ac.bz, iface)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(iface).Elem().Interface()
	err = UnpackInterfaces(val, a)
	if err != nil {
		return err
	}
	if m, ok := val.(proto.Message); ok {
		err := any.Pack(m)
		if err != nil {
			return err
		}
	} else {
		any.cachedValue = val
	}

	// this is necessary for tests that use reflect.DeepEqual and compare
	// proto vs amino marshaled values
	any.aminoCompat = nil

	return nil
}

// AminoUnpacker is an AnyUnpacker provided for backwards compatibility with
// amino for the binary marshaling phase
type AminoPacker struct {
	Cdc *amino.Codec
}

var _ AnyUnpacker = AminoPacker{}

func (a AminoPacker) UnpackAny(any *Any, _ interface{}) error {
	err := UnpackInterfaces(any.cachedValue, a)
	if err != nil {
		return err
	}
	bz, err := a.Cdc.MarshalBinaryBare(any.cachedValue)
	any.aminoCompat = &aminoCompat{
		bz:  bz,
		err: err,
	}
	return err
}

// AminoUnpacker is an AnyUnpacker provided for backwards compatibility with
// amino for the JSON marshaling phase
type AminoJSONUnpacker struct {
	Cdc *amino.Codec
}

var _ AnyUnpacker = AminoJSONUnpacker{}

func (a AminoJSONUnpacker) UnpackAny(any *Any, iface interface{}) error {
	ac := any.aminoCompat
	if ac == nil {
		return aminoCompatError("JSON unmarshal", reflect.TypeOf(iface))
	}
	err := a.Cdc.UnmarshalJSON(ac.jsonBz, iface)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(iface).Elem().Interface()
	err = UnpackInterfaces(val, a)
	if err != nil {
		return err
	}
	if m, ok := val.(proto.Message); ok {
		err := any.Pack(m)
		if err != nil {
			return err
		}
	} else {
		any.cachedValue = val
	}

	// this is necessary for tests that use reflect.DeepEqual and compare
	// proto vs amino marshaled values
	any.aminoCompat = nil

	return nil
}

// AminoUnpacker is an AnyUnpacker provided for backwards compatibility with
// amino for the JSON un-marshaling phase
type AminoJSONPacker struct {
	Cdc *amino.Codec
}

var _ AnyUnpacker = AminoJSONPacker{}

func (a AminoJSONPacker) UnpackAny(any *Any, _ interface{}) error {
	err := UnpackInterfaces(any.cachedValue, a)
	if err != nil {
		return err
	}
	bz, err := a.Cdc.MarshalJSON(any.cachedValue)
	any.aminoCompat = &aminoCompat{
		jsonBz: bz,
		err:    err,
	}
	return err
}
