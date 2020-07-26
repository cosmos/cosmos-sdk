package keys

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cryptocodec "github.com/KiraCore/cosmos-sdk/crypto/codec"
)

// KeysCdc defines codec to be used with key operations
var KeysCdc *codec.Codec

func init() {
	KeysCdc = codec.New()
	cryptocodec.RegisterCrypto(KeysCdc)
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
