package types

import (
	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/distribution interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward")
	legacy.RegisterAminoMsg(cdc, &MsgWithdrawValidatorCommission{}, "cosmos-sdk/MsgWithdrawValCommission")
	legacy.RegisterAminoMsg(cdc, &MsgSetWithdrawAddress{}, "cosmos-sdk/MsgModifyWithdrawAddress")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/distribution/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgDepositValidatorRewardsPool{}, "cosmos-sdk/distr/MsgDepositValRewards")

	cdc.RegisterConcrete(Params{}, "cosmos-sdk/x/distribution/Params", nil)
}

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgWithdrawDelegatorReward{},
		&MsgWithdrawValidatorCommission{},
		&MsgSetWithdrawAddress{},
		&MsgUpdateParams{},
		&MsgDepositValidatorRewardsPool{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
