package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AuthCodec defines the interface needed to serialize x/auth state. It must be
// aware of all concrete account types.
type AuthCodec interface {
	codec.Marshaler

	MarshalAccount(acc exported.Account) ([]byte, error)
	UnmarshalAccount(bz []byte) (exported.Account, error)

	MarshalAccountJSON(acc exported.Account) ([]byte, error)
	UnmarshalAccountJSON(bz []byte) (exported.Account, error)
}

// RegisterCodec registers the account interfaces and concrete types on the
// provided Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.GenesisAccount)(nil), nil)
	cdc.RegisterInterface((*exported.Account)(nil), nil)
	cdc.RegisterConcrete(&BaseAccount{}, "cosmos-sdk/Account", nil)
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx", nil)
}
