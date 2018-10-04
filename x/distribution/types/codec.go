package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawDelegatorRewardsAll{}, "cosmos-sdk/MsgWithdrawDelegationRewardsAll", nil)
	cdc.RegisterConcrete(MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(MsgWithdrawValidatorRewardsAll{}, "cosmos-sdk/MsgWithdrawValidatorRewardsAll", nil)
	cdc.RegisterConcrete(MsgSetWithdrawAddress{}, "cosmos-sdk/MsgModifyWithdrawAddress", nil)
}

// generic sealed codec to be used throughout module
var MsgCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	MsgCdc = cdc.Seal()
}
