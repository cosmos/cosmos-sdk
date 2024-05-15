package types

import (
	corelegacy "cosmossdk.io/core/legacy"
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc corelegacy.Amino) {
	cdc.RegisterConcrete(Plan{}, "cosmos-sdk/Plan")
	cdc.RegisterConcrete(&SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal")
	cdc.RegisterConcrete(&CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal")
	legacy.RegisterAminoMsg(cdc, &MsgSoftwareUpgrade{}, "cosmos-sdk/MsgSoftwareUpgrade")
	legacy.RegisterAminoMsg(cdc, &MsgCancelUpgrade{}, "cosmos-sdk/MsgCancelUpgrade")
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*coretransaction.Msg)(nil),
		&MsgSoftwareUpgrade{},
		&MsgCancelUpgrade{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
