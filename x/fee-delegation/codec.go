package fee_delegation

import "github.com/cosmos/cosmos-sdk/codec"

var moduleCodec = codec.New()

// RegisterCodec registers all the necessary types and interfaces for the module
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDelegateFeeAllowance{}, "delegation/MsgDelegateFeeAllowance", nil)
	cdc.RegisterConcrete(MsgRevokeFeeAllowance{}, "delegation/MsgRevokeFeeAllowance", nil)
	cdc.RegisterConcrete(BasicFeeAllowance{}, "delegation/BasicFeeAllowance", nil)
	cdc.RegisterInterface((*FeeAllowance)(nil), nil)
}
