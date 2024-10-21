package types

import (
	"cosmossdk.io/core/registry"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/feemarket interfaces (messages) on the
// provided LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	legacy.RegisterAminoMsg(cdc, &MsgParams{}, "feemarket/MsgParams")
}

// RegisterInterfaces registers the x/feemarket interfaces (messages + msg server) on the
// provided InterfaceRegistry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
