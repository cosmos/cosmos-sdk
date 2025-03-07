package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
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

// RegisterLegacyAminoCodec registers the necessary x/protocolpool interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgFundCommunityPool{}, "cosmos-sdk/x/protocolpool/MsgFundCommunityPool")
	legacy.RegisterAminoMsg(cdc, &MsgCommunityPoolSpend{}, "cosmos-sdk/x/protocolpoolMsgCommunityPoolSpend")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitBudgetProposal{}, "cosmos-sdk/x/protocolpool/MsgSubmitBudgetProposal")
	legacy.RegisterAminoMsg(cdc, &MsgClaimBudget{}, "cosmos-sdk/x/protocolpool/MsgClaimBudget")
	legacy.RegisterAminoMsg(cdc, &MsgCreateContinuousFund{}, "cosmos-sdk/x/protocolpool/MsgCreateContinuousFund")
	legacy.RegisterAminoMsg(cdc, &MsgCancelContinuousFund{}, "cosmos-sdk/x/protocolpool/MsgCancelContinuousFund")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawContinuousFund{}, "cosmos-sdk/x/protocolpool/MsgWithdrawContinuousFund")

	cdc.RegisterConcrete(&Budget{}, "cosmos-sdk/x/protocolpool/Budget", nil)
	cdc.RegisterConcrete(&ContinuousFund{}, "cosmos-sdk/x/protocolpool/ContinuousFund", nil)
	cdc.RegisterConcrete(&DistributionAmount{}, "cosmos-sdk/x/protocolpool/DistributionAmount", nil)
	cdc.RegisterConcrete(&Params{}, "cosmos-sdk/x/protocolpool/Params", nil)
}
