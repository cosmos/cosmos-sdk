package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(MsgWithdrawValidatorCommission{}, "cosmos-sdk/MsgWithdrawValidatorCommission", nil)
	cdc.RegisterConcrete(MsgSetWithdrawAddress{}, "cosmos-sdk/MsgModifyWithdrawAddress", nil)
	cdc.RegisterConcrete(CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal", nil)
}

// generic sealed codec to be used throughout module
var ModuleCdc = codec.New()

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
