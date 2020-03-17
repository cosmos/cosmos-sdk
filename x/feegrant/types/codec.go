package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/feegrant/exported"
)

// Codec defines the interface required to serialize custom x/gov types.
type Codec interface {
	codec.Marshaler

	MarshalFeeAllowanceGrant(exported.FeeAllowanceGrant) ([]byte, error)
	UnmarshalFeeAllowanceGrant([]byte) (exported.FeeAllowanceGrant, error)
}

// RegisterCodec registers the account types and interface
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.FeeAllowance)(nil), nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "cosmos-sdk/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "cosmos-sdk/PeriodicFeeAllowance", nil)

	cdc.RegisterConcrete(FeeGrantTx{}, "cosmos-sdk/FeeGrantTx", nil)
}

var (
	amino = codec.New()

	ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
