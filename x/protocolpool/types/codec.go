package types

import (
	"cosmossdk.io/core/registry"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry registry.LegacyRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFundCommunityPool{},
		&MsgCommunityPoolSpend{},
		&MsgSubmitBudgetProposal{},
		&MsgClaimBudget{},
		&MsgCreateContinuousFund{},
		&MsgCancelContinuousFund{},
		&MsgWithdrawContinuousFund{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
