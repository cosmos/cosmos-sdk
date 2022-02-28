package group

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
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
	cdc.RegisterConcrete(&PercentageDecisionPolicy{}, "cosmos-sdk/PercentageDecisionPolicy", nil)
	cdc.RegisterConcrete(&MsgCreateGroup{}, "cosmos-sdk/MsgCreateGroup", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupMembers{}, "cosmos-sdk/MsgUpdateGroupMembers", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupAdmin{}, "cosmos-sdk/MsgUpdateGroupAdmin", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupMetadata{}, "cosmos-sdk/MsgUpdateGroupMetadata", nil)
	cdc.RegisterConcrete(&MsgCreateGroupWithPolicy{}, "cosmos-sdk/MsgCreateGroupWithPolicy", nil)
	cdc.RegisterConcrete(&MsgCreateGroupPolicy{}, "cosmos-sdk/MsgCreateGroupPolicy", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupPolicyAdmin{}, "cosmos-sdk/MsgUpdateGroupPolicyAdmin", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupPolicyDecisionPolicy{}, "cosmos-sdk/MsgUpdateGroupPolicyDecisionPolicy", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupPolicyMetadata{}, "cosmos-sdk/MsgUpdateGroupPolicyMetadata", nil)
	cdc.RegisterConcrete(&MsgSubmitProposal{}, "cosmos-sdk/group/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(&MsgWithdrawProposal{}, "cosmos-sdk/group/MsgWithdrawProposal", nil)
	cdc.RegisterConcrete(&MsgVote{}, "cosmos-sdk/group/MsgVote", nil)
	cdc.RegisterConcrete(&MsgExec{}, "cosmos-sdk/group/MsgExec", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGroup{},
		&MsgUpdateGroupMembers{},
		&MsgUpdateGroupAdmin{},
		&MsgUpdateGroupMetadata{},
		&MsgCreateGroupWithPolicy{},
		&MsgCreateGroupPolicy{},
		&MsgUpdateGroupPolicyAdmin{},
		&MsgUpdateGroupPolicyDecisionPolicy{},
		&MsgUpdateGroupPolicyMetadata{},
		&MsgSubmitProposal{},
		&MsgWithdrawProposal{},
		&MsgVote{},
		&MsgExec{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)

	registry.RegisterInterface(
		"cosmos.group.v1beta1.DecisionPolicy",
		(*DecisionPolicy)(nil),
		&ThresholdDecisionPolicy{},
		&PercentageDecisionPolicy{},
	)
}

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}
