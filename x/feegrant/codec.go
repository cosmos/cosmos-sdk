package feegrant

import (
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/feegrant interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	legacy.RegisterAminoMsg(cdc, &MsgGrantAllowance{}, "cosmos-sdk/MsgGrantAllowance")
	legacy.RegisterAminoMsg(cdc, &MsgRevokeAllowance{}, "cosmos-sdk/MsgRevokeAllowance")

	cdc.RegisterInterface((*FeeAllowanceI)(nil), nil)
	cdc.RegisterConcrete(&BasicAllowance{}, "cosmos-sdk/BasicAllowance")
	cdc.RegisterConcrete(&PeriodicAllowance{}, "cosmos-sdk/PeriodicAllowance")
	cdc.RegisterConcrete(&AllowedMsgAllowance{}, "cosmos-sdk/AllowedMsgAllowance")
}

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*coretransaction.Msg)(nil),
		&MsgGrantAllowance{},
		&MsgRevokeAllowance{},
	)

	registrar.RegisterInterface(
		"cosmos.feegrant.v1beta1.FeeAllowanceI",
		(*FeeAllowanceI)(nil),
		&BasicAllowance{},
		&PeriodicAllowance{},
		&AllowedMsgAllowance{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
