package keys

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	codec.RegisterCrypto(cdc)
}

// marshal keys
func MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.MarshalJSON(o)
}

// unmarshal json
func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return cdc.UnmarshalJSON(bz, ptr)
}
