package adr036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}

// RegisterInterfaces adds offchain sdk.Msg types to the interface registry
func RegisterInterfaces(ir types.InterfaceRegistry) {
	ir.RegisterImplementations((*sdk.Msg)(nil), &MsgSignArbitraryData{})
}

// RegisterLegacyAminoCodec registers amino's legacy codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(MsgSignArbitraryData{}, "offchain/adr036/MsgSignArbitraryData", nil)
}
