package keys

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// KeysCdc defines codec to be used with key operations
var KeysCdc *codec.Codec

func init() {
	KeysCdc = codec.New()
	codec.RegisterCrypto(KeysCdc)
	KeysCdc.Seal()
}

// marshal keys
func MarshalJSON(o interface{}) ([]byte, error) {
	return KeysCdc.MarshalJSON(o)
}

// unmarshal json
func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return KeysCdc.UnmarshalJSON(bz, ptr)
}
