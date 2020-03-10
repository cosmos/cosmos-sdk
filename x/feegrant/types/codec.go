package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.FeeAllowance)(nil), nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "cosmos-sdk/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "cosmos-sdk/PeriodicFeeAllowance", nil)

	cdc.RegisterConcrete(FeeGrantTx{}, "cosmos-sdk/FeeGrantTx", nil)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/staking module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/staking and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
