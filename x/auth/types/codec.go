package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AuthCodec defines the interface needed to serialize x/auth state. It must be
// aware of all concrete account types.
type AuthCodec interface {
	codec.Marshaler

	MarshalAccount(acc exported.AccountI) ([]byte, error)
	UnmarshalAccount(bz []byte) (exported.AccountI, error)

	MarshalAccountJSON(acc exported.AccountI) ([]byte, error)
	UnmarshalAccountJSON(bz []byte) (exported.AccountI, error)
}

// RegisterCodec registers the account interfaces and concrete types on the
// provided Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*exported.AccountI)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}
