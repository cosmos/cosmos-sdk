package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/nft interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgIssue{}, "cosmos-sdk/MsgIssue", nil)
	cdc.RegisterConcrete(&MsgMint{}, "cosmos-sdk/MsgMint", nil)
	cdc.RegisterConcrete(&MsgEdit{}, "cosmos-sdk/MsgEdit", nil)
	cdc.RegisterConcrete(&MsgSend{}, "cosmos-sdk/MsgSend", nil)
	cdc.RegisterConcrete(&MsgBurn{}, "cosmos-sdk/MsgBurn", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIssue{},
		&MsgMint{},
		&MsgEdit{},
		&MsgSend{},
		&MsgBurn{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
