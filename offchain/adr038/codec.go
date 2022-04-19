package adr038

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ModuleCdc is the codec used by the module to serialize and deserialize data
var ModuleCdc = codec.NewAminoCodec(amino)
var amino = codec.NewLegacyAmino()

// RegisterInterfaces adds offchain sdk.Msg types to the interface registry
func RegisterInterfaces(ir types.InterfaceRegistry) {
	ir.RegisterImplementations((*sdk.Msg)(nil), &MsgSignData{})
}

// RegisterLegacyAminoCodec registers amino's legacy codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSignData{}, "offchain/MsgSignData", nil)
}

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
}
