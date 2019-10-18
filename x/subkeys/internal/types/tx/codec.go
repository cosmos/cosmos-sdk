package tx

import "github.com/cosmos/cosmos-sdk/codec"

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(DelegatedTx{}, "cosmos-sdk/DelegatedTx", nil)
}
