package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// Codec defines the interface needed to serialize x/supply state. It must
// be aware of all concrete supply types.
type Codec interface {
	codec.Marshaler

	MarshalSupply(supply exported.SupplyI) ([]byte, error)
	UnmarshalSupply(bz []byte) (exported.SupplyI, error)

	MarshalSupplyJSON(supply exported.SupplyI) ([]byte, error)
	UnmarshalSupplyJSON(bz []byte) (exported.SupplyI, error)
}

// RegisterCodec registers the necessary x/supply interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*exported.SupplyI)(nil), nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(&Supply{}, "cosmos-sdk/Supply", nil)
}

var (
	amino = codec.New()

	// ModuleCdc references the global x/supply module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/supply and
	// defined at the application level.
	ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
	amino.Seal()
}
