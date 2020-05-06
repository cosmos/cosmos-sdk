package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// Codec defines the interface needed to serialize x/auth state. It must be
// aware of all concrete account types.
type Codec interface {
	codec.Marshaler

	MarshalAccount(acc exported.Account) ([]byte, error)
	UnmarshalAccount(bz []byte) (exported.Account, error)

	MarshalAccountJSON(acc exported.Account) ([]byte, error)
	UnmarshalAccountJSON(bz []byte) (exported.Account, error)
}

// RegisterCodec registers the account interfaces and concrete types on the
// provided Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ModuleAccountI)(nil), nil)
	cdc.RegisterInterface((*exported.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*exported.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(&ModuleAccount{}, "cosmos-sdk/ModuleAccount", nil)
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}

// RegisterKeyTypeCodec registers an external concrete type defined in
// another module for the internal ModuleCdc.
func RegisterKeyTypeCodec(o interface{}, name string) {
	amino.RegisterConcrete(o, name, nil)
}

var (
	amino = codec.New()
)

func init() {
	RegisterCodec(amino)
	codec.RegisterCrypto(amino)
}
