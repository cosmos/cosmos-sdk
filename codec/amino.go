package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	amino "github.com/tendermint/go-amino"
	cryptoamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Cdc defines a global generic sealed Amino codec to be used throughout sdk. It
// has all Tendermint crypto and evidence types registered.
//
// TODO: Consider removing this global.
var Cdc *Codec

func init() {
	cdc := New()
	RegisterCrypto(cdc)
	RegisterEvidences(cdc)
	Cdc = cdc.Seal()
}

// Codec defines a type alias for an Amino codec.
type Codec = amino.Codec

func New() *Codec {
	return amino.NewCodec()
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *Codec) {
	cryptoamino.RegisterAmino(cdc)
}

// RegisterEvidences registers Tendermint evidence types with the provided Amino
// codec.
func RegisterEvidences(cdc *Codec) {
	tmtypes.RegisterEvidences(cdc)
}

// MarshalJSONIndent provides a utility for indented JSON encoding of an object
// via an Amino codec. It returns an error if it cannot serialize or indent as
// JSON.
func MarshalJSONIndent(m JSONMarshaler, obj interface{}) ([]byte, error) {
	bz, err := m.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	if err = json.Indent(&out, bz, "", "  "); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

// MustMarshalJSONIndent executes MarshalJSONIndent except it panics upon failure.
func MustMarshalJSONIndent(m JSONMarshaler, obj interface{}) []byte {
	bz, err := MarshalJSONIndent(m, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}
