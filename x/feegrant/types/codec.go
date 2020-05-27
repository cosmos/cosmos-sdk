package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*FeeAllowanceI)(nil), nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "cosmos-sdk/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "cosmos-sdk/PeriodicFeeAllowance", nil)
	cdc.RegisterConcrete(FeeGrantTx{}, "cosmos-sdk/FeeGrantTx", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.gov.v1.FeeAllowance",
		(*FeeAllowanceI)(nil),
		&BasicFeeAllowance{},
		&PeriodicFeeAllowance{},
	)
}

var (
	amino = codec.New()

	ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
