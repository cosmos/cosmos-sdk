package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// RegisterLegacyAminoCodec registers the account types and interface
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// cdc.RegisterInterface((*isFeeAllowance_Sum)(nil), nil)
	// cdc.RegisterConcrete(&FeeAllowance_BasicFeeAllowance{}, "cosmos-sdk/ProtoBasicFeeAllowance", nil)
	cdc.RegisterInterface((*FeeAllowanceI)(nil), nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "cosmos-sdk/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "cosmos-sdk/PeriodicFeeAllowance", nil)
	cdc.RegisterConcrete(FeeGrantTx{}, "cosmos-sdk/FeeGrantTx", nil)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/feegrant module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/feegrant and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
