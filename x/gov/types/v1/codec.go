package v1

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
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/v1/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitMultipleChoiceProposal{}, "gov/MsgSubmitMultipleChoiceProposal")
	legacy.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/v1/MsgDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/v1/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/v1/MsgVoteWeighted")
	legacy.RegisterAminoMsg(cdc, &MsgExecLegacyContent{}, "cosmos-sdk/v1/MsgExecLegacyContent")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/gov/v1/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateMessageParams{}, "x/gov/v1/MsgUpdateMessageParams")
	legacy.RegisterAminoMsg(cdc, &MsgSudoExec{}, "cosmos-sdk/x/gov/v1/MsgSudoExec")
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgSubmitMultipleChoiceProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
		&MsgExecLegacyContent{},
		&MsgUpdateParams{},
		&MsgUpdateMessageParams{},
		&MsgSudoExec{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
