package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/fee_grant/exported"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.FeeAllowance)(nil), nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "feegrant/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "feegrant/PeriodicFeeAllowance", nil)
}

// ModuleCdc generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
