package v1beta1

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
	)
	registry.RegisterInterface(
		"cosmos.gov.v1beta1.Content",
		(*Content)(nil),
		&TextProposal{},
	)
	registry.RegisterImplementations(
		(*Content)(nil),
		&distrtypes.CommunityPoolSpendProposal{}, //nolint: staticcheck // avoid using `CommunityPoolSpendProposal`, might be reomved in future.
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
