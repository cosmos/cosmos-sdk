package group

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary group module concrete
// types and interfaces with the provided codec reference.
// These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*DecisionPolicy)(nil), nil)
	cdc.RegisterConcrete(&ThresholdDecisionPolicy{}, "cosmos-sdk/ThresholdDecisionPolicy", nil)
	cdc.RegisterConcrete(&MsgCreateGroup{}, "cosmos-sdk/MsgCreateGroup", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupMembers{}, "cosmos-sdk/MsgUpdateGroupMembers", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupAdmin{}, "cosmos-sdk/MsgUpdateGroupAdmin", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupMetadata{}, "cosmos-sdk/MsgUpdateGroupMetadata", nil)
	cdc.RegisterConcrete(&MsgCreateGroupAccount{}, "cosmos-sdk/MsgCreateGroupAccount", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupAccountAdmin{}, "cosmos-sdk/MsgUpdateGroupAccountAdmin", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupAccountDecisionPolicy{}, "cosmos-sdk/MsgUpdateGroupAccountDecisionPolicy", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupAccountMetadata{}, "cosmos-sdk/MsgUpdateGroupAccountMetadata", nil)
	cdc.RegisterConcrete(&MsgCreateProposal{}, "cosmos-sdk/group/MsgCreateProposal", nil)
	cdc.RegisterConcrete(&MsgVote{}, "cosmos-sdk/group/MsgVote", nil)
	cdc.RegisterConcrete(&MsgExec{}, "cosmos-sdk/group/MsgExec", nil)
}

func RegisterTypes(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGroup{},
		&MsgUpdateGroupMembers{},
		&MsgUpdateGroupAdmin{},
		&MsgUpdateGroupMetadata{},
		&MsgCreateGroupAccount{},
		&MsgUpdateGroupAccountAdmin{},
		&MsgUpdateGroupAccountDecisionPolicy{},
		&MsgUpdateGroupAccountMetadata{},
		&MsgCreateProposal{},
		&MsgVote{},
		&MsgExec{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	registry.RegisterInterface(
		"regen.group.v1alpha1.DecisionPolicy",
		(*DecisionPolicy)(nil),
		&ThresholdDecisionPolicy{},
	)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
}
