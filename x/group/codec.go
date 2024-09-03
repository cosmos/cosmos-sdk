package group

import (
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary group module concrete
// types and interfaces with the provided codec reference.
// These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	cdc.RegisterInterface((*DecisionPolicy)(nil), nil)
	cdc.RegisterConcrete(&ThresholdDecisionPolicy{}, "cosmos-sdk/ThresholdDecisionPolicy")
	cdc.RegisterConcrete(&PercentageDecisionPolicy{}, "cosmos-sdk/PercentageDecisionPolicy")

	legacy.RegisterAminoMsg(cdc, &MsgCreateGroup{}, "cosmos-sdk/MsgCreateGroup")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupMembers{}, "cosmos-sdk/MsgUpdateGroupMembers")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupAdmin{}, "cosmos-sdk/MsgUpdateGroupAdmin")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupMetadata{}, "cosmos-sdk/MsgUpdateGroupMetadata")
	legacy.RegisterAminoMsg(cdc, &MsgCreateGroupWithPolicy{}, "cosmos-sdk/MsgCreateGroupWithPolicy")
	legacy.RegisterAminoMsg(cdc, &MsgCreateGroupPolicy{}, "cosmos-sdk/MsgCreateGroupPolicy")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyAdmin{}, "cosmos-sdk/MsgUpdateGroupPolicyAdmin")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyDecisionPolicy{}, "cosmos-sdk/MsgUpdateGroupDecisionPolicy")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateGroupPolicyMetadata{}, "cosmos-sdk/MsgUpdateGroupPolicyMetadata")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/group/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawProposal{}, "cosmos-sdk/group/MsgWithdrawProposal")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/group/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgExec{}, "cosmos-sdk/group/MsgExec")
	legacy.RegisterAminoMsg(cdc, &MsgLeaveGroup{}, "cosmos-sdk/group/MsgLeaveGroup")
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*coretransaction.Msg)(nil),
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
		&MsgLeaveGroup{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)

	registrar.RegisterInterface(
		"cosmos.group.v1.DecisionPolicy",
		(*DecisionPolicy)(nil),
		&ThresholdDecisionPolicy{},
		&PercentageDecisionPolicy{},
	)
}
