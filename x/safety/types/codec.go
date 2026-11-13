package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the x/safety msgs with the interface registry.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgTrip{},
		&MsgReset{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the x/safety msgs on the LegacyAmino
// codec for Amino JSON signing.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/safety/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgTrip{}, "cosmos-sdk/x/safety/MsgTrip")
	legacy.RegisterAminoMsg(cdc, &MsgReset{}, "cosmos-sdk/x/safety/MsgReset")
	cdc.RegisterConcrete(&Params{}, "cosmos-sdk/x/safety/Params", nil)
}
