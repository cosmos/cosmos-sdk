package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(ir codectypes.InterfaceRegistry) {
	ir.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgFundCommunityPool{},
		&MsgCommunityPoolSpend{},
		&MsgSubmitBudgetProposal{},
		&MsgClaimBudget{},
		&MsgCreateContinuousFund{},
		&MsgCancelContinuousFund{},
		&MsgWithdrawContinuousFund{},
	)

	msgservice.RegisterMsgServiceDesc(ir, &_Msg_serviceDesc)
}
