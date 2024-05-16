package types

import (
	corelegacy "cosmossdk.io/core/legacy"
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
		&MsgUpdateCometInfo{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}

// RegisterLegacyAminoCodec registers the necessary x/consensus interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc corelegacy.Amino) {
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/consensus/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateCometInfo{}, "cosmos-sdk/x/consensus/MsgCometInfo")
}
