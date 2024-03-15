package types

import (
	"cosmossdk.io/core/registry"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFundCommunityPool{},
		&MsgCommunityPoolSpend{},
		&MsgSubmitBudgetProposal{},
		&MsgClaimBudget{},
		&MsgCreateContinuousFund{},
		&MsgCancelContinuousFund{},
		&MsgWithdrawContinuousFund{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
