package v1beta1

import (
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Content)(nil), nil)
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/MsgDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/MsgVoteWeighted")
	cdc.RegisterConcrete(&TextProposal{}, "cosmos-sdk/TextProposal", nil)
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
	)
	registrar.RegisterInterface(
		"cosmos.gov.v1beta1.Content",
		(*Content)(nil),
		&TextProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
