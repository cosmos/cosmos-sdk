package v1beta2

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
	codec.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/v1beta2/MsgSubmitProposal")
	codec.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/v1beta2/MsgDeposit")
	codec.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/v1beta2/MsgVote")
	codec.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/v1beta2/MsgVoteWeighted")
	codec.RegisterAminoMsg(cdc, &MsgExecLegacyContent{}, "cosmos-sdk/v1beta2/MsgExecLegacyContent")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
		&MsgExecLegacyContent{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}
