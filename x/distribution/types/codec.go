package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec concrete distribution types on amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(MsgWithdrawValidatorCommission{}, "cosmos-sdk/MsgWithdrawValidatorCommission", nil)
	cdc.RegisterConcrete(MsgSetWithdrawAddress{}, "cosmos-sdk/MsgModifyWithdrawAddress", nil)
	cdc.RegisterConcrete(CommunityPoolSpendProposal{}, "cosmos-sdk/CommunityPoolSpendProposal", nil)
	cdc.RegisterConcrete(MsgFundCommunityPool{}, "cosmos-sdk/MsgFundCommunityPool", nil)
}

// ModuleCdc is a generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
