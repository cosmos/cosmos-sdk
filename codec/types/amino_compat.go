package types

import (
	"fmt"
	"reflect"

	amino "github.com/tendermint/go-amino"
)

type aminoCompat struct {
	bz     []byte
	jsonBz []byte
	err    error
}

func (any Any) MarshalAmino() ([]byte, error) {
	ac := any.aminoCompat
	if ac == nil {
		return nil, fmt.Errorf("can't amino unmarshal")
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
		return nil, fmt.Errorf("can't JSON marshal")
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
		return fmt.Errorf("can't amino unmarshal %T", iface)
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
	any.cachedValue = val
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
		return fmt.Errorf("can't amino unmarshal %T", iface)
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
	any.cachedValue = val
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
