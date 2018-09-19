package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawDelegationRewardsAll{}, "cosmos-sdk/MsgWithdrawDelegationRewardsAll", nil)
	cdc.RegisterConcrete(MsgWithdrawDelegationReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(MsgWithdrawValidatorRewardsAll{}, "cosmos-sdk/MsgWithdrawValidatorRewardsAll", nil)
	cdc.RegisterConcrete(MsgModifyWithdrawAddress{}, "cosmos-sdk/MsgModifyWithdrawAddress", nil)
}

// generic sealed codec to be used throughout sdk
var MsgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	MsgCdc = cdc.Seal()
}
