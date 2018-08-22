package keys

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

var cdc *wire.Codec

func init() {
	cdc = wire.NewCodec()

	cdc.RegisterInterface((*sdk.Address)(nil), nil)
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
