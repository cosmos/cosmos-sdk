package v1

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/v1/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/v1/MsgDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/v1/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/v1/MsgVoteWeighted")
	legacy.RegisterAminoMsg(cdc, &MsgExecLegacyContent{}, "cosmos-sdk/v1/MsgExecLegacyContent")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/gov/v1/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgProposeConstitutionAmendment{}, "cosmos-sdk/v1/MsgProposeAmendment")
	legacy.RegisterAminoMsg(cdc, &MsgProposeLaw{}, "cosmos-sdk/v1/MsgProposeLaw")
	legacy.RegisterAminoMsg(cdc, &MsgCreateGovernor{}, "cosmos-sdk/v1/MsgCreateGovernor")
	legacy.RegisterAminoMsg(cdc, &MsgEditGovernor{}, "cosmos-sdk/v1/MsgEditGovernor")
	legacy.RegisterAminoMsg(cdc, &MsgDelegateGovernor{}, "cosmos-sdk/v1/MsgDelegateGovernor")
	legacy.RegisterAminoMsg(cdc, &MsgUndelegateGovernor{}, "cosmos-sdk/v1/MsgUndelegateGovernor")
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
		&MsgExecLegacyContent{},
		&MsgUpdateParams{},
		&MsgProposeConstitutionAmendment{},
		&MsgProposeLaw{},
		&MsgCreateGovernor{},
		&MsgEditGovernor{},
		&MsgDelegateGovernor{},
		&MsgUndelegateGovernor{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
