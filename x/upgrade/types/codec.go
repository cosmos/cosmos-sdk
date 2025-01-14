package types

import (
<<<<<<< HEAD
	"github.com/cosmos/cosmos-sdk/codec"
=======
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"
	"cosmossdk.io/x/gov/types/v1beta1"

>>>>>>> 5581225a9 (fix(x/upgrade): register missing implementation for SoftwareUpgradeProposal (#23179))
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Plan{}, "cosmos-sdk/Plan", nil)
	cdc.RegisterConcrete(&SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)
	cdc.RegisterConcrete(&CancelSoftwareUpgradeProposal{}, "cosmos-sdk/CancelSoftwareUpgradeProposal", nil)
	legacy.RegisterAminoMsg(cdc, &MsgSoftwareUpgrade{}, "cosmos-sdk/MsgSoftwareUpgrade")
	legacy.RegisterAminoMsg(cdc, &MsgCancelUpgrade{}, "cosmos-sdk/MsgCancelUpgrade")
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&SoftwareUpgradeProposal{},
		&CancelSoftwareUpgradeProposal{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSoftwareUpgrade{},
		&MsgCancelUpgrade{},
	)
<<<<<<< HEAD

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
=======
	registrar.RegisterImplementations(
		(*v1beta1.Content)(nil),
		&SoftwareUpgradeProposal{},
	)
	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
>>>>>>> 5581225a9 (fix(x/upgrade): register missing implementation for SoftwareUpgradeProposal (#23179))
}
