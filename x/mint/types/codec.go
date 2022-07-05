package types

import (
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

var amino = codec.NewLegacyAmino()

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(registrar registry.AminoRegistrar) {
	registrar.RegisterConcrete(Params{}, "cosmos-sdk/x/mint/Params")
	legacy.RegisterAminoMsg(registrar, &MsgUpdateParams{}, "cosmos-sdk/x/mint/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*coretransaction.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
