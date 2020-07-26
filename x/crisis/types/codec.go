package types

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	codectypes "github.com/KiraCore/cosmos-sdk/codec/types"
	cryptocodec "github.com/KiraCore/cosmos-sdk/crypto/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
)

// RegisterCodec registers the necessary x/crisis interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(&MsgVerifyInvariant{}, "cosmos-sdk/MsgVerifyInvariant", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgVerifyInvariant{},
	)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/crisis module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/crisis and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino, codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
