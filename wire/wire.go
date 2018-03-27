package wire

import (
	"bytes"
	"reflect"

	"github.com/tendermint/go-wire"
)

type Codec struct{}

func NewCodec() *Codec {
	return &Codec{}
}

func (cdc *Codec) MarshalBinary(o interface{}) ([]byte, error) {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(o, w, n, err)
	return w.Bytes(), *err
}

func (cdc *Codec) UnmarshalBinary(bz []byte, o interface{}) error {
	r, n, err := bytes.NewBuffer(bz), new(int), new(error)

	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		wire.ReadBinaryPtr(o, r, len(bz), n, err)
	} else {
		wire.ReadBinary(o, r, len(bz), n, err)
	}
	return *err
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	w, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteJSON(o, w, n, err)
	return w.Bytes(), *err
}

func (cdc *Codec) UnmarshalJSON(bz []byte, o interface{}) (err error) {

	rv := reflect.ValueOf(o)
	if rv.Kind() == reflect.Ptr {
		wire.ReadJSONPtr(o, bz, &err)
	} else {
		wire.ReadJSON(o, bz, &err)
	}
	return err
}

//----------------------------------------------

func RegisterCrypto(cdc *Codec) {
	// TODO
}
