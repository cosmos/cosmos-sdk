package codec

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/tendermint/go-amino"
	cryptoamino "github.com/tendermint/tendermint/crypto/encoding/amino"
	tmtypes "github.com/tendermint/tendermint/types"
)

// amino codec to marshal/unmarshal
type Codec = amino.Codec

func New() *Codec {
	return amino.NewCodec()
}

// Register the go-crypto to the codec
func RegisterCrypto(cdc *Codec) {
	cryptoamino.RegisterAmino(cdc)
}

// RegisterEvidences registers Tendermint evidence types with the provided codec.
func RegisterEvidences(cdc *Codec) {
	tmtypes.RegisterEvidences(cdc)
}

// attempt to make some pretty json
func MarshalJSONIndent(cdc *Codec, obj interface{}) ([]byte, error) {
	bz, err := cdc.MarshalJSON(obj)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	err = json.Indent(&out, bz, "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// MustMarshalJSONIndent executes MarshalJSONIndent except it panics upon failure.
func MustMarshalJSONIndent(cdc *Codec, obj interface{}) []byte {
	bz, err := MarshalJSONIndent(cdc, obj)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %s", err))
	}

	return bz
}

//__________________________________________________________________

// generic sealed codec to be used throughout sdk
var Cdc *Codec

func init() {
	cdc := New()
	RegisterCrypto(cdc)
	RegisterEvidences(cdc)
	Cdc = cdc.Seal()
}
