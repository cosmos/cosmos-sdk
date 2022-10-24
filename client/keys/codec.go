package keys

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	cryptocodec "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
)

// TODO: remove this file https://github.com/pointnetwork/cosmos-point-sdk/issues/8047

// KeysCdc defines codec to be used with key operations
var KeysCdc *codec.LegacyAmino

func init() {
	KeysCdc = codec.NewLegacyAmino()
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
