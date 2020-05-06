package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers the necessary x/crisis interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgVerifyInvariant{}, "cosmos-sdk/MsgVerifyInvariant", nil)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/crisis module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/crisis and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
