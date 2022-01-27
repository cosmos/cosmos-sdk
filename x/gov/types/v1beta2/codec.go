package v1beta2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitProposal{}, "cosmos-sdk/v1beta2/MsgSubmitProposal", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "cosmos-sdk/v1beta2/MsgDeposit", nil)
	cdc.RegisterConcrete(&MsgVote{}, "cosmos-sdk/v1beta2/MsgVote", nil)
	cdc.RegisterConcrete(&MsgVoteWeighted{}, "cosmos-sdk/v1beta2/MsgVoteWeighted", nil)
	cdc.RegisterConcrete(&MsgExecLegacyContent{}, "cosmos-sdk/v1beta2/MsgExecLegacyContent", nil)
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
	RegisterLegacyAminoCodec(types.ModuleCdc.LegacyAmino)
}
