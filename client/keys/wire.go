package keys

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc)
}

// marshal keys
func MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.MarshalJSON(o)
}

// unmarshal json
func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return cdc.UnmarshalJSON(bz, ptr)
}
