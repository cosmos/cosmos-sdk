package types

import (
	"fmt"

	amino "github.com/tendermint/go-amino"
)

type aminoCompat struct {
	aminoBz []byte
	err     error
}

func (any Any) MarshalAmino() ([]byte, error) {
	ac := any.aminoCompat
	if ac == nil {
		return nil, fmt.Errorf("can't amino unmarshal")
	}
	return ac.aminoBz, ac.err
}

func (any *Any) UnmarshalAmino(bz []byte) error {
	any.aminoCompat = &aminoCompat{
		aminoBz: bz,
		err:     nil,
	}
	return nil
}

type AminoUnpacker struct {
	Cdc *amino.Codec
}

var _ AnyUnpacker = AminoUnpacker{}

func (a AminoUnpacker) UnpackAny(any *Any, iface interface{}) error {
	ac := any.aminoCompat
	if ac == nil {
		return fmt.Errorf("can't amino unmarshal %T", iface)
	}
	err := a.Cdc.UnmarshalBinaryBare(ac.aminoBz, iface)
	if err != nil {
		return err
	}
	any.cachedValue = iface
	return nil
}

type AminoPacker struct {
	Cdc *amino.Codec
}

func (a AminoPacker) UnpackAny(any *Any, _ interface{}) error {
	bz, err := a.Cdc.MarshalBinaryBare(any.cachedValue)
	any.aminoCompat = &aminoCompat{
		aminoBz: bz,
		err:     err,
	}
	return err
}

var _ AnyUnpacker = AminoPacker{}
