package params

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EncodingConfig struct {
	InterfaceRegistry types.InterfaceRegistry
	Marshaler         codec.Marshaler
	TxDecoder         sdk.TxDecoder
	TxJSONDecoder     sdk.TxDecoder
	TxGenerator       client.TxGenerator
	Amino             *codec.Codec
}
